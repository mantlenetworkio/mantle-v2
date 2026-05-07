package replace_voter_with_spare

import (
	"context"
	"errors"
	"fmt"
	"sort"
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
	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// TestConductorReplaceVoterWithSpare is the wire-level "replacement
// smoke test" for the deployment scenario documented in
// op-conductor/README.md ("split brain during deployment where there
// will be 4 sequencers running"). It mirrors the runbook's prescribed
// ordering exactly:
//
//  1. Add new sequencer as **non-voter** first.
//  2. When new sequencer catches up to tip:
//     a. **Remove** old sequencer from server group.
//     b. **Promote** new sequencer to be voter.
//
// Why this order, and why we don't just AddVoter(new) → RemoveServer(old)
//
// The naive "promote then evict" path passes raft membership checks but
// transits a 4-voter (even-numbered) quorum state where:
//   - Quorum is 3/4 — so the cluster cannot tolerate any single failure
//     while the new voter (just added, possibly not fully caught up on
//     log apply) and the about-to-be-evicted voter (the very node we
//     want out, possibly already misbehaving) both carry voting weight.
//   - A 2-2 partition (e.g., new node deployed in a different AZ from
//     the others) leaves neither side with quorum.
//   - Between AddVoter and Resume, the new node can be elected leader
//     while its conductor is still paused — leader-but-can't-sequence,
//     a chain-stall failure mode.
//
// The runbook order side-steps all three by going 3 voters → 2 voters
// (with the new server already present as a non-voter) → 3 voters in
// a single hop. The new node is never simultaneously a voter AND
// possibly-stale; the old node never sits in the cluster as a voter
// alongside the new one.
//
// What this test proves
//
//   - AddNonvoter, RemoveServer, AddVoter all work on the live RPC
//     surface (no other test in this gate exercises any of them).
//   - The 2-voter intermediate state ({leader, kept_original} + the
//     new non-voter) is operable: the leader's chain keeps advancing
//     under a strict 2/2 quorum, proving the runbook's middle step
//     does not stall sequencing in practice.
//   - The replaced 3-voter cluster is healthy at the end: leader still
//     reports IsLeader, membership is exactly {leader, kept_original,
//     new_voter}, and the leader's EL keeps producing blocks.
//
// Procedure
//
//	a) Baseline. Cluster has 3 voters (a, b, c). Spare (d) is registered
//	   as an in-process sysgo conductor but NOT in the raft cluster.
//	b) Locate the live raft leader; capture leader/spare EL handles.
//	c) AddNonvoter(d). Wait until membership shows 3 voters {a,b,c} + 1
//	   non-voter {d}.
//	d) Wait until d's EL is within 1 block of the leader's unsafe head
//	   (the runbook's "catches up to tip" gate). On sysgo this catch-up
//	   happens via the P2P peering wired in DefaultMantleConductorSystem
//	   — d has been gossiping with a/b/c since orchestrator bootstrap,
//	   so this is normally a fast pass that just confirms the wiring.
//	e) Pick victim from {a, b, c} \ {leader}. Identify keptOriginal.
//	f) RemoveServer(victim). Wait until membership shows 2 voters
//	   {leader, keptOriginal} + 1 non-voter {d}.
//	g) Mid-window assertion: under the 2/2 quorum, the leader's EL is
//	   still producing blocks. This is the operationally critical claim:
//	   the runbook's "shrink first" pattern is only safe if the cluster
//	   can keep sequencing under the temporarily strict quorum.
//	h) Resume(d) — must precede AddVoter so we never transit a
//	   "voter-but-paused" state in which raft could elect d but d's
//	   conductor would refuse to drive its op-node.
//	i) AddVoter(d). Wait until membership shows 3 voters
//	   {leader, keptOriginal, d}, no non-voters.
//	j) Final invariants:
//	   - Leader still IsLeader == true (the swap did not destabilise
//	     leadership).
//	   - Membership is exactly the post-replacement set; victim absent.
//	   - Leader's EL keeps advancing (AssertChainAdvances). This is the
//	     only direct, falsifiable proof the replaced cluster is usable.
func TestConductorReplaceVoterWithSpare(gt *testing.T) {
	t := devtest.SerialT(gt)
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestConductorReplaceVoterWithSpare",
	)

	sys := presets.NewMantleMinimalWithSpareConductor(t)
	r := t.Require()

	// Spare must be present as a sysgo in-process conductor; otherwise
	// we are running against a backend that doesn't expose the
	// 4-conductor topology and there's nothing to test.
	spareSysgoCond, ok := sys.SysgoConductors[sys.SpareConductorID]
	if !ok {
		t.Skipf("spare conductor %s not registered (active backend "+
			"doesn't expose the 4-conductor sysgo topology)",
			sys.SpareConductorID)
	}

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; replace-voter scenario "+
				"needs >=3 voters", chainID, len(conductors))
			continue
		}

		// (a) Baseline.
		//
		// The shared baseline helper iterates over every conductor in
		// the set and asserts that every non-leader's EL has caught up
		// to the leader's head — i.e. it treats every member as a raft
		// follower. At this point the spare is intentionally NOT in
		// the raft cluster (its AddNonvoter+AddVoter happen later in
		// this test), and conductorhelpers.NodeKeyForConductor does
		// not map "d", so feeding the spare into the baseline helper
		// would trip the "no EL paired with conductor" branch and
		// stall. Pass only the cluster members (a, b, c) here, and
		// operate on the full `conductors` slice below for the
		// membership-change steps.
		clusterConductors := make(dsl.ConductorSet, 0, len(conductors))
		for _, c := range conductors {
			if c.Escape().ID() == sys.SpareConductorID {
				continue
			}
			clusterConductors = append(clusterConductors, c)
		}
		conductorhelpers.RequireHealthyConductorCluster(
			t, sys.L2Chain, chainID, clusterConductors)

		// (b) Locate the live leader and the EL handles we'll need.
		var leaderDsl *dsl.Conductor
		for _, c := range clusterConductors {
			if c.IsLeader() {
				leaderDsl = c
				break
			}
		}
		r.NotNil(leaderDsl,
			"chain %s: no leader found among in-process conductors before "+
				"replace-voter test", chainID)
		leaderID := leaderDsl.Escape().ID()

		leaderEL := conductorhelpers.ELPairedWithConductor(sys.L2Chain, string(leaderID))
		r.NotNilf(leaderEL,
			"chain %s: could not locate L2EL paired with leader %s",
			chainID, leaderID)

		// The spare's EL is provisioned by DefaultMantleConductorSystemWithSpare
		// under key "sequencer-d". The shared NodeKeyForConductor helper
		// only knows about a/b/c by design (the spare topology is a one-off
		// for this test), so we do the lookup directly.
		const spareELKey = "sequencer-d"
		var spareEL *dsl.L2ELNode
		for _, el := range sys.L2Chain.L2ELNodes() {
			if el.Escape().ID().Key() == spareELKey {
				spareEL = el
				break
			}
		}
		r.NotNilf(spareEL,
			"chain %s: could not locate spare L2EL with key %q",
			chainID, spareELKey)

		// Pre-flight: confirm the spare is NOT yet in the cluster, and
		// that the cluster has exactly 3 voters as the bootstrap left
		// it. If the topology drifted, downstream assertions would
		// fail with confusing intermediate errors.
		preMembership := leaderDsl.FetchClusterMembership()
		preVoters, preNonvoters := splitBySuffrage(preMembership)
		r.Equal(3, len(preVoters),
			"chain %s: pre-test cluster has %d voters; expected 3 (a,b,c). "+
				"membership=%+v", chainID, len(preVoters),
			preMembership.Servers)
		r.Equal(0, len(preNonvoters),
			"chain %s: pre-test cluster already has non-voters %v; "+
				"expected none. membership=%+v",
			chainID, preNonvoters, preMembership.Servers)
		r.False(membershipContains(preMembership, string(sys.SpareConductorID)),
			"chain %s: spare %s is already a cluster member before the "+
				"test starts; topology bootstrap leaked the spare into "+
				"the raft cluster", chainID, sys.SpareConductorID)

		spareAddr := spareSysgoCond.ConsensusEndpoint()
		r.NotEmpty(spareAddr,
			"chain %s: spare %s has empty consensus endpoint",
			chainID, sys.SpareConductorID)
		spareID := spareSysgoCond.ServerID()

		leaderRPC := leaderDsl.Escape().RpcAPI()

		ctx, cancel := context.WithTimeout(t.Ctx(), 90*time.Second)
		defer cancel()

		// (c) Step 1 of the runbook: AddNonvoter(d).
		//
		// Non-voter is the runbook's "catch-up step": d starts receiving
		// raft log replication immediately but does NOT count toward
		// quorum and CANNOT be elected leader. So if d is slow / not
		// fully synced, no sequencing is blocked and no spurious
		// election can land on d.
		err := retry.Do0(ctx, 30, retry.Fixed(250*time.Millisecond), func() error {
			return leaderRPC.AddServerAsNonvoter(ctx, spareID, spareAddr, 0)
		})
		r.NoErrorf(err,
			"chain %s: leader %s could not add spare %s as non-voter (addr=%s)",
			chainID, leaderID, spareID, spareAddr)
		logger.Info("Spare added as non-voter",
			"chain", chainID, "spare", spareID, "addr", spareAddr)

		// Membership now: 3 voters {a,b,c} + 1 non-voter {d}.
		r.NoError(waitForMembership(ctx, leaderDsl,
			[]string{string(leaderID)}, // must include leader as voter (others verified by counts)
			[]string{spareID},          // spare is the only non-voter
			3, 1,
			nil),
			"chain %s: cluster did not converge on 3 voters + 1 non-voter "+
				"after AddNonvoter(spare)", chainID)

		// (d) Wait for d's EL to catch up to the leader's unsafe head.
		// This is the runbook's "When new sequencer catches up to tip"
		// gate. On sysgo, d's op-node has been P2P-peered with a/b/c
		// since orchestrator bootstrap, so its EL should already be at
		// the tip; this is a fast verify that doesn't usually wait.
		err = retry.Do0(ctx, 60, retry.Fixed(500*time.Millisecond), func() error {
			leaderHead := leaderEL.BlockRefByLabel(eth.Unsafe).Number
			spareHead := spareEL.BlockRefByLabel(eth.Unsafe).Number
			// Allow the spare to be 1 block behind, matching the
			// 1-block tolerance startSequencer's compareUnsafeHead
			// uses (op-conductor/conductor/service.go:874).
			if spareHead+1 < leaderHead {
				return fmt.Errorf("spare EL head=%d, leader head=%d (gap %d > 1)",
					spareHead, leaderHead, leaderHead-spareHead)
			}
			return nil
		})
		r.NoErrorf(err,
			"chain %s: spare %s EL never caught up to within 1 block of "+
				"leader's tip; the runbook's 'catches up to tip' gate "+
				"would refuse to proceed in production",
			chainID, spareID)
		logger.Info("Spare caught up to tip",
			"chain", chainID, "spare", spareID,
			"leaderHead", leaderEL.BlockRefByLabel(eth.Unsafe).Number,
			"spareHead", spareEL.BlockRefByLabel(eth.Unsafe).Number)

		// (e) Pick a non-leader victim from the original {a, b, c}. We
		// iterate clusterConductors (the original 3) so we cannot
		// accidentally evict the spare we just added.
		var victimDsl *dsl.Conductor
		for _, c := range clusterConductors {
			if c.Escape().ID() == leaderID {
				continue
			}
			victimDsl = c
			break
		}
		r.NotNil(victimDsl,
			"chain %s: could not find a non-leader voter to evict",
			chainID)
		victimID := victimDsl.Escape().ID()

		var keptOriginalID stack.ConductorID
		for _, c := range clusterConductors {
			id := c.Escape().ID()
			if id == leaderID || id == victimID {
				continue
			}
			keptOriginalID = id
			break
		}
		r.NotEmpty(string(keptOriginalID),
			"chain %s: could not identify the kept-original follower",
			chainID)

		logger.Info("Replace-voter topology",
			"chain", chainID,
			"leader", leaderID,
			"victimToEvict", victimID,
			"keptOriginal", keptOriginalID,
			"spareToPromote", spareID,
		)

		// (f) Step 2a of the runbook: RemoveServer(victim).
		//
		// After this returns and membership converges, the cluster has
		// 2 voters {leader, keptOriginal} and 1 non-voter {d}. Quorum
		// is now 2/2 — strict, but does not require d's vote.
		err = retry.Do0(ctx, 30, retry.Fixed(250*time.Millisecond), func() error {
			return leaderRPC.RemoveServer(ctx, string(victimID), 0)
		})
		r.NoErrorf(err,
			"chain %s: leader %s could not RemoveServer(%s)",
			chainID, leaderID, victimID)

		r.NoError(waitForMembership(ctx, leaderDsl,
			[]string{string(leaderID), string(keptOriginalID)},
			[]string{spareID},
			2, 1,
			[]string{string(victimID)}),
			"chain %s: cluster did not converge on the 2-voter + "+
				"1-non-voter intermediate state after RemoveServer(victim)",
			chainID)

		// (g) Mid-window assertion. Under the 2/2 quorum the cluster
		// must still produce blocks; this is the runbook order's load-
		// bearing operational claim. We verify a small +delta over a
		// short window — long enough to rule out a frozen leader, short
		// enough to keep the test snappy.
		const midWindow = 6 * time.Second
		const midExpectedDelta uint64 = 2
		midBaseline := leaderEL.BlockRefByLabel(eth.Unsafe).Number
		time.Sleep(midWindow)
		midHead := leaderEL.BlockRefByLabel(eth.Unsafe).Number
		r.GreaterOrEqualf(midHead, midBaseline+midExpectedDelta,
			"chain %s: leader EL stalled during the {leader=%s, kept=%s} "+
				"2-voter + non-voter window (baseline=%d, head=%d, expected "+
				"+%d in %s) — the runbook's shrink-then-grow ordering "+
				"is only safe if the 2-voter intermediate keeps "+
				"sequencing, and it didn't.",
			chainID, leaderID, keptOriginalID, midBaseline, midHead,
			midExpectedDelta, midWindow)
		logger.Info("Mid-window 2-voter intermediate is operable",
			"chain", chainID,
			"baseline", midBaseline,
			"head", midHead,
			"window", midWindow)

		// (h) Resume(d) — strictly before AddVoter.
		//
		// Promoting d to voter while its conductor is still paused
		// would create a window where raft could elect d (it has the
		// most up-to-date log among the new voters) but d's conductor
		// would refuse to call StartSequencer on its op-node — a
		// chain-stall failure mode the runbook's order is designed to
		// avoid. We Resume first so d is fully usable as a voter the
		// instant it gains voting weight.
		err = retry.Do0(ctx, 20, retry.Fixed(250*time.Millisecond), func() error {
			return spareSysgoCond.Resume()
		})
		r.NoErrorf(err, "chain %s: failed to Resume spare %s",
			chainID, spareID)

		// (i) Step 2b of the runbook: AddVoter(d).
		err = retry.Do0(ctx, 30, retry.Fixed(250*time.Millisecond), func() error {
			return leaderRPC.AddServerAsVoter(ctx, spareID, spareAddr, 0)
		})
		r.NoErrorf(err,
			"chain %s: leader %s could not promote spare %s to voter",
			chainID, leaderID, spareID)
		logger.Info("Spare promoted to voter",
			"chain", chainID, "spare", spareID)

		// Membership now: 3 voters {leader, kept, d}, 0 non-voters.
		r.NoError(waitForMembership(ctx, leaderDsl,
			[]string{string(leaderID), string(keptOriginalID), spareID},
			nil,
			3, 0,
			[]string{string(victimID)}),
			"chain %s: cluster did not converge on the post-replace "+
				"3-voter set", chainID)

		// (j) Final invariants.
		// (j.1) Leader still leader (membership change did not
		// destabilise leadership).
		r.True(leaderDsl.IsLeader(),
			"chain %s: leader %s lost leadership during the replace; "+
				"the runbook's order is supposed to keep the leader "+
				"stable when a non-leader is removed and a non-voter "+
				"is promoted", chainID, leaderID)

		// (j.2) Final membership is exactly the post-replacement set.
		finalMembership := leaderDsl.FetchClusterMembership()
		finalVoters, finalNonvoters := splitBySuffrage(finalMembership)
		r.Equal(3, len(finalVoters),
			"chain %s: final cluster has %d voters; expected 3 "+
				"(membership=%+v)",
			chainID, len(finalVoters), finalMembership.Servers)
		r.Equal(0, len(finalNonvoters),
			"chain %s: final cluster still has non-voters %v; "+
				"expected 0 (membership=%+v)",
			chainID, finalNonvoters, finalMembership.Servers)
		r.True(membershipContains(finalMembership, string(leaderID)),
			"chain %s: leader %s missing from final membership %+v",
			chainID, leaderID, finalMembership.Servers)
		r.True(membershipContains(finalMembership, string(keptOriginalID)),
			"chain %s: kept original follower %s missing from final "+
				"membership %+v", chainID, keptOriginalID,
			finalMembership.Servers)
		r.True(membershipContains(finalMembership, spareID),
			"chain %s: promoted spare %s missing from final membership %+v",
			chainID, spareID, finalMembership.Servers)
		r.False(membershipContains(finalMembership, string(victimID)),
			"chain %s: evicted victim %s still present in final "+
				"membership %+v", chainID, victimID,
			finalMembership.Servers)

		// (j.3) Leader's EL keeps producing blocks.
		// AssertChainAdvances waits 12s and asserts at least +3 blocks
		// at the 2s L2 block time, with margin for the conductor
		// action-loop tick that re-drives sequencing after the topology
		// change.
		conductorhelpers.AssertChainAdvances(t, leaderEL,
			fmt.Sprintf("chain %s: post-replace leader %s",
				chainID, leaderID))

		logger.Info("Replace-voter scenario verified",
			"chain", chainID,
			"finalLeader", leaderID,
			"evicted", victimID,
			"promoted", spareID,
		)
	}
}

// membershipContains reports whether the given server ID appears in the
// cluster membership snapshot (regardless of suffrage).
func membershipContains(m *consensus.ClusterMembership, id string) bool {
	if m == nil {
		return false
	}
	for _, s := range m.Servers {
		if s.ID == id {
			return true
		}
	}
	return false
}

// splitBySuffrage returns the voter and non-voter ID lists from a
// ClusterMembership snapshot. Both lists are sorted for stable
// diagnostics in failure messages.
func splitBySuffrage(m *consensus.ClusterMembership) (voters, nonvoters []string) {
	if m == nil {
		return nil, nil
	}
	for _, s := range m.Servers {
		switch s.Suffrage {
		case consensus.Voter:
			voters = append(voters, s.ID)
		case consensus.Nonvoter:
			nonvoters = append(nonvoters, s.ID)
		}
	}
	sort.Strings(voters)
	sort.Strings(nonvoters)
	return voters, nonvoters
}

// waitForMembership polls the leader's ClusterMembership view until the
// cluster has exactly wantVoterCount voters and wantNonvoterCount
// non-voters, includes every ID in mustBeVoter as a Voter, includes
// every ID in mustBeNonvoter as a Non-voter, and excludes every ID in
// mustNotContain entirely. Returns an error describing the
// last-observed state on timeout.
//
// The voter/non-voter distinction matters for the runbook order:
// between RemoveServer(victim) and AddVoter(spare) the cluster has a
// mixed-suffrage state (2 voters + 1 non-voter) that a
// suffrage-blind Servers-count assertion would silently accept even if
// the wrong server got promoted.
func waitForMembership(
	ctx context.Context,
	leader *dsl.Conductor,
	mustBeVoter []string,
	mustBeNonvoter []string,
	wantVoterCount int,
	wantNonvoterCount int,
	mustNotContain []string,
) error {
	var lastMembership *consensus.ClusterMembership
	err := retry.Do0(ctx, 120, retry.Fixed(500*time.Millisecond), func() error {
		m := leader.FetchClusterMembership()
		lastMembership = m

		voters, nonvoters := splitBySuffrage(m)
		if len(voters) != wantVoterCount {
			return fmt.Errorf("want %d voters, got %d (%v)",
				wantVoterCount, len(voters), voters)
		}
		if len(nonvoters) != wantNonvoterCount {
			return fmt.Errorf("want %d non-voters, got %d (%v)",
				wantNonvoterCount, len(nonvoters), nonvoters)
		}
		for _, id := range mustBeVoter {
			if !containsAtSuffrage(m, id, consensus.Voter) {
				return fmt.Errorf("required voter %q not present as Voter",
					id)
			}
		}
		for _, id := range mustBeNonvoter {
			if !containsAtSuffrage(m, id, consensus.Nonvoter) {
				return fmt.Errorf("required non-voter %q not present as Nonvoter",
					id)
			}
		}
		for _, id := range mustNotContain {
			if membershipContains(m, id) {
				return fmt.Errorf("forbidden member %q still present",
					id)
			}
		}
		return nil
	})
	if err != nil {
		return errors.Join(err, fmt.Errorf("last membership=%s",
			renderMembership(lastMembership)))
	}
	return nil
}

// containsAtSuffrage reports whether id appears in m with the given
// suffrage exactly.
func containsAtSuffrage(
	m *consensus.ClusterMembership,
	id string,
	want consensus.ServerSuffrage,
) bool {
	if m == nil {
		return false
	}
	for _, s := range m.Servers {
		if s.ID == id {
			return s.Suffrage == want
		}
	}
	return false
}

// renderMembership formats a ClusterMembership snapshot for diagnostic
// output. Format: "[a:Voter b:Voter c:Voter d:Nonvoter]".
func renderMembership(m *consensus.ClusterMembership) string {
	if m == nil {
		return "<nil>"
	}
	parts := make([]string, 0, len(m.Servers))
	for _, s := range m.Servers {
		parts = append(parts, fmt.Sprintf("%s:%s", s.ID, s.Suffrage))
	}
	sort.Strings(parts)
	return "[" + strings.Join(parts, " ") + "]"
}
