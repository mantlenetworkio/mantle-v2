package two_follower_opnodes_failure

import (
	"context"
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

// TestTwoFollowerOpNodesFailureAndRecovery is the entry point for the
// paired follower-op-node failure + recovery scenario. Both halves
// share a single orchestrator boot. "Failure" stops both follower
// op-nodes (NOT their conductors) and asserts the README guarantee
// that the active sequencer keeps producing blocks and raft does NOT
// rotate leadership. "Recovery" restarts the two crashed op-nodes and
// asserts the cluster returns to its pre-failure healthy 3-member
// baseline.
//
// Subtests run in declaration order; "Recovery" requires the post-
// "Failure" state and must not be invoked in isolation. To run the
// recovery half:
//
//	go test -run TestTwoFollowerOpNodesFailureAndRecovery
//
// which executes both subtests in order.
func TestTwoFollowerOpNodesFailureAndRecovery(gt *testing.T) {
	parentT := devtest.SerialT(gt)
	sys := presets.NewMantleMinimalWithFaultyConductors(parentT)

	// Suite-wide pre-failure baseline.
	for chainID, conductors := range sys.ConductorSets {
		conductorhelpers.RequireHealthyConductorCluster(parentT, sys.L2Chain, chainID, conductors)
	}

	gt.Run("Failure", func(sgt *testing.T) {
		runFailure(devtest.SerialT(sgt), sys)
	})
	if gt.Failed() {
		return
	}
	gt.Run("Recovery", func(sgt *testing.T) {
		runRecovery(devtest.SerialT(sgt), sys)
	})
}

// runFailure exercises op-conductor README failure mode (table row
// "2 standby are down" sub-scenario):
//
//	"Cluster will still be healthy, active sequencer is still
//	 working, and raft consensus is healthy as well, so no leadership
//	 transfer will happen (standby sequencer is not leader and will
//	 not be able to start leadership transfer process)."
//
// What we crash here is the FOLLOWER op-nodes — not their conductors.
// Their conductors stay healthy, raft heartbeats keep flowing
// conductor-to-conductor, and quorum is preserved. The README's claim
// is that under this failure shape:
//
//   - the active sequencer's local chain keeps advancing, and
//   - no leadership transfer is initiated.
//
// State left for the recovery subtest:
//   - both follower op-nodes: stopped
//   - active sequencer (sys.L2CL): unchanged, still leading
//   - raft leader: unchanged
//
// We deliberately do NOT register a t.Cleanup restart — the recovery
// subtest owns the restart-and-catch-up proof.
func runFailure(t devtest.T, sys *presets.MantleMinimalWithFaultyConductors) {
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestTwoFollowerOpNodesFailureAndRecovery/Failure",
	)
	r := t.Require()

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; this scenario needs >= 3 "+
				"to have any standby sequencers to crash",
				chainID, len(conductors))
			continue
		}

		// 1+2. Identify followers as every CL except sys.L2CL (the
		//      active sequencer captured by the preset's
		//      match.WithSequencerActive).
		activeCLID := sys.L2CL.Escape().ID()
		allCLs := sys.L2Chain.Escape().L2CLNodes()
		r.Equal(3, len(allCLs),
			"chain %s: expected exactly 3 L2 CL nodes, got %d",
			chainID, len(allCLs))

		var followerCLs []*dsl.L2CLNode
		var followerIDs []stack.L2CLNodeID
		for _, c := range allCLs {
			if c.ID() == activeCLID {
				continue
			}
			followerCLs = append(followerCLs, dsl.NewL2CLNode(c, sys.ControlPlane))
			followerIDs = append(followerIDs, c.ID())
		}
		r.Len(followerCLs, 2,
			"chain %s: expected 2 standby CLs after excluding active %s, got %d",
			chainID, activeCLID, len(followerCLs))

		activeProbeCtx, cancelActiveProbe := context.WithTimeout(
			t.Ctx(), 3*time.Second)
		isActive, err := sys.L2CL.Escape().RollupAPI().SequencerActive(activeProbeCtx)
		cancelActiveProbe()
		r.NoError(err,
			"chain %s: SequencerActive RPC failed on supposed leader %s",
			chainID, activeCLID)
		r.True(isActive,
			"chain %s: sys.L2CL %s is not actively sequencing — preset "+
				"selection diverged from cluster state",
			chainID, activeCLID)

		// 3. Capture the baseline leader unsafe head and the current
		//    raft leader ID. Both must hold steady through the
		//    follower-down window (modulo head advancement).
		baseline := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		leaderInfoBefore := conductors[0].FetchLeader()
		r.NotNil(leaderInfoBefore,
			"chain %s: cluster has no leader before stopping followers",
			chainID)
		leaderIDBefore := leaderInfoBefore.ID

		logger.Info("Stopping standby op-nodes",
			"chain", chainID,
			"activeCL", activeCLID,
			"followers", followerIDs,
			"raftLeader", leaderIDBefore,
			"baselineUnsafe", baseline)

		for _, f := range followerCLs {
			f.Stop()
		}

		// 4. Observation window: ~5 blocks at 2s block time, plus
		//    margin for the action-loop tick that drives sequencing.
		const observationWindow = 10 * time.Second
		time.Sleep(observationWindow)

		// 4a. Active leader's chain advanced.
		const expectedDelta uint64 = 3
		head := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		r.GreaterOrEqual(head, baseline+expectedDelta,
			"chain %s: leader EL unsafe head did not advance past "+
				"baseline+%d (baseline=%d, head=%d) while followers were "+
				"stopped — this is the README guarantee that 'active "+
				"sequencer is still working'",
			chainID, expectedDelta, baseline, head)

		// 4b. Raft did not transfer leadership.
		leaderInfoAfter := conductors[0].FetchLeader()
		r.NotNil(leaderInfoAfter,
			"chain %s: cluster lost the leader during follower-stop window",
			chainID)
		r.Equal(leaderIDBefore, leaderInfoAfter.ID,
			"chain %s: raft transferred leadership during follower-stop "+
				"window (%s -> %s); README says no transfer should occur "+
				"because standby sequencers cannot initiate transfers and "+
				"raft heartbeats are conductor-to-conductor",
			chainID, leaderIDBefore, leaderInfoAfter.ID)

		logger.Info("README scenario verified: 2 standby sequencers down, "+
			"active still working, no leadership transfer",
			"chain", chainID,
			"raftLeader", leaderInfoAfter.ID,
			"baselineUnsafe", baseline,
			"observedUnsafe", head,
			"advancedBy", head-baseline)
	}
}
