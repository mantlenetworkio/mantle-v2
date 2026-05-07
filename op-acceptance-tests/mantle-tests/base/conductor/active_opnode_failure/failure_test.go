package active_opnode_failure

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-conductor/consensus"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// TestActiveOpNodeFailureAndRecovery is the entry point for the paired
// active-op-node failure + recovery scenario. The two subtests share a
// single orchestrator boot — and therefore a single conductor cluster
// state machine. "Failure" drives the cluster into the rotated-leader
// state described by op-conductor README #1, and "Recovery" picks up
// that state and asserts the cluster returns to a healthy 3-member
// baseline after the operator restarts the crashed op-node.
//
// Subtests run in declaration order. "Recovery" REQUIRES the post-
// "Failure" state and must not be invoked in isolation (`go test
// -run .../Recovery` will skip the failure setup and fail the
// pre-recovery sanity check). To run the recovery half explicitly:
//
//	go test -run TestActiveOpNodeFailureAndRecovery
//
// which executes both subtests in order.
func TestActiveOpNodeFailureAndRecovery(gt *testing.T) {
	parentT := devtest.SerialT(gt)
	sys := presets.NewMantleMinimalWithFaultyConductors(parentT)

	// Suite-wide pre-failure baseline. We assert it ONCE here so that
	// neither subtest re-asserts it; "Failure" then drives degradation
	// and "Recovery" asserts the return to baseline at the end.
	for chainID, conductors := range sys.ConductorSets {
		conductorhelpers.RequireHealthyConductorCluster(parentT, sys.L2Chain, chainID, conductors)
	}

	gt.Run("Failure", func(sgt *testing.T) {
		runFailure(devtest.SerialT(sgt), sys)
	})
	if gt.Failed() {
		// Recovery is meaningful only on top of a successful failure
		// drive. Skip rather than running it on a poisoned cluster.
		return
	}
	gt.Run("Recovery", func(sgt *testing.T) {
		runRecovery(devtest.SerialT(sgt), sys)
	})
}

// runFailure exercises op-conductor README failure scenario #1
// (op-conductor/README.md:63):
//
//	"1 sequencer down, stopped producing blocks → Conductor will detect
//	 sequencer failure and start to transfer leadership to another
//	 node, which will start sequencing instead."
//
// Sysgo caveat — modeling "Conductor will detect"
//
// On the sysgo backend, op-conductor is wired to a no-op
// SequencerHealthMonitor (op-devstack/sysgo/l2_conductor.go:237 +
// noopHealthMonitor at line 391). That monitor never produces health
// events, so SequencerHealthy stays true and the conductor's automatic
// "unhealthy → transfer leadership" action loop never triggers. We
// therefore cannot cause AUTOMATIC failover on sysgo — what we drive
// here is the post-detection consequence: once leadership is
// transferred to a healthy node (which is exactly what either a real
// health monitor or a human operator would do), that node MUST resume
// sequencing and the L2 chain MUST advance again.
//
// State left for the recovery subtest:
//   - sys.L2CL: stopped (its op-node is dead)
//   - leadership: rotated to a healthy follower; sys.L2CL's pair
//     conductor is now a follower
func runFailure(t devtest.T, sys *presets.MantleMinimalWithFaultyConductors) {
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestActiveOpNodeFailureAndRecovery/Failure",
	)
	r := t.Require()

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; failover scenario needs "+
				">= 3 to have any healthy follower to promote",
				chainID, len(conductors))
			continue
		}

		// 1. Identify the active sequencer's CL/EL via the preset, and
		//    sanity-probe that it really is sequencing right now.
		activeCLID := sys.L2CL.Escape().ID()
		activeProbeCtx, cancelActiveProbe := context.WithTimeout(
			t.Ctx(), 3*time.Second)
		isActive, err := sys.L2CL.Escape().RollupAPI().SequencerActive(
			activeProbeCtx)
		cancelActiveProbe()
		r.NoError(err,
			"chain %s: SequencerActive RPC failed on supposed leader %s",
			chainID, activeCLID)
		r.True(isActive,
			"chain %s: sys.L2CL %s is not actively sequencing — preset "+
				"selection diverged from cluster state",
			chainID, activeCLID)

		// 2. Capture the current raft leader.
		oldLeaderInfo := conductors[0].FetchLeader()
		r.NotNil(oldLeaderInfo,
			"chain %s: cluster has no leader before failover", chainID)
		oldLeaderID := oldLeaderInfo.ID
		r.NotEmpty(oldLeaderID,
			"chain %s: empty leader ID before failover", chainID)

		// 3. Pick a healthy follower as the failover target.
		membership := conductors[0].FetchClusterMembership()

		var oldLeaderDsl, newLeaderDsl *dsl.Conductor
		var newLeaderInfo consensus.ServerInfo
		for _, c := range conductors {
			id := strings.TrimPrefix(
				c.String(), stack.ConductorKind.String()+"-")
			if id == oldLeaderID {
				oldLeaderDsl = c
				continue
			}
			if newLeaderDsl != nil {
				continue
			}
			// First non-leader voter wins. Skip nonvoters: transferring
			// to a nonvoter is rejected by raft.
			for _, mi := range membership.Servers {
				if mi.ID == id && mi.Suffrage == consensus.Voter {
					newLeaderInfo = mi
					newLeaderDsl = c
					break
				}
			}
		}
		r.NotNil(oldLeaderDsl,
			"chain %s: dsl.Conductor for old leader %s not found",
			chainID, oldLeaderID)
		r.NotNil(newLeaderDsl,
			"chain %s: no eligible voter to promote as new leader",
			chainID)

		baseline := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		logger.Info("Pre-failover cluster state",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"newLeaderID", newLeaderInfo.ID,
			"activeCL", activeCLID,
			"baselineUnsafe", baseline)

		// 4. Stop the active sequencer's op-node. We deliberately do
		//    NOT register a t.Cleanup restart — the recovery subtest
		//    owns the restart-and-rejoin proof.
		sys.L2CL.Stop()

		// 5. Trigger the failover.
		logger.Info("Transferring leadership to healthy follower",
			"chain", chainID,
			"from", oldLeaderID,
			"to", newLeaderInfo.ID)
		oldLeaderDsl.TransferLeadershipTo(newLeaderInfo)

		// Wait until raft reports the transfer is complete on both
		// ends.
		transferDeadline := time.Now().Add(10 * time.Second)
		var transferred bool
		for time.Now().Before(transferDeadline) {
			if newLeaderDsl.IsLeader() && !oldLeaderDsl.IsLeader() {
				transferred = true
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.True(transferred,
			"chain %s: leadership transfer %s -> %s did not complete "+
				"within 10s — README scenario #1 cannot recover if raft "+
				"refuses to promote the healthy follower",
			chainID, oldLeaderID, newLeaderInfo.ID)

		newLeaderEL := conductorhelpers.ELPairedWithConductor(sys.L2Chain, newLeaderInfo.ID)
		r.NotNil(newLeaderEL,
			"chain %s: could not locate L2EL paired with new leader %s",
			chainID, newLeaderInfo.ID)

		const observationWindow = 12 * time.Second
		time.Sleep(observationWindow)

		// 5c. Block production resumed on the new leader's EL.
		const expectedDelta uint64 = 3
		head := newLeaderEL.BlockRefByLabel(eth.Unsafe).Number
		r.GreaterOrEqual(head, baseline+expectedDelta,
			"chain %s: new leader EL %s unsafe head did not advance past "+
				"baseline+%d (baseline=%d, head=%d) after failover — "+
				"this is the README guarantee that 'the new leader will "+
				"start sequencing instead'",
			chainID, newLeaderEL.Escape().ID(),
			expectedDelta, baseline, head)

		// 5d. The new leader is in fact actively sequencing.
		newLeaderCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, newLeaderInfo.ID)
		r.NotNil(newLeaderCL,
			"chain %s: could not locate L2CL paired with new leader %s",
			chainID, newLeaderInfo.ID)
		seqProbeCtx, cancelSeqProbe := context.WithTimeout(
			t.Ctx(), 3*time.Second)
		newLeaderActive, err := newLeaderCL.RollupAPI().SequencerActive(
			seqProbeCtx)
		cancelSeqProbe()
		r.NoError(err,
			"chain %s: SequencerActive RPC failed on new leader %s",
			chainID, newLeaderInfo.ID)
		r.True(newLeaderActive,
			"chain %s: new leader CL %s reports SequencerActive=false "+
				"after failover — conductor leader-takeover did not "+
				"propagate into op-node sequencing state",
			chainID, newLeaderCL.ID())

		logger.Info("README scenario #1 verified: active op-node down, "+
			"leadership transferred to healthy follower, chain advancing",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"newLeaderID", newLeaderInfo.ID,
			"baselineUnsafe", baseline,
			"observedUnsafe", head,
			"advancedBy", head-baseline)
	}
}
