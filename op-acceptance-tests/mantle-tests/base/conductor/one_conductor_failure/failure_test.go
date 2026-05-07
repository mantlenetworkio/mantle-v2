package one_conductor_failure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// TestOneConductorFailureAndRecovery is the entry point for the paired
// single-conductor-crash failure + recovery scenario. Both halves share
// a single orchestrator boot (and therefore a single conductor cluster
// state machine). "Failure" crashes the LEADER's conductor — exercising
// op-conductor README failure scenario #5 ("one Conductor failed /
// stopped working") and the README's "100% uptime" guarantee — and
// "Recovery" restarts that same dead conductor and asserts the cluster
// returns to its pre-failure healthy 3-member baseline.
//
// We deliberately crash the LEADER (not a follower): a leader-crash is
// the strictly stronger scenario because recovery must reconcile both
// (a) the dead conductor's raft re-attachment AND (b) the orphan
// op-node that was the active sequencer at crash time. A follower
// crash would only exercise (a). Leader-crash subsumes follower-crash
// for recovery-correctness purposes.
//
// Subtests run in declaration order; "Recovery" requires the post-
// "Failure" state and must not be invoked in isolation. To run the
// recovery half:
//
//	go test -run TestOneConductorFailureAndRecovery
//
// which executes both subtests in order.
func TestOneConductorFailureAndRecovery(gt *testing.T) {
	parentT := devtest.SerialT(gt)
	sys := presets.NewMantleMinimalWithFaultyConductors(parentT)

	// Suite-wide pre-failure baseline.
	for chainID, conductors := range sys.ConductorSets {
		conductorhelpers.RequireHealthyConductorCluster(parentT, sys.L2Chain, chainID, conductors)
	}

	// State carried across subtests: which conductor was crashed in
	// "Failure", so "Recovery" knows which one to restart. Keyed by
	// chain ID so this scales naturally to multi-chain presets.
	deadLeaders := map[stack.L2NetworkID]stack.ConductorID{}

	if len(sys.SysgoConductors) < 3 {
		parentT.Skipf("scenario needs in-process conductors and >=3 voters; "+
			"got %d sysgo conductors (kurtosis/persistent backends do "+
			"not expose Conductor.Stop)", len(sys.SysgoConductors))
	}

	gt.Run("Failure", func(sgt *testing.T) {
		runFailure(devtest.SerialT(sgt), sys, deadLeaders)
	})
	if gt.Failed() {
		return
	}
	gt.Run("Recovery", func(sgt *testing.T) {
		runRecovery(devtest.SerialT(sgt), sys, deadLeaders)
	})
}

// runFailure exercises op-conductor README failure scenario #5
// (op-conductor/README.md ~line 95):
//
//	"one Conductor failed / stopped working"
//
// and the README guarantee:
//
//	"100% uptime with no more than 1 node failure (for a standard 3
//	 node setup)"
//
// We crash the LEADER's in-process conductor — its op-node and op-geth
// keep running as an orphan sequencer. The two surviving conductors
// must elect a new leader from voters within ~1s heartbeat timeout
// (RaftHeartbeatTimeout in op-devstack/sysgo/l2_conductor.go), the new
// leader's action loop must drive startSequencer, the new leader's EL
// must produce fresh blocks, and — crucially — the orphan EL must stay
// on the new leader's canonical chain (no fork).
//
// Orphan-state semantics
//
// With its conductor dead, the old leader's op-node is in the dangerous
// "still claims to sequence, but blocks go nowhere" state the README's
// commit-then-publish gate is designed to produce. Concretely:
//
//   - SequencerActive on the orphan op-node MUST still be true. Nothing
//     auto-stops a sequencer when its conductor dies — only a live
//     conductor's action loop calls StopSequencer on a non-leader.
//   - The orphan EL's head MUST be on the new leader's canonical chain
//     (no fork). The orphan op-node's onBuildSealed
//     (op-node/rollup/sequencing/sequencer.go) commits-then-publishes-
//     then-inserts; when CommitUnsafePayload fails (conductor RPC gone),
//     onBuildSealed early-returns and BOTH asyncGossip.Gossip AND
//     engine.PayloadProcessEvent are skipped — so the orphan never
//     inserts its OWN un-committed payloads. However, the orphan
//     op-node is still subscribed to the unsafe-payload libp2p topic
//     (op-node/p2p/event.go OnUnsafeL2Payload), so it RECEIVES the new
//     leader's gossiped blocks and feeds them into its local op-geth
//     via engine_newPayload + FCU just like any follower would. Net
//     result: the orphan EL stays on the new leader's chain; we verify
//     safety by hash-matching the orphan's head against the new leader
//     EL at the same height.
//
// State left for the recovery subtest:
//   - the dead leader's conductor: stopped (its op-conductor service is
//     gone)
//   - leadership: rotated to a healthy follower
//   - dead leader's op-node + op-geth: still running as an orphan, EL
//     on the new leader's canonical chain
//   - deadLeaders[chainID]: the dead conductor's ID, so runRecovery can
//     restart it
func runFailure(t devtest.T, sys *presets.MantleMinimalWithFaultyConductors, deadLeaders map[stack.L2NetworkID]stack.ConductorID) {
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestOneConductorFailureAndRecovery/Failure",
	)
	r := t.Require()

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; failover scenario "+
				"needs >= 3", chainID, len(conductors))
			continue
		}

		// 1. Capture the current leader before crashing it.
		leaderInfoPre := conductors[0].FetchLeader()
		r.NotNil(leaderInfoPre,
			"chain %s: cluster has no current leader before crash", chainID)
		oldLeaderID := leaderInfoPre.ID
		r.NotEmpty(oldLeaderID,
			"chain %s: empty leader ID before crash", chainID)
		logger.Info("Pre-crash leader",
			"chain", chainID, "leaderID", oldLeaderID)

		// 2. Resolve the in-process conductor wrapper for the leader,
		//    and pick at least one survivor's dsl.Conductor.
		leaderSysgo, ok := sys.SysgoConductors[stack.ConductorID(oldLeaderID)]
		r.True(ok,
			"chain %s: in-process conductor for leader %s missing "+
				"from sys.SysgoConductors", chainID, oldLeaderID)

		var survivor *dsl.Conductor
		for _, c := range conductors {
			if c.Escape().ID() != stack.ConductorID(oldLeaderID) {
				survivor = c
				break
			}
		}
		r.NotNil(survivor,
			"chain %s: no survivor conductor (every dsl.Conductor "+
				"resolved to the leader being crashed)", chainID)

		// 2b. Capture handles to the soon-to-be-orphan op-node and op-geth.
		orphanCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, oldLeaderID)
		r.NotNil(orphanCL,
			"chain %s: could not locate L2CL paired with old leader %s",
			chainID, oldLeaderID)
		orphanEL := conductorhelpers.ELPairedWithConductor(sys.L2Chain, oldLeaderID)
		r.NotNil(orphanEL,
			"chain %s: could not locate L2EL paired with old leader %s",
			chainID, oldLeaderID)
		orphanBaseline := orphanEL.BlockRefByLabel(eth.Unsafe).Number
		logger.Info("Captured orphan baseline pre-crash",
			"chain", chainID,
			"orphanEL", orphanEL.Escape().ID(),
			"baselineHeight", orphanBaseline)

		// 3. Crash the leader's conductor. Conductor.Stop shuts down
		//    the underlying *conductor.OpConductor service, which in
		//    turn stops raft and the HTTP RPC server.
		logger.Info("Crashing leader conductor",
			"chain", chainID, "leaderID", oldLeaderID)
		leaderSysgo.Stop()

		// Record the dead conductor for recovery to pick up. Done
		// immediately after Stop so that even if subsequent
		// assertions fail, the parent test's gt.Failed() short-circuit
		// will skip Recovery (no risk of running Recovery against a
		// poisoned cluster). We deliberately do NOT register a
		// t.Cleanup restart — Recovery owns the restart-and-rejoin
		// proof.
		deadLeaders[chainID] = stack.ConductorID(oldLeaderID)

		// 4. Poll the survivor for a new leader.
		var newLeaderID string
		electionDeadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(electionDeadline) {
			li := survivor.FetchLeader()
			if li != nil && li.ID != "" && li.ID != oldLeaderID {
				newLeaderID = li.ID
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.NotEmpty(newLeaderID,
			"chain %s: no new leader elected within 15s after old "+
				"leader %s crashed (raft re-election did not complete)",
			chainID, oldLeaderID)
		r.NotEqual(oldLeaderID, newLeaderID,
			"chain %s: raft kept the dead conductor %s as leader",
			chainID, oldLeaderID)
		logger.Info("New leader elected by survivors",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"newLeaderID", newLeaderID)

		// 5. The new leader's conductor must converge on
		//    (IsLeader && !Paused && Healthy).
		var newLeaderDsl *dsl.Conductor
		for _, c := range conductors {
			if c.Escape().ID() == stack.ConductorID(newLeaderID) {
				newLeaderDsl = c
				break
			}
		}
		r.NotNil(newLeaderDsl,
			"chain %s: dsl.Conductor for new leader %s missing",
			chainID, newLeaderID)

		convergeDeadline := time.Now().Add(15 * time.Second)
		var converged bool
		for time.Now().Before(convergeDeadline) {
			if newLeaderDsl.IsLeader() &&
				!newLeaderDsl.FetchPaused() &&
				newLeaderDsl.FetchSequencerHealthy() {
				converged = true
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		r.True(converged,
			"chain %s: new leader conductor %s never converged on "+
				"(IsLeader && !Paused && Healthy) within 15s",
			chainID, newLeaderID)

		// 5b. Wait for new leader's op-node to flip SequencerActive=true.
		newLeaderCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, newLeaderID)
		r.NotNil(newLeaderCL,
			"chain %s: could not locate L2CL paired with new leader %s",
			chainID, newLeaderID)
		const startDeadlineDur = 15 * time.Second
		startDeadline := time.Now().Add(startDeadlineDur)
		var newActive bool
		var lastActiveErr error
		for time.Now().Before(startDeadline) {
			probeCtx, cancelProbe := context.WithTimeout(
				t.Ctx(), 2*time.Second)
			newActive, lastActiveErr = newLeaderCL.RollupAPI().
				SequencerActive(probeCtx)
			cancelProbe()
			if lastActiveErr == nil && newActive {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.NoError(lastActiveErr,
			"chain %s: SequencerActive RPC failed on new leader CL %s",
			chainID, newLeaderCL.ID())
		r.True(newActive,
			"chain %s, new leader CL %s: SequencerActive=false %s after "+
				"election — conductor took over raft leadership but its "+
				"action loop never drove startSequencer on op-node",
			chainID, newLeaderCL.ID(), startDeadlineDur)

		// 6. Direct uptime guarantee: chain producing blocks under new leader.
		newLeaderEL := conductorhelpers.ELPairedWithConductor(sys.L2Chain, newLeaderID)
		r.NotNil(newLeaderEL,
			"chain %s: could not locate L2EL paired with new leader %s",
			chainID, newLeaderID)
		conductorhelpers.AssertChainAdvances(t, newLeaderEL,
			fmt.Sprintf("chain %s: post-conductor-crash recovery", chainID))

		// 7. Orphan-state assertions.
		const orphanObservationWindow = 6 * time.Second
		time.Sleep(orphanObservationWindow)

		probeCtx, cancelProbe := context.WithTimeout(t.Ctx(), 2*time.Second)
		orphanActive, orphanActiveErr := orphanCL.RollupAPI().
			SequencerActive(probeCtx)
		cancelProbe()
		r.NoError(orphanActiveErr,
			"chain %s: SequencerActive RPC failed on orphan CL %s",
			chainID, orphanCL.ID())
		r.True(orphanActive,
			"chain %s: orphan CL %s reports SequencerActive=false after "+
				"its conductor crashed — sysgo has no auto-stop on conductor "+
				"death, so the orphan op-node MUST still claim to be "+
				"sequencing. If this fires, op-node grew a new auto-shutdown "+
				"path and the test's threat model is wrong.",
			chainID, orphanCL.ID())

		// No-fork check: orphan EL on new leader's canonical chain.
		orphanHeadRef := orphanEL.BlockRefByLabel(eth.Unsafe)
		orphanHead := orphanHeadRef.Number
		newLeaderAtOrphanHeight := newLeaderEL.BlockRefByNumber(orphanHead)
		r.Equal(newLeaderAtOrphanHeight.Hash, orphanHeadRef.Hash,
			"chain %s: orphan EL %s head at height %d (hash %s) does not "+
				"match new leader EL %s at the same height (hash %s) — "+
				"orphan is on a divergent fork. This means a non-canonical "+
				"payload reached engine_newPayload on the orphan EL "+
				"(orphan-built payload that bypassed the commit gate, or "+
				"corrupt gossip).",
			chainID, orphanEL.Escape().ID(), orphanHead,
			orphanHeadRef.Hash, newLeaderEL.Escape().ID(),
			newLeaderAtOrphanHeight.Hash)
		logger.Info("Orphan EL on new leader's canonical chain (no fork)",
			"chain", chainID,
			"orphanEL", orphanEL.Escape().ID(),
			"baseline", orphanBaseline,
			"orphanHead", orphanHead,
			"sequencerActive", orphanActive)

		logger.Info("Single-conductor failure scenario verified",
			"chain", chainID,
			"deadLeaderID", oldLeaderID,
			"newLeaderID", newLeaderID)
	}
}
