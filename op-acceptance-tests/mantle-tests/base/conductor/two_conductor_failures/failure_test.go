package two_conductor_failures

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// TestTwoConductorFailuresAndRecovery is the entry point for the paired
// quorum-loss + restart scenario. Both halves share a single
// orchestrator boot. "Failure" stops both follower conductors, breaking
// quorum, and asserts the existing leader correctly steps down (the
// "no split-brain on partition" raft guarantee). "Recovery" restarts
// both crashed conductors and asserts the cluster returns to the
// pre-failure healthy 3-member baseline.
//
// Subtests run in declaration order; "Recovery" requires the post-
// "Failure" state and must not be invoked in isolation. To run the
// recovery half:
//
//	go test -run TestTwoConductorFailuresAndRecovery
//
// which executes both subtests in order.
func TestTwoConductorFailuresAndRecovery(gt *testing.T) {
	parentT := devtest.SerialT(gt)
	sys := presets.NewMantleMinimalWithFaultyConductors(parentT)

	if len(sys.SysgoConductors) < 3 {
		parentT.Skipf("scenario needs in-process conductors and >=3 voters; "+
			"got %d sysgo conductors (kurtosis/persistent backends do "+
			"not expose Conductor.Stop)", len(sys.SysgoConductors))
	}

	// Suite-wide pre-failure baseline.
	for chainID, conductors := range sys.ConductorSets {
		conductorhelpers.RequireHealthyConductorCluster(parentT, sys.L2Chain, chainID, conductors)
	}

	// State carried across subtests: which two conductors were crashed
	// per chain, so Recovery can restart the same set. Order within
	// each slice is non-deterministic but irrelevant.
	deadFollowers := map[stack.L2NetworkID][]stack.ConductorID{}

	gt.Run("Failure", func(sgt *testing.T) {
		runFailure(devtest.SerialT(sgt), sys, deadFollowers)
	})
	if gt.Failed() {
		return
	}
	gt.Run("Recovery", func(sgt *testing.T) {
		runRecovery(devtest.SerialT(sgt), sys, deadFollowers)
	})
}

// runFailure exercises op-conductor README failure scenario:
//
//	"2 standby Conductors failed -> Network down (no leader)"
//
// In a 3-voter raft cluster, leadership requires a quorum of 2. If both
// followers die, the leader can no longer commit log entries (no quorum)
// and, after the leader-lease timeout, MUST step down — raft.Leader
// transitions to raft.Candidate/Follower because it can no longer prove
// it is still the leader. The op-conductor wrapper observes this via
// raft.LeaderCh() and stops driving sequencing.
//
// README guarantee being validated:
//   - With 2/3 conductors dead, the cluster MUST NOT continue acting as
//     if it has a leader. Continuing to lead under quorum loss is the
//     classical "split-brain on partition" bug; raft prevents it via
//     leader leases.
//
// State left for the recovery subtest:
//   - Both follower conductors: stopped (their op-conductor services
//     are gone). Their op-nodes/op-geth keep running but raft is dead.
//   - The pre-failure leader: stepped down (IsLeader=false), no
//     successor possible because no quorum.
//   - deadFollowers[chainID]: the two stopped conductor IDs, so
//     Recovery can restart them.
func runFailure(t devtest.T, sys *presets.MantleMinimalWithFaultyConductors, deadFollowers map[stack.L2NetworkID][]stack.ConductorID) {
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestTwoConductorFailuresAndRecovery/Failure",
	)
	r := t.Require()

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; quorum-loss scenario "+
				"needs >= 3", chainID, len(conductors))
			continue
		}

		// 1. Locate the live leader.
		var leaderDsl *dsl.Conductor
		for _, c := range conductors {
			if c.IsLeader() {
				leaderDsl = c
				break
			}
		}
		r.NotNil(leaderDsl,
			"chain %s: no leader found among in-process conductors before "+
				"quorum-loss test", chainID)
		leaderID := leaderDsl.Escape().ID()
		logger.Info("Leader before quorum-loss",
			"chain", chainID, "leaderID", leaderID)

		// 2. Stop every non-leader sysgo conductor on this chain. We
		//    need 2/3 down to provoke the leader-lease step-down.
		var stopped []stack.ConductorID
		for id, sysgoC := range sys.SysgoConductors {
			if id == leaderID {
				continue
			}
			if !sysgoC.IsRunning() {
				continue
			}
			logger.Info("Stopping standby conductor",
				"chain", chainID, "id", id)
			sysgoC.Stop()
			stopped = append(stopped, id)
		}
		r.GreaterOrEqual(len(stopped), 2,
			"chain %s: expected to stop at least 2 standby conductors, "+
				"got %d — cluster topology unexpected", chainID, len(stopped))
		// Record exactly the two we crashed for Recovery to pick up.
		// If the cluster has > 3 conductors (it shouldn't on Mantle
		// minimal but be defensive), we restart only the two we
		// actually stopped — that's faithful to the failure shape.
		deadFollowers[chainID] = stopped[:2]

		// 3. Poll the leader for IsLeader == false.
		//    sysgo's RaftLeaderLeaseTimeout is 500ms (see l2_conductor.go).
		//    Raft semantics: a leader that cannot heartbeat to a quorum
		//    must step down within LeaderLeaseTimeout. We give 15s of
		//    slack for CI scheduling jitter.
		stepDownDeadline := time.Now().Add(15 * time.Second)
		var steppedDown bool
		for time.Now().Before(stepDownDeadline) {
			if !leaderDsl.IsLeader() {
				steppedDown = true
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.True(steppedDown,
			"chain %s: leader %s did NOT step down within 15s after "+
				"losing quorum (this is a serious raft-layer bug — the "+
				"leader is acting as if it still has a quorum)",
			chainID, leaderID)

		logger.Info("Leader correctly stepped down on quorum loss",
			"chain", chainID, "exLeaderID", leaderID)

		// 4. Sanity: with both followers dead, no other conductor should
		//    be elected leader. Walk every dsl.Conductor (those whose
		//    sysgo wrapper is still running) and assert IsLeader is
		//    false. The dead conductors' RPCs would fail, so we skip
		//    them; the once-leader we already verified above.
		for _, c := range conductors {
			id := c.Escape().ID()
			sysgoC, ok := sys.SysgoConductors[id]
			if !ok || !sysgoC.IsRunning() {
				continue
			}
			r.False(c.IsLeader(),
				"chain %s: conductor %s reports leadership while a "+
					"quorum is impossible (cluster size 3, 2 stopped) — "+
					"this would be a split-brain", chainID, id)
		}

		logger.Info("Quorum-loss step-down verified",
			"chain", chainID, "exLeaderID", leaderID)
	}
}
