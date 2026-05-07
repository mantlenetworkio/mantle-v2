package two_conductor_failures_active_plus_follower

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// TestConductorNoLeaderWhenLeaderAndFollowerFail covers the
// "active sequencer + 1 standby Conductor failed" 2-of-3 conductor
// failure pair (complementary to the "2 standbys failed" case in the
// tolerate_two_conductor_failures package): we kill the active
// sequencer's conductor PLUS one follower. The remaining single
// follower is the only voter alive.
//
// Why this is a distinct scenario from the "2 standbys failed" case:
//
//   - "2 standbys failed": the leader survives. Raft demotes it via
//     LeaderLeaseTimeout because its AppendEntries to the now-dead
//     followers cannot get a quorum acknowledgement.
//   - This test: the leader is killed. The surviving follower starts
//     RequestVote elections to become leader — but a 3-voter cluster
//     needs 2 votes (its own + 1 peer) to win, and both peers are dead.
//     Per Raft §5.4 (election restriction) the candidate stays stuck in
//     the candidate→follower oscillation, and never reaches Leader.
//
// Both code paths converge on the same external invariant — "no
// conductor in the cluster reports IsLeader == true" — which is what we
// assert. The README guarantee being validated is:
//
//	"Active sequencer + standby Conductor failed -> Network down (no leader)"
//
// "Network down" mechanism (why the leader's EL freezes even though
// only its conductor — not its op-node — was killed):
//
//   - The leader's op-node still has SequencerActive=true. It enters
//     each block's build phase and emits BuildSealEvent.
//   - Sequencer.onBuildSealed (op-node/rollup/sequencing/sequencer.go:283)
//     calls ConductorClient.CommitUnsafePayload BEFORE gossip and BEFORE
//     PayloadProcessEvent. With the leader's local conductor stopped,
//     the RPC dial fails on every attempt (the embedded retry logic in
//     ConductorClient is 2 × 50ms then errors out).
//   - On the commit error, onBuildSealed early-returns after emitting
//     EngineTemporaryErrorEvent. That early return SKIPS both
//     asyncGossip.Gossip and engine.PayloadProcessEvent — so the new
//     block is never advertised to followers and never inserted into
//     the leader's own canonical chain via engine_newPayload + FCU.
//   - onEngineTemporaryError re-arms the seal loop after a 1s backoff,
//     producing a tight failed-commit/back-off cycle. The leader's EL
//     unsafe head therefore stops advancing past at most one in-flight
//     payload from the moment the conductor was stopped.
//   - Followers were already passive; without gossip from the leader,
//     their ELs also stop. Network is fully halted.
//
// This scenario lives in its own package (per the mantle-conductor-gate
// convention) so it gets a fresh 3-conductor cluster and the failure
// state cannot pollute or be polluted by sibling scenarios.
//
// Procedure:
//
//  1. Locate the live leader.
//  2. Pick exactly one follower (deterministic: first non-leader in
//     iteration order) to kill alongside the leader.
//  3. Snapshot the unsafe-head heights of the leader's and lone
//     survivor's ELs before any kills happen.
//  4. Stop both conductors via Conductor.Stop(), which closes the
//     in-process *conductor.OpConductor service (raft + RPC).
//  5. Watch the lone surviving follower for ~15s. It must NEVER report
//     leadership: with 2 of 3 voters dead it cannot reach a quorum to
//     commit the term-bump that elections require. We poll
//     continuously rather than just sleep-then-check so a transient
//     false leadership claim — which would be a real bug — is caught
//     at the moment it happens, not masked by a quiescent end-state
//     recheck.
//  6. Final cluster-wide assertion: every still-running conductor
//     reports IsLeader == false.
//  7. Network-down assertion: re-snapshot the leader's and survivor's
//     EL heads and assert delta <= maxAllowedAdvance. Even though only
//     the leader's CONDUCTOR was killed (not its op-node or its EL),
//     the network must halt because the leader's own seal pipeline
//     blocks on a successful CommitUnsafePayload (see mechanism above).
func TestConductorNoLeaderWhenLeaderAndFollowerFail(gt *testing.T) {
	t := devtest.SerialT(gt)
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestConductorNoLeaderWhenLeaderAndFollowerFail",
	)

	sys := presets.NewMantleMinimalWithFaultyConductors(t)
	r := t.Require()

	if len(sys.SysgoConductors) < 3 {
		t.Skipf("active+follower-failure scenario needs in-process "+
			"conductors and >=3 voters; got %d sysgo conductors",
			len(sys.SysgoConductors))
	}

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; scenario needs >=3",
				chainID, len(conductors))
			continue
		}

		// Suite-wide baseline: cluster must be healthy and have a
		// leader before we start killing things.
		conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)

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
				"active+follower-failure test", chainID)
		leaderID := leaderDsl.Escape().ID()

		// 2. Pick one follower to co-kill with the leader. The other
		//    follower is the "lone survivor" we expect to be unable to
		//    win an election.
		var (
			followerToKillDsl *dsl.Conductor
			survivorDsl       *dsl.Conductor
		)
		for _, c := range conductors {
			id := c.Escape().ID()
			if id == leaderID {
				continue
			}
			if followerToKillDsl == nil {
				followerToKillDsl = c
			} else {
				survivorDsl = c
				break
			}
		}
		r.NotNil(followerToKillDsl,
			"chain %s: could not find a follower to kill alongside leader",
			chainID)
		r.NotNil(survivorDsl,
			"chain %s: could not find a lone-survivor follower", chainID)
		followerToKillID := followerToKillDsl.Escape().ID()
		survivorID := survivorDsl.Escape().ID()

		logger.Info("Active+follower-failure topology",
			"chain", chainID,
			"leaderToKill", leaderID,
			"followerToKill", followerToKillID,
			"loneSurvivor", survivorID,
		)

		// 3. Snapshot unsafe-head heights of leader & survivor BEFORE
		//    the conductor kills. We use these later to prove the
		//    network is down: with no live raft quorum, the leader's
		//    onBuildSealed → CommitUnsafePayload RPC will fail on every
		//    attempt (its own conductor is dead), so its EL must stop
		//    advancing.
		leaderEL := conductorhelpers.ELPairedWithConductor(sys.L2Chain, string(leaderID))
		r.NotNil(leaderEL,
			"chain %s: could not locate L2EL paired with leader %s",
			chainID, leaderID)
		survivorEL := conductorhelpers.ELPairedWithConductor(sys.L2Chain, string(survivorID))
		r.NotNil(survivorEL,
			"chain %s: could not locate L2EL paired with survivor %s",
			chainID, survivorID)
		leaderHeadBefore := leaderEL.BlockRefByLabel(eth.Unsafe).Number
		survivorHeadBefore := survivorEL.BlockRefByLabel(eth.Unsafe).Number
		logger.Info("EL unsafe heads before kills",
			"chain", chainID,
			"leaderHead", leaderHeadBefore,
			"survivorHead", survivorHeadBefore,
		)

		// 4. Stop both targeted conductors.
		leaderSysgo, ok := sys.SysgoConductors[leaderID]
		r.True(ok, "chain %s: leader %s missing from SysgoConductors",
			chainID, leaderID)
		followerSysgo, ok := sys.SysgoConductors[followerToKillID]
		r.True(ok, "chain %s: follower %s missing from SysgoConductors",
			chainID, followerToKillID)

		logger.Info("Stopping leader's conductor",
			"chain", chainID, "id", leaderID)
		leaderSysgo.Stop()
		logger.Info("Stopping co-killed follower's conductor",
			"chain", chainID, "id", followerToKillID)
		followerSysgo.Stop()

		// 5. Watch the lone survivor for ~15s. It must NEVER report
		//    leadership. Continuous polling catches a transient false
		//    leadership claim at the moment it would happen.
		watchUntil := time.Now().Add(15 * time.Second)
		for time.Now().Before(watchUntil) {
			r.False(survivorDsl.IsLeader(),
				"chain %s: lone-survivor follower %s falsely reported "+
					"leadership while 2 of 3 conductors are dead — this "+
					"is a quorum-violation bug",
				chainID, survivorID)
			time.Sleep(500 * time.Millisecond)
		}

		// 6. Final cluster-wide check: every still-running conductor on
		//    this chain reports IsLeader == false. Skip the two we
		//    intentionally killed (their RPC dials would fail).
		for _, c := range conductors {
			id := c.Escape().ID()
			if id == leaderID || id == followerToKillID {
				continue
			}
			sysgoC, ok := sys.SysgoConductors[id]
			if !ok || !sysgoC.IsRunning() {
				continue
			}
			r.False(c.IsLeader(),
				"chain %s: conductor %s reports leadership while a "+
					"quorum is impossible (cluster size 3, leader+1 "+
					"follower stopped) — this would be a split-brain",
				chainID, id)
		}

		logger.Info("Active+follower-failure no-leader invariant verified",
			"chain", chainID,
			"killedLeader", leaderID,
			"killedFollower", followerToKillID,
			"loneSurvivor", survivorID,
		)

		// 7. Network-down assertion. We've now waited ~15s with two
		//    conductors dead. Per the README guarantee
		//    ("Active sequencer + standby Conductor failed -> Network
		//    down (no leader)") the chain MUST NOT advance further.
		//    Mechanism: the leader's op-node still has SequencerActive,
		//    but every BuildSealEvent → onBuildSealed →
		//    CommitUnsafePayload RPC dial hits a closed local conductor
		//    and errors out, so the early-return path skips both
		//    asyncGossip.Gossip and PayloadProcessEvent. The leader's
		//    own EL therefore receives no engine_newPayload and its
		//    head freezes; followers' heads freeze with it (no gossip).
		//
		//    We allow a single-block slack: a payload could already be
		//    in-flight at the moment the conductor was stopped. Any
		//    advance beyond that proves either (a) the leader is
		//    sequencing without conductor approval — split-brain at
		//    the EL layer — or (b) some unrelated retry path is
		//    succeeding when it shouldn't.
		const maxAllowedAdvance = uint64(1)
		leaderHeadAfter := leaderEL.BlockRefByLabel(eth.Unsafe).Number
		survivorHeadAfter := survivorEL.BlockRefByLabel(eth.Unsafe).Number
		leaderDelta := leaderHeadAfter - leaderHeadBefore
		survivorDelta := survivorHeadAfter - survivorHeadBefore
		logger.Info("EL unsafe heads after kills",
			"chain", chainID,
			"leaderHeadBefore", leaderHeadBefore,
			"leaderHeadAfter", leaderHeadAfter,
			"leaderDelta", leaderDelta,
			"survivorHeadBefore", survivorHeadBefore,
			"survivorHeadAfter", survivorHeadAfter,
			"survivorDelta", survivorDelta,
		)
		r.LessOrEqualf(leaderDelta, maxAllowedAdvance,
			"chain %s: leader %s EL advanced by %d blocks during the 15s "+
				"window with 2/3 conductors dead — README invariant "+
				"\"Active sequencer + standby Conductor failed -> Network "+
				"down\" is violated. Either the leader's op-node sequenced "+
				"around its own dead conductor (split-brain at the EL "+
				"layer) or some unrelated retry path is succeeding.",
			chainID, leaderID, leaderDelta)
		r.LessOrEqualf(survivorDelta, maxAllowedAdvance,
			"chain %s: lone-survivor follower %s EL advanced by %d blocks "+
				"during the 15s window with 2/3 conductors dead — there is "+
				"no live leader to gossip new payloads, so this advance "+
				"must come from an out-of-protocol path.",
			chainID, survivorID, survivorDelta)
		logger.Info("Network-down invariant verified",
			"chain", chainID,
			"leaderDelta", leaderDelta,
			"survivorDelta", survivorDelta,
		)
	}
}
