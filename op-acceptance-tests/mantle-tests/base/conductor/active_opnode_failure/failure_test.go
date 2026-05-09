package active_opnode_failure

import (
	"context"
	"strings"
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
// With the production SequencerHealthMonitor wired into sysgo conductors
// (op-devstack/sysgo/l2_conductor.go: Conductor.Start passes nil →
// NewOpConductor.initHealthMonitor builds the real monitor), automatic
// failover IS the flow under test: the test only kills the active
// op-node and waits — it does NOT manually transfer leadership. The
// conductor's action loop on the leader observes
// (leader && !healthy && active), stops its sequencer, and calls
// transferLeader(); raft picks a healthy voter as the new leader; that
// new leader's action loop calls startSequencer on its own op-node,
// which begins producing blocks.
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

		// Bind oldLeaderDsl so we can later assert it's NOT the leader
		// post-failover.
		var oldLeaderDsl *dsl.Conductor
		for _, c := range conductors {
			id := strings.TrimPrefix(
				c.String(), stack.ConductorKind.String()+"-")
			if id == oldLeaderID {
				oldLeaderDsl = c
				break
			}
		}
		r.NotNil(oldLeaderDsl,
			"chain %s: dsl.Conductor for old leader %s not found",
			chainID, oldLeaderID)

		baseline := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		logger.Info("Pre-failover cluster state",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"activeCL", activeCLID,
			"baselineUnsafe", baseline)

		// 3. Stop the active sequencer's op-node. We deliberately do
		//    NOT register a t.Cleanup restart — the recovery subtest
		//    owns the restart-and-rejoin proof.
		//
		//    From this point the production health monitor on the
		//    leader conductor will fail its checkNodeSyncStatus poll
		//    (op-node RPC is down) and emit ErrSequencerConnectionDown
		//    on the next tick (cfg.HealthCheck.Interval = 1s). The
		//    leader's action loop then enters the
		//    (leader && !healthy && active) branch, stops its sequencer,
		//    and calls transferLeader().
		sys.L2CL.Stop()

		// 4. Wait for raft to autonomously rotate leadership to a healthy
		//    voter. We poll every conductor's IsLeader to identify the
		//    new leader without relying on the dead leader's RPC (which
		//    may itself be racing the rotation).
		const failoverDeadline = 30 * time.Second
		failoverEnd := time.Now().Add(failoverDeadline)
		var newLeaderDsl *dsl.Conductor
		var newLeaderID string
		for time.Now().Before(failoverEnd) {
			for _, c := range conductors {
				id := strings.TrimPrefix(
					c.String(), stack.ConductorKind.String()+"-")
				if id == oldLeaderID {
					continue
				}
				if c.IsLeader() {
					newLeaderDsl = c
					newLeaderID = id
					break
				}
			}
			if newLeaderDsl != nil {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.NotNil(newLeaderDsl,
			"chain %s: no healthy follower acquired raft leadership "+
				"within %s after the active op-node was stopped — the "+
				"production health monitor's auto-failover did not "+
				"trigger", chainID, failoverDeadline)
		r.False(oldLeaderDsl.IsLeader(),
			"chain %s: old leader %s still reports IsLeader=true after "+
				"%s — leadership transfer did not complete",
			chainID, oldLeaderID, failoverDeadline)

		newLeaderEL := conductorhelpers.ELPairedWithConductor(sys.L2Chain, newLeaderID)
		r.NotNil(newLeaderEL,
			"chain %s: could not locate L2EL paired with new leader %s",
			chainID, newLeaderID)
		newLeaderCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, newLeaderID)
		r.NotNil(newLeaderCL,
			"chain %s: could not locate L2CL paired with new leader %s",
			chainID, newLeaderID)

		logger.Info("Auto-failover detected new leader",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"newLeaderID", newLeaderID)

		// 5. Wait for the new leader's action loop to call
		//    startSequencer on its op-node.
		const sequencerStartDeadline = 15 * time.Second
		startEnd := time.Now().Add(sequencerStartDeadline)
		var newActive bool
		var lastNewErr error
		for time.Now().Before(startEnd) {
			probeCtx, cancelProbe := context.WithTimeout(
				t.Ctx(), 2*time.Second)
			newActive, lastNewErr = newLeaderCL.RollupAPI().
				SequencerActive(probeCtx)
			cancelProbe()
			if lastNewErr == nil && newActive {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.NoError(lastNewErr,
			"chain %s: SequencerActive RPC failed on new leader CL %s",
			chainID, newLeaderCL.ID())
		r.True(newActive,
			"chain %s, new leader CL %s: SequencerActive=false after "+
				"leadership rotated — the surviving healthy follower "+
				"never started sequencing, so the cluster is effectively "+
				"bricked despite raft having a leader",
			chainID, newLeaderCL.ID())

		// 6. Block production resumed on the new leader's EL.
		const observationWindow = 12 * time.Second
		time.Sleep(observationWindow)
		const expectedDelta uint64 = 3
		head := newLeaderEL.BlockRefByLabel(eth.Unsafe).Number
		r.GreaterOrEqual(head, baseline+expectedDelta,
			"chain %s: new leader EL %s unsafe head did not advance past "+
				"baseline+%d (baseline=%d, head=%d) after auto-failover — "+
				"this is the README guarantee that 'the new leader will "+
				"start sequencing instead'",
			chainID, newLeaderEL.Escape().ID(),
			expectedDelta, baseline, head)

		logger.Info("README scenario #1 verified: active op-node down, "+
			"production health monitor auto-rotated leadership to "+
			"healthy follower, chain advancing",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"newLeaderID", newLeaderID,
			"baselineUnsafe", baseline,
			"observedUnsafe", head,
			"advancedBy", head-baseline)
	}
}
