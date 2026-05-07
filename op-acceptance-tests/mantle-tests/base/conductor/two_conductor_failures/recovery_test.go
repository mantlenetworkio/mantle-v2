package two_conductor_failures

import (
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// runRecovery is the recovery half of the paired two-conductor-failure
// scenario. It picks up the cluster state left by runFailure — namely:
//
//   - both follower conductors stopped, so quorum is impossible;
//   - the pre-failure leader has stepped down (IsLeader=false);
//   - deadFollowers[chainID] holds the two crashed conductor IDs so we
//     know which ones to restart.
//
// Recovery (per the project's strict definition) means BOTH:
//
//  1. A sequencer is producing blocks, AND
//  2. The conductor cluster is healthy with 3 members (all 3 op-nodes
//     synced to within 1 block of the leader).
//
// Why no manual override is needed
//
// Op-conductor's RUNBOOK does describe a manual override path
// (`conductor_pause` + `conductor_overrideLeader` +
// `admin_overrideLeader`) for cases where on-disk raft state is
// destroyed or the cluster cannot reach a quorum after restart. But for
// the bog-standard "the two follower conductor processes crashed and we
// restarted them" case, no override is needed: the on-disk raft state
// in cfg.RaftStorageDir is intact, the surviving conductor already has
// the latest log, and once the two restarted conductors rejoin via raft
// heartbeats, quorum is restored automatically and raft re-elects a
// leader (or the prior leader simply re-acquires its lease). This test
// pins down THAT path — the no-override "two-follower restart" recovery.
func runRecovery(t devtest.T, sys *presets.MantleMinimalWithFaultyConductors, deadFollowers map[stack.L2NetworkID][]stack.ConductorID) {
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestTwoConductorFailuresAndRecovery/Recovery",
	)
	r := t.Require()

	for chainID, conductors := range sys.ConductorSets {
		dead, ok := deadFollowers[chainID]
		r.True(ok && len(dead) == 2,
			"chain %s: Failure did not record exactly 2 dead followers; "+
				"Recovery cannot run in isolation",
			chainID)
		victimAID := dead[0]
		victimBID := dead[1]
		victimA, ok := sys.SysgoConductors[victimAID]
		r.True(ok, "chain %s: in-process wrapper missing for victim %s",
			chainID, victimAID)
		victimB, ok := sys.SysgoConductors[victimBID]
		r.True(ok, "chain %s: in-process wrapper missing for victim %s",
			chainID, victimBID)

		// Recovery action: restart both crashed conductors. Their on-
		// disk raft state is intact, so they should re-attach under the
		// existing server entries. Once one of them is back, quorum is
		// restored and a leader can serve commits; once both are back,
		// the cluster is at the documented healthy 3-voter steady state.
		logger.Info("Recovery: restarting both crashed conductors",
			"chain", chainID,
			"victimA", victimAID,
			"victimB", victimBID)
		victimA.Start()
		victimB.Start()

		// Wait for both restarted conductors' RPC servers to bind.
		// Until the HTTP listener is up, dsl.FetchLeader against them
		// would fail.
		const rpcReadyBudget = 15 * time.Second
		rpcDeadline := time.Now().Add(rpcReadyBudget)
		var rpcReady bool
		for time.Now().Before(rpcDeadline) {
			if victimA.IsRunning() && victimA.HTTPEndpoint() != "" &&
				victimB.IsRunning() && victimB.HTTPEndpoint() != "" {
				rpcReady = true
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.True(rpcReady,
			"chain %s: restarted conductors %s and %s did not bind RPC "+
				"within %s",
			chainID, victimAID, victimBID, rpcReadyBudget)

		// Quorum-restored gate: poll any conductor for a STABLE leader.
		// Raft may re-elect a different node than the pre-crash leader
		// (the original leader gave up its lease when its heartbeats
		// started failing); we don't care WHICH node leads, only that
		// the same leader is still leading on two consecutive polls.
		const quorumBudget = 30 * time.Second
		quorumDeadline := time.Now().Add(quorumBudget)
		var quorumLeaderID, prevQuorumLeaderID string
		for time.Now().Before(quorumDeadline) {
			li := conductors[0].FetchLeader()
			if li != nil && li.ID != "" {
				if li.ID == prevQuorumLeaderID {
					quorumLeaderID = li.ID
					break
				}
				prevQuorumLeaderID = li.ID
			} else {
				prevQuorumLeaderID = ""
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.NotEmpty(quorumLeaderID,
			"chain %s: cluster did not converge on a stable leader "+
				"within %s after restarting both crashed conductors — "+
				"quorum was not restored",
			chainID, quorumBudget)
		logger.Info("Quorum restored",
			"chain", chainID, "newLeader", quorumLeaderID)

		// THE recovery assertion: re-run RequireHealthyConductorCluster.
		// This proves both criteria of the recovery definition:
		//   (1) a sequencer is producing blocks, AND
		//   (2) the conductor cluster is healthy with 3 members.
		conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)
		logger.Info("Recovery verified: cluster restored to baseline",
			"chain", chainID,
			"recoveredA", victimAID,
			"recoveredB", victimBID,
			"finalLeader", quorumLeaderID)
	}
}
