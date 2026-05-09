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
//   - the active leader's op-node (sys.L2CL) is still alive but its
//     conductor reports SequencerHealthy=false because the static-mesh
//     peers (the two dead follower op-nodes) are gone, so PeerStats
//     returns 0 and MinPeerCount=1 fails;
//   - raft leader may have rotated away from the original leader as
//     its action loop saw (leader && !healthy && active) and called
//     transferLeader();
//   - no node is sustainably sequencing.
//
// Recovery (per the project's strict definition) means BOTH:
//
//  1. A sequencer is producing blocks (whichever voter the cluster
//     currently has as raft leader, once the peer-count constraint is
//     satisfied), AND
//  2. The conductor cluster is healthy with 3 members (all 3 op-nodes
//     synced to within 1 block of the leader).
//
// The recovery mechanism is straightforward: restarting the two
// crashed follower op-nodes restores the A↔B, A↔C, B↔C static mesh,
// every op-node reconnects to its 2 peers, every conductor's health
// monitor flips SequencerHealthy=true on its next tick, and whichever
// voter is currently raft leader resumes sequencing. The two
// recovering op-nodes catch up via standard P2P unsafe-payload
// subscription + engine_newPayload into the local op-geth.
//
// We re-derive the follower CL set in Recovery by the same rule
// runFailure used (every L2 CL except sys.L2CL). sys.L2CL.ID() is
// fixed by the preset, so the "everything-except-sys.L2CL" set
// continues to pick the same two stopped op-nodes regardless of any
// raft leadership rotation that occurred during Failure.
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
		//   (1) a sequencer is producing blocks (whichever voter is
		//       currently raft leader; the helper finds whoever is
		//       leading and asserts EL advance), AND
		//   (2) the conductor cluster is healthy with 3 members —
		//       both restarted follower op-nodes are caught up to
		//       within 1 block of the leader's unsafe head.
		conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)

		logger.Info("Recovery verified: cluster restored to baseline",
			"chain", chainID,
			"recoveredFollowers", followerIDs)
	}
}
