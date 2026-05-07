package two_follower_opnodes_failure

import (
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// runRecovery is the recovery half of the paired follower-op-node
// failure scenario. It picks up the cluster state left by runFailure —
// namely:
//
//   - both follower op-nodes are stopped;
//   - the active sequencer (sys.L2CL) is unchanged and still leading;
//   - raft leader is unchanged (no rotation in this scenario).
//
// Recovery (per the project's strict definition) means BOTH:
//
//  1. A sequencer is producing blocks (the original leader; never
//     rotated in this scenario), AND
//  2. The conductor cluster is healthy with 3 members (all 3 op-nodes
//     synced to within 1 block of the leader).
//
// During the failure window the op-nodes ARE the lagging members; raft
// itself is fine (heartbeats are conductor-to-conductor and quorum is
// preserved), but the two follower ELs cannot apply new payloads
// because their op-nodes are dead. After restart, op-node resubscribes
// to the unsafe-payload P2P topic and feeds gossiped payloads through
// engine_newPayload + FCU into the local op-geth — that is the
// op-conductor side recovery path we exercise.
//
// We re-derive the follower CL set in Recovery by the same rule
// runFailure used (every L2 CL except sys.L2CL). Since runFailure
// asserted no leadership rotation occurred, sys.L2CL is still the
// active sequencer, so the "everything-except-active" set picks the
// same two stopped op-nodes and we don't need to thread state across
// subtests.
func runRecovery(t devtest.T, sys *presets.MantleMinimalWithFaultyConductors) {
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestTwoFollowerOpNodesFailureAndRecovery/Recovery",
	)
	r := t.Require()

	for chainID, conductors := range sys.ConductorSets {
		// Re-derive the followers using the same rule as runFailure.
		activeCLID := sys.L2CL.Escape().ID()
		allCLs := sys.L2Chain.Escape().L2CLNodes()
		r.Equal(3, len(allCLs),
			"chain %s: expected 3 L2 CL nodes, got %d",
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
			"chain %s: expected 2 follower CLs after excluding active %s, got %d",
			chainID, activeCLID, len(followerCLs))

		// Recovery action: restart both crashed follower op-nodes.
		// After Start each op-node will resubscribe to P2P, drive its
		// op-geth via engine_newPayload from inbound gossip, and catch
		// up to the leader's unsafe head.
		logger.Info("Recovery: restarting both crashed follower op-nodes",
			"chain", chainID, "followers", followerIDs)
		for _, f := range followerCLs {
			f.Start()
		}

		// THE recovery assertion: re-run RequireHealthyConductorCluster.
		// This proves both criteria of the recovery definition:
		//   (1) a sequencer is producing blocks (leader was never
		//       rotated — the same leader from baseline still
		//       sequences), AND
		//   (2) the conductor cluster is healthy with 3 members —
		//       both restarted follower op-nodes are caught up to
		//       within 1 block of the leader's unsafe head.
		conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)

		logger.Info("Recovery verified: cluster restored to baseline",
			"chain", chainID,
			"recoveredFollowers", followerIDs)
	}
}
