package active_opnode_failure

import (
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// runRecovery is the recovery half of the paired active-op-node failure
// scenario. It picks up the cluster state left by runFailure — namely:
//
//   - sys.L2CL is stopped (its op-node is dead),
//   - leadership has rotated from sys.L2CL's pair conductor to a healthy
//     follower, which is now actively sequencing.
//
// Recovery (per the project's strict definition) means BOTH:
//
//  1. A sequencer is producing blocks (the rotated leader, since
//     leadership does NOT roll back), AND
//  2. The conductor cluster is healthy with 3 members (all 3 op-nodes
//     synced to within 1 block of the leader).
//
// We restart sys.L2CL (the operator's recovery action) and then
// re-assert RequireHealthyConductorCluster — the same baseline check
// the parent test asserted before "Failure" ran. A pass proves the
// rejoined op-node caught up to the rotated leader's unsafe head via
// P2P gossip + engine_newPayload, and that the cluster is back to its
// pre-failure healthy steady state.
func runRecovery(t devtest.T, sys *presets.MantleMinimalWithFaultyConductors) {
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestActiveOpNodeFailureAndRecovery/Recovery",
	)

	// Recovery action: restart the crashed op-node. Its conductor was
	// never killed, so the conductor's action loop will see its op-node
	// come back, observe IsLeader=false (because leadership rotated in
	// the failure subtest) and call StopSequencer on it; the op-node
	// then syncs from the rotated leader's gossip via the standard
	// onUnsafeL2Payload path.
	logger.Info("Recovery: restarting crashed active op-node",
		"victim", sys.L2CL.Escape().ID())
	sys.L2CL.Start()

	// THE recovery assertion: re-run RequireHealthyConductorCluster.
	// This proves both criteria of the recovery definition:
	//   (1) a sequencer is producing blocks (the rotated leader; the
	//       helper finds whoever is leading and asserts EL advance),
	//       AND
	//   (2) the conductor cluster is healthy with 3 members — the
	//       restarted op-node has caught up to within 1 block of the
	//       leader's unsafe head.
	for chainID, conductors := range sys.ConductorSets {
		conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)
		logger.Info("Recovery verified: cluster restored to baseline",
			"chain", chainID,
			"recoveredOpNode", sys.L2CL.Escape().ID())
	}
}
