package active_plus_follower_opnodes_failure

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

// TestActivePlusFollowerOpNodesFailureAndRecovery is the entry point
// for the paired 2-of-3 op-node failure + recovery scenario. Both
// halves share a single orchestrator boot. "Failure" stops the active
// sequencer's op-node + one follower op-node and asserts leadership
// can be transferred to the lone surviving healthy follower (README
// scenario #3a). "Recovery" restarts both crashed op-nodes and asserts
// the cluster returns to the pre-failure healthy 3-member baseline.
//
// Subtests run in declaration order; "Recovery" requires the post-
// "Failure" state and must not be invoked in isolation. To run the
// recovery half:
//
//	go test -run TestActivePlusFollowerOpNodesFailureAndRecovery
//
// which executes both subtests in order.
func TestActivePlusFollowerOpNodesFailureAndRecovery(gt *testing.T) {
	parentT := devtest.SerialT(gt)
	sys := presets.NewMantleMinimalWithFaultyConductors(parentT)

	// Suite-wide pre-failure baseline.
	for chainID, conductors := range sys.ConductorSets {
		conductorhelpers.RequireHealthyConductorCluster(parentT, sys.L2Chain, chainID, conductors)
	}

	// State carried across subtests: which follower CL was crashed
	// alongside sys.L2CL during Failure, so Recovery can restart the
	// same one. Keyed by chain ID for multi-chain robustness.
	deadFollowerCLs := map[stack.L2NetworkID]stack.L2CLNodeID{}

	gt.Run("Failure", func(sgt *testing.T) {
		runFailure(devtest.SerialT(sgt), sys, deadFollowerCLs)
	})
	if gt.Failed() {
		return
	}
	gt.Run("Recovery", func(sgt *testing.T) {
		runRecovery(devtest.SerialT(sgt), sys, deadFollowerCLs)
	})
}

// runFailure exercises op-conductor README failure scenario #3,
// sub-scenario "Scenario #1" (op-conductor/README.md ~line 65):
//
//	"2 or more sequencer down (regardless of if active sequencer is
//	 healthy): Scenario #1: active sequencer and 1 follower are down.
//
//	 Solution: At this point, the Conductor will notice the sequencer
//	 not being healthy and start to transfer leadership to another
//	 node."
//
// What this test pins down
//
// Given 2-of-3 sequencer op-nodes are down (the active one + one
// follower), the cluster is NOT bricked: leadership can be successfully
// transferred to the single surviving healthy follower, that follower
// starts sequencing, and the L2 chain resumes block production.
//
// State left for the recovery subtest:
//   - sys.L2CL: stopped (was the active sequencer at test start).
//   - One follower op-node: stopped (the dead follower we picked).
//     Recorded in deadFollowerCLs[chainID] so Recovery knows which
//     one to restart.
//   - Leadership: rotated to the surviving healthy follower; that
//     node is now actively sequencing.
//   - Conductors: all 3 still running (we only killed op-nodes).
func runFailure(t devtest.T, sys *presets.MantleMinimalWithFaultyConductors, deadFollowerCLs map[stack.L2NetworkID]stack.L2CLNodeID) {
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestActivePlusFollowerOpNodesFailureAndRecovery/Failure",
	)
	r := t.Require()

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; 2-of-3 failover test "+
				"needs >= 3 to have a surviving healthy follower",
				chainID, len(conductors))
			continue
		}

		// 1+2. Identify A (active sequencer) and sanity-probe its state.
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
			"chain %s: sys.L2CL %s reports SequencerActive=false at "+
				"test start; the failover scenario requires A to be "+
				"actually sequencing before we kill it",
			chainID, activeCLID)

		oldLeaderInfo := conductors[0].FetchLeader()
		r.NotNil(oldLeaderInfo,
			"chain %s: cluster has no leader before failover", chainID)
		oldLeaderID := oldLeaderInfo.ID
		r.NotEmpty(oldLeaderID, "chain %s: empty leader ID", chainID)

		// Find the dsl.Conductor for the old leader and pick TWO
		// followers from the voter set.
		membership := conductors[0].FetchClusterMembership()

		var oldLeaderDsl, deadFollowerDsl, newLeaderDsl *dsl.Conductor
		var newLeaderInfo consensus.ServerInfo
		var deadFollowerID string
		for _, c := range conductors {
			id := strings.TrimPrefix(
				c.String(), stack.ConductorKind.String()+"-")
			if id == oldLeaderID {
				oldLeaderDsl = c
				continue
			}
			isVoter := false
			var memberInfo consensus.ServerInfo
			for _, mi := range membership.Servers {
				if mi.ID == id && mi.Suffrage == consensus.Voter {
					isVoter = true
					memberInfo = mi
					break
				}
			}
			if !isVoter {
				continue
			}
			if deadFollowerDsl == nil {
				deadFollowerDsl = c
				deadFollowerID = id
				continue
			}
			if newLeaderDsl == nil {
				newLeaderDsl = c
				newLeaderInfo = memberInfo
				continue
			}
		}
		r.NotNil(oldLeaderDsl,
			"chain %s: dsl.Conductor for old leader %s not found",
			chainID, oldLeaderID)
		r.NotNil(deadFollowerDsl,
			"chain %s: could not find a follower to label as the "+
				"'dead' arm of the 2-of-3 failure", chainID)
		r.NotNil(newLeaderDsl,
			"chain %s: could not find a second follower to promote — "+
				"need exactly 1 leader + 1 dead follower + 1 healthy "+
				"follower", chainID)
		_ = oldLeaderDsl // referenced via TransferLeadershipTo below

		// 3. Resolve the (CL, EL) pairs for A, B and C.
		deadFollowerCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, deadFollowerID)
		r.NotNil(deadFollowerCL,
			"chain %s: could not locate L2CL paired with dead "+
				"follower conductor %s", chainID, deadFollowerID)
		newLeaderEL := conductorhelpers.ELPairedWithConductor(sys.L2Chain, newLeaderInfo.ID)
		r.NotNil(newLeaderEL,
			"chain %s: could not locate L2EL paired with promotion "+
				"target %s", chainID, newLeaderInfo.ID)
		newLeaderCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, newLeaderInfo.ID)
		r.NotNil(newLeaderCL,
			"chain %s: could not locate L2CL paired with promotion "+
				"target %s", chainID, newLeaderInfo.ID)

		deadFollowerCLDsl := dsl.NewL2CLNode(deadFollowerCL, sys.ControlPlane)

		baseline := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		logger.Info("Pre-failover cluster state",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"deadFollowerID", deadFollowerID,
			"newLeaderID", newLeaderInfo.ID,
			"activeCL", activeCLID,
			"baselineUnsafe", baseline)

		// 4. Stop A's op-node AND the chosen dead-follower's op-node.
		//    We deliberately do NOT register a t.Cleanup restart —
		//    the recovery subtest owns the restart-and-rejoin proof.
		sys.L2CL.Stop()
		deadFollowerCLDsl.Stop()
		// Record the dead follower CL ID for Recovery to pick up.
		// Done immediately after Stop so that even if subsequent
		// assertions fail, the parent test's gt.Failed() short-circuit
		// will skip Recovery (no risk of running Recovery against a
		// poisoned cluster).
		deadFollowerCLs[chainID] = deadFollowerCL.ID()

		// 5. Trigger failover.
		logger.Info("Transferring leadership to surviving healthy follower",
			"chain", chainID,
			"from", oldLeaderID,
			"deadFollower", deadFollowerID,
			"to", newLeaderInfo.ID)
		oldLeaderDsl.TransferLeadershipTo(newLeaderInfo)

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
				"within 10s — README scenario #3a cannot recover if raft "+
				"refuses to promote the surviving healthy follower",
			chainID, oldLeaderID, newLeaderInfo.ID)

		// 5a. Wait for the new leader's conductor action loop to call
		//     StartSequencer on its op-node.
		const startDeadlineDur = 10 * time.Second
		startDeadline := time.Now().Add(startDeadlineDur)
		var newActive bool
		var lastNewErr error
		for time.Now().Before(startDeadline) {
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
			"chain %s, new leader CL %s: SequencerActive=false %s after "+
				"leadership transfer — the surviving follower never "+
				"started sequencing, so the cluster is effectively "+
				"bricked despite raft having a leader",
			chainID, newLeaderCL.ID(), startDeadlineDur)

		// 5b. Block production on the new leader's EL.
		const observationWindow = 12 * time.Second
		time.Sleep(observationWindow)

		const expectedDelta uint64 = 3
		head := newLeaderEL.BlockRefByLabel(eth.Unsafe).Number
		r.GreaterOrEqual(head, baseline+expectedDelta,
			"chain %s: new leader EL %s unsafe head did not advance past "+
				"baseline+%d (baseline=%d, head=%d) after failover with "+
				"2-of-3 op-nodes down — README scenario #3a's promise "+
				"that recovery to the surviving healthy follower works "+
				"is broken",
			chainID, newLeaderEL.Escape().ID(),
			expectedDelta, baseline, head)

		logger.Info("README scenario #3a verified: with active sequencer "+
			"+ 1 follower down, leadership transferred to surviving "+
			"healthy follower; chain advancing",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"deadFollowerID", deadFollowerID,
			"newLeaderID", newLeaderInfo.ID,
			"baselineUnsafe", baseline,
			"observedUnsafe", head,
			"advancedBy", head-baseline)
	}
}
