package pickserver_determinism

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
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// TestPickServerDeterminismFromBLandsOnA verifies a low-level property of
// hashicorp/raft's untargeted LeadershipTransfer() under sysgo: when B is
// raft leader and op-conductor's no-target TransferLeader RPC fires (the
// same call op-conductor's action loop issues internally on
// (leader,!healthy,...) cases), leadership ALWAYS lands on A — never on C.
//
// Why this property holds — short version
//
// hashicorp/raft v1.7.3's pickServer (raft.go:2163) iterates the cluster's
// configuration in slice order, skips self and non-voters, and selects the
// non-self voter with the highest replState[id].nextIndex. The comparison
// is strict ">", so on ties the iteration-first server wins.
//
// In sysgo:
//   - cluster registration order is [A, B, C]: see
//     op-devstack/sysgo/mantle_system_conductor.go where AddVoter is
//     called for B and C (in that order) on top of A's bootstrap.
//   - in-process IPC + Go scheduler determinism mean that, even when the
//     log is advancing, AppendEntries replies tend to be processed in the
//     same iteration order as the dispatch order, so replState[A] is
//     never strictly behind replState[C] in the leader's bookkeeping.
//
// Consequence: when B is leader, pickServer iterates [A, C] and either:
//   (a) finds nextIndex_A == nextIndex_C → tie → A wins (strict ">"), or
//   (b) finds nextIndex_A > nextIndex_C → A wins.
// (c) finds nextIndex_A < nextIndex_C — would pick C, but does not occur
//     deterministically under sysgo's in-process scheduling.
//
// Why this matters in practice
//
// This test pins the determinism that makes
// active_plus_follower_opnodes_failure's organic-raft path fail to
// elect C: under sysgo, raft's natural untargeted transfer locks into
// an A↔B 2-cycle and never delivers leadership to C. Tests that rely
// on the "raft round-robins through all voters" promise from the
// op-conductor README must therefore either drive C explicitly via
// TransferLeaderToServer, OR introduce real replication-progress
// divergence (not currently feasible in sysgo).
//
// Test shape
//
// The test runs N iterations. In each iteration:
//
//  1. Drive raft leadership to B explicitly via TransferLeaderToServer
//     (we don't rely on natural election to land on B; we force it).
//  2. Wait for B to actually become leader.
//  3. Issue B's no-target TransferLeader RPC. On the wire this is
//     `transferLeader` with no args — same call op-conductor's action
//     loop uses on (leader,!healthy,...) cases. It calls hashicorp
//     raft's LeadershipTransfer() which uses pickServer.
//  4. Poll for the new leader. Assert it is A and not C.
//
// We require strict A every iteration. If the property ever breaks
// under sysgo (e.g., raft library upgrade changes pickServer or
// scheduler ordering), this test catches it loudly.
//
// 5 iterations is enough to make a single chance occurrence
// statistically implausible while keeping the test fast (each
// iteration costs ~2 leadership transfers, ~2-3 s each).
func TestPickServerDeterminismFromBLandsOnA(gt *testing.T) {
	t := devtest.SerialT(gt)
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestPickServerDeterminismFromBLandsOnA",
	)
	r := t.Require()

	sys := presets.NewMantleMinimalWithConductors(t)

	// Pre-test baseline: cluster is healthy with 3 voters.
	for chainID, conductors := range sys.ConductorSets {
		conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)
	}

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; pickServer determinism "+
				"test needs >= 3 (one for each of A, B, C)",
				chainID, len(conductors))
			continue
		}

		// Resolve membership. We rely on sysgo's cluster-formation order
		// putting [A, B, C] in the configuration's Servers slice. The
		// helper RequireHealthyConductorCluster above already asserts
		// every conductor reports the expected membership, so we can
		// trust this view.
		membership := conductors[0].FetchClusterMembership()
		r.Equal(3, len(membership.Servers),
			"chain %s: expected exactly 3 voters, got %d",
			chainID, len(membership.Servers))

		// Pin A, B, C by their canonical IDs. Under sysgo, conductor
		// IDs are literally "a", "b", "c" (op-devstack/sysgo/
		// mantle_system_conductor.go derives them from the suffix of
		// the L2CL ID).
		const (
			idA = "a"
			idB = "b"
			idC = "c"
		)
		var infoA, infoB, infoC consensus.ServerInfo
		for _, mi := range membership.Servers {
			switch mi.ID {
			case idA:
				infoA = mi
			case idB:
				infoB = mi
			case idC:
				infoC = mi
			}
		}
		r.NotEmpty(infoA.ID, "chain %s: voter A not in membership", chainID)
		r.NotEmpty(infoB.ID, "chain %s: voter B not in membership", chainID)
		r.NotEmpty(infoC.ID, "chain %s: voter C not in membership", chainID)

		// Map IDs to dsl.Conductors for issuing RPCs.
		dslByID := map[string]*dsl.Conductor{}
		for _, c := range conductors {
			id := strings.TrimPrefix(
				c.String(), stack.ConductorKind.String()+"-")
			dslByID[id] = c
		}
		r.NotNil(dslByID[idA],
			"chain %s: dsl.Conductor for A not found", chainID)
		r.NotNil(dslByID[idB],
			"chain %s: dsl.Conductor for B not found", chainID)
		r.NotNil(dslByID[idC],
			"chain %s: dsl.Conductor for C not found", chainID)

		// Helper: poll for a specific conductor to become leader.
		// We need this both to set up B-as-leader and to observe the
		// post-transfer leader after the no-target call.
		waitLeader := func(ctx context.Context, want string,
			budget time.Duration) (gotLeaderID string, ok bool) {
			deadline := time.Now().Add(budget)
			for time.Now().Before(deadline) {
				lctx, cancel := context.WithTimeout(ctx, 1*time.Second)
				info, err := conductors[0].Escape().RpcAPI().
					LeaderWithID(lctx)
				cancel()
				if err != nil || info == nil {
					time.Sleep(200 * time.Millisecond)
					continue
				}
				if info.ID == want {
					return info.ID, true
				}
				gotLeaderID = info.ID
				time.Sleep(200 * time.Millisecond)
			}
			return gotLeaderID, false
		}

		// Helper: poll for ANY change of leader away from a given ID.
		waitLeaderAwayFrom := func(ctx context.Context, away string,
			budget time.Duration) (gotLeaderID string, ok bool) {
			deadline := time.Now().Add(budget)
			for time.Now().Before(deadline) {
				lctx, cancel := context.WithTimeout(ctx, 1*time.Second)
				info, err := conductors[0].Escape().RpcAPI().
					LeaderWithID(lctx)
				cancel()
				if err == nil && info != nil && info.ID != away {
					return info.ID, true
				}
				time.Sleep(200 * time.Millisecond)
			}
			return "", false
		}

		const iterations = 5
		for i := range iterations {
			logger.Info("pickServer determinism iteration",
				"chain", chainID, "iter", i+1, "of", iterations)

			// 1. Drive leadership to B. We may already be there from a
			//    previous iteration; if so, the RPC is a fast no-op
			//    (raft accepts targeted transfers to the current
			//    leader as a no-op).
			curLeaderInfo, lerr := func() (*consensus.ServerInfo, error) {
				lctx, cancel := context.WithTimeout(t.Ctx(), 2*time.Second)
				defer cancel()
				return conductors[0].Escape().RpcAPI().LeaderWithID(lctx)
			}()
			r.NoError(lerr,
				"chain %s iter %d: failed to fetch current leader",
				chainID, i+1)
			r.NotNil(curLeaderInfo,
				"chain %s iter %d: nil leader info from RpcAPI",
				chainID, i+1)
			if curLeaderInfo.ID != idB {
				curLeaderDsl := dslByID[curLeaderInfo.ID]
				r.NotNil(curLeaderDsl,
					"chain %s iter %d: no dsl.Conductor for current "+
						"leader %s", chainID, i+1, curLeaderInfo.ID)
				curLeaderDsl.TransferLeadershipTo(infoB)
				_, ok := waitLeader(t.Ctx(), idB, 10*time.Second)
				r.True(ok,
					"chain %s iter %d: failed to drive leadership to "+
						"B within 10s after TransferLeadershipTo(B) "+
						"from %s",
					chainID, i+1, curLeaderInfo.ID)
			}

			// 2. Confirm B is leader before the no-target transfer.
			r.True(dslByID[idB].IsLeader(),
				"chain %s iter %d: B did not actually take leadership "+
					"despite waitLeader=ok",
				chainID, i+1)

			// 3. Fire the no-target TransferLeader on B. This goes
			//    through op-conductor's RPC ("transferLeader") which
			//    calls oc.cons.TransferLeader() →
			//    hashicorp raft's LeadershipTransfer() → pickServer.
			tCtx, tCancel := context.WithTimeout(
				t.Ctx(), 5*time.Second)
			err := dslByID[idB].Escape().RpcAPI().TransferLeader(tCtx)
			tCancel()
			r.NoError(err,
				"chain %s iter %d: no-target TransferLeader RPC on B "+
					"returned error",
				chainID, i+1)

			// 4. Observe the new leader. Under our sysgo scheduling the
			//    transfer typically completes in well under 1s, but
			//    raft's leadership-change goroutine plus the conductor's
			//    leadership-changed handler can take a few hundred ms;
			//    a 5s budget is comfortable.
			newLeaderID, ok := waitLeaderAwayFrom(
				t.Ctx(), idB, 5*time.Second)
			r.True(ok,
				"chain %s iter %d: leadership did not move away from "+
					"B within 5s after TransferLeader — raft did not "+
					"hand over leadership at all",
				chainID, i+1)

			// 5. The whole point: the new leader must be A, not C.
			//    pickServer with localID=B iterates [A, B, C] →
			//    skip B (self) → A wins on equal nextIndex (strict
			//    ">" tiebreak in iteration order) and also wins
			//    whenever A's replication is at least as caught up
			//    as C's (the steady-state under sysgo's deterministic
			//    in-process IPC).
			r.Equal(idA, newLeaderID,
				"chain %s iter %d: new leader after no-target "+
					"TransferLeader on B was %q; expected %q. "+
					"This breaks the pickServer determinism "+
					"property documented on this test — either "+
					"hashicorp/raft's pickServer logic changed, or "+
					"sysgo's cluster registration order is no longer "+
					"[A, B, C], or the in-process scheduler now "+
					"produces non-deterministic AppendEntries reply "+
					"interleavings that allow C's nextIndex to lead "+
					"A's. Investigate %s registration order via "+
					"FetchClusterMembership and raft.pickServer "+
					"behaviour.",
				chainID, i+1, newLeaderID, idA, chainID)

			logger.Info("Iteration confirmed pickServer→A",
				"chain", chainID,
				"iter", i+1,
				"newLeader", newLeaderID)
		}

		logger.Info("pickServer determinism property holds for chain",
			"chain", chainID,
			"iterations", iterations)
	}
}
