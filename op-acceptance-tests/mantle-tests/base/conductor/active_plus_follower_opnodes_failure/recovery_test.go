package active_plus_follower_opnodes_failure

import (
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// runRecovery is the recovery half of the paired 2-of-3 op-node failure
// scenario. It picks up the cluster state left by runFailure — namely:
//
//   - sys.L2CL is stopped (was the active sequencer at test start);
//   - one follower op-node is stopped, identified via
//     deadFollowerCLs[chainID];
//   - every voter's conductor reports SequencerHealthy=false (the
//     sysgo-faithful peer-degraded state established by Failure: the
//     lone surviving op-node has zero peers because its only
//     static-mesh peers were the two killed op-nodes, so its
//     conductor's MinPeerCount=1 check fails);
//   - the lone-live-op-node voter is latched into escape-hatch
//     sequencing (op-conductor/conductor/service.go:748 + the
//     shouldWaitForHealthRecovery latch at service.go:760), so the
//     chain has been advancing on its EL even though the conductor
//     reports SequencerHealthy=false.
//
// Recovery (per the project's strict definition) means BOTH:
//
//  1. A sequencer is producing blocks (whichever voter the cluster
//     elects once the peer-count constraint is satisfied), AND
//  2. The conductor cluster is healthy with 3 members (all 3 op-nodes
//     synced to within 1 block of the leader).
//
// The recovery mechanism is straightforward: restarting the two
// crashed op-nodes restores the A↔B, A↔C, B↔C static mesh, every
// op-node reconnects to its 2 peers, every conductor's health
// monitor flips SequencerHealthy=true on its next tick (the latched
// survivor exits the (T,F,T) shape and settles into (T,T,T) — the
// shouldWaitForHealthRecovery latch is naturally released because
// healthy=T no longer takes the unhealthy-active branch), and the
// survivor continues sequencing as a now-genuinely-healthy leader.
// The two recovering op-nodes catch up to the leader's head via the
// standard P2P unsafe-payload subscription path: each op-node's
// onUnsafeL2Payload feeds gossiped blocks through engine_newPayload +
// FCU into its local op-geth. No conductor membership change is
// needed — neither conductor was ever stopped, just their op-nodes.
func runRecovery(t devtest.T, sys *presets.MantleMinimalWithFaultyConductors, deadFollowerCLs map[stack.L2NetworkID]stack.L2CLNodeID) {
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestActivePlusFollowerOpNodesFailureAndRecovery/Recovery",
	)
	r := t.Require()

	for chainID, conductors := range sys.ConductorSets {
		deadFollowerID, ok := deadFollowerCLs[chainID]
		r.True(ok,
			"chain %s: Failure did not record a dead follower CL; "+
				"Recovery cannot run in isolation",
			chainID)

		// Resolve the dead-follower L2CLNode from the chain's CL set.
		var deadFollowerCL *dsl.L2CLNode
		for _, c := range sys.L2Chain.Escape().L2CLNodes() {
			if c.ID() == deadFollowerID {
				deadFollowerCL = dsl.NewL2CLNode(c, sys.ControlPlane)
				break
			}
		}
		r.NotNil(deadFollowerCL,
			"chain %s: could not locate L2CLNode for dead follower %s",
			chainID, deadFollowerID)

		// Recovery action: restart both crashed op-nodes (the original
		// active sequencer + the dead follower).
		logger.Info("Recovery: restarting both crashed op-nodes",
			"chain", chainID,
			"victim1", sys.L2CL.Escape().ID(),
			"victim2", deadFollowerID)
		sys.L2CL.Start()
		deadFollowerCL.Start()

		// THE recovery assertion: re-run RequireHealthyConductorCluster.
		// This proves both criteria of the recovery definition:
		//   (1) a sequencer is producing blocks (the rotated leader; the
		//       helper finds whoever is leading and asserts EL advance),
		//       AND
		//   (2) the conductor cluster is healthy with 3 members — both
		//       restarted op-nodes have caught up to within 1 block of
		//       the leader's unsafe head via P2P gossip + engine_newPayload.
		conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)

		logger.Info("Recovery verified: cluster restored to baseline",
			"chain", chainID,
			"recovered1", sys.L2CL.Escape().ID(),
			"recovered2", deadFollowerID)
	}
}
