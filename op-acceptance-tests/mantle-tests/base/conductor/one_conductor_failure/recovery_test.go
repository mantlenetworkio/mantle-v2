package one_conductor_failure

import (
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// runRecovery is the recovery half of the paired single-conductor-
// failure scenario. It picks up the cluster state left by runFailure —
// namely:
//
//   - The pre-failure leader's conductor is stopped (its op-conductor
//     service is gone), but its op-node and op-geth are still running
//     as an orphan sequencer (SequencerActive=true, EL on the new
//     leader's canonical chain via inbound P2P gossip).
//   - Leadership has rotated to a healthy survivor, which is now
//     actively sequencing.
//   - deadLeaders[chainID] holds the dead conductor's ID so we know
//     which one to restart.
//
// Recovery (per the project's strict definition) means BOTH:
//
//  1. A sequencer is producing blocks (the rotated leader; leadership
//     does NOT roll back), AND
//  2. The conductor cluster is healthy with 3 members (all 3 op-nodes
//     synced to within 1 block of the leader).
//
// We restart the dead conductor (the operator's recovery action). Since
// `Stop` does NOT call RemoveServer, the on-disk raft state in
// cfg.RaftStorageDir is intact and the rejoining server reattaches to
// its existing entry under the new leader. Then we re-run
// RequireHealthyConductorCluster — the same baseline check the parent
// test asserted before "Failure". A pass proves:
//
//   - the rejoined conductor's action loop sees IsLeader=false and
//     calls StopSequencer on the orphan op-node (so it stops producing
//     locally),
//   - the orphan EL stays caught up via its existing P2P gossip path,
//   - the cluster is back to its pre-failure healthy steady state.
func runRecovery(t devtest.T, sys *presets.MantleMinimalWithFaultyConductors, deadLeaders map[stack.L2NetworkID]stack.ConductorID) {
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestOneConductorFailureAndRecovery/Recovery",
	)
	r := t.Require()

	for chainID, conductors := range sys.ConductorSets {
		deadID, ok := deadLeaders[chainID]
		r.True(ok,
			"chain %s: no dead-leader handle was recorded by Failure; "+
				"Recovery cannot run in isolation",
			chainID)
		deadSysgo, ok := sys.SysgoConductors[deadID]
		r.True(ok,
			"chain %s: in-process wrapper missing for dead conductor %s",
			chainID, deadID)

		// Recovery action: restart the crashed leader's conductor.
		logger.Info("Recovery: restarting crashed leader conductor",
			"chain", chainID, "victim", deadID)
		deadSysgo.Start()

		// Wait for the rejoined conductor's RPC to come back. Until its
		// HTTP listener is bound, dsl queries against it would error.
		const rpcReadyBudget = 15 * time.Second
		rpcDeadline := time.Now().Add(rpcReadyBudget)
		var rpcReady bool
		for time.Now().Before(rpcDeadline) {
			if deadSysgo.IsRunning() && deadSysgo.HTTPEndpoint() != "" {
				rpcReady = true
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.True(rpcReady,
			"chain %s: restarted conductor %s did not bind RPC within %s",
			chainID, deadID, rpcReadyBudget)

		// Find the dsl wrappers for the new (post-rotation) leader and
		// the recovered conductor — we need both to assert that the
		// recovered conductor stabilizes as a follower, not as a leader.
		var leaderDsl, recoveredDsl *dsl.Conductor
		for _, c := range conductors {
			if c.Escape().ID() == deadID {
				recoveredDsl = c
				continue
			}
			if c.IsLeader() {
				leaderDsl = c
			}
		}
		r.NotNil(recoveredDsl,
			"chain %s: dsl.Conductor for recovered conductor %s missing",
			chainID, deadID)
		r.NotNil(leaderDsl,
			"chain %s: no live leader found among survivors after recovery",
			chainID)

		// The recovered conductor MUST report IsLeader=false — leadership
		// rotated during Failure and does NOT roll back on rejoin.
		const followerBudget = 15 * time.Second
		followerDeadline := time.Now().Add(followerBudget)
		var asFollower bool
		for time.Now().Before(followerDeadline) {
			if !recoveredDsl.IsLeader() && leaderDsl.IsLeader() {
				asFollower = true
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.True(asFollower,
			"chain %s: recovered conductor %s did not stabilize as a "+
				"follower under post-rotation leader %s within %s",
			chainID, deadID, leaderDsl.Escape().ID(), followerBudget)

		// THE recovery assertion: re-run RequireHealthyConductorCluster.
		// This proves both criteria of the recovery definition:
		//   (1) a sequencer is producing blocks (the rotated leader; the
		//       helper finds whoever is leading and asserts EL advance),
		//       AND
		//   (2) the conductor cluster is healthy with 3 members — the
		//       restarted conductor's action loop has driven the orphan
		//       op-node into a clean follower state and the orphan EL
		//       is within 1 block of the leader's unsafe head.
		conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)
		logger.Info("Recovery verified: cluster restored to baseline",
			"chain", chainID,
			"recoveredConductor", deadID,
			"stableLeader", leaderDsl.Escape().ID())
	}
}
