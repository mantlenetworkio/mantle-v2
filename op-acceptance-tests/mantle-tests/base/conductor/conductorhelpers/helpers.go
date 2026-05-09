// Package conductorhelpers contains shared helpers used across the
// conductor acceptance tests in this directory tree.
//
// Each test under op-acceptance-tests/mantle-tests/base/conductor/<name>/
// is its own Go package with its own TestMain (and therefore its own
// freshly-bootstrapped orchestrator). Helpers that need to be reused
// across packages cannot live in *_test.go files of any one test package
// — Go does not export _test.go symbols across package boundaries — so
// they live here, in a regular (non-test) Go file.
package conductorhelpers

import (
	"context"
	"strings"
	"time"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testreq"
)

// RequireHealthyConductorCluster is the common precondition that every
// conductor acceptance test in this package tree MUST call before its
// scenario-specific setup. It is the test-suite-wide "the cluster is
// healthy" baseline.
//
// What it verifies
//
//  1. Raft has a stable leader (FetchLeader returns a non-empty ID and
//     the same ID twice in a row, so we don't catch a mid-election
//     window).
//  2. The leader's op-node reports SequencerActive=true — the leader is
//     not just claimed in raft state, it is actually sequencing.
//  3. The leader's EL is in fact producing blocks: the unsafe head
//     advances by at least `expectedAdvanceDelta` blocks within
//     `advanceWindow` (≈ block_time × delta + margin).
//  4. Every follower op-node's local unsafe head is within 1 block of
//     the leader's head — i.e. P2P gossip / FSM commits are keeping
//     followers in sync. We poll up to `followerCatchupBudget` because
//     a follower that was stopped+restarted by an earlier subtest's
//     cleanup may need extra time to resync.
//
// Why this is non-negotiable
//
// op-conductor's startSequencer (op-conductor/conductor/service.go:867)
// calls compareUnsafeHead before issuing StartSequencer to op-node. The
// PostUnsafePayload bypass at service.go:874 only handles a 1-block gap.
// If a follower's op-geth lags by more than 1 block at the moment we
// crash or transfer leadership, that follower is not a viable failover
// target — startSequencer will return ErrUnsafeHeadMismatch
// indefinitely, and the chain will stall despite raft electing a new
// leader.
//
// What this is NOT
//
// This is not a generic "wait for the cluster to settle" — it is an
// *assertion*. If the cluster cannot be brought into the documented
// healthy steady state within the budget, the test fails immediately
// here with a diagnostic that names the lagging node. That fail-fast
// posture surfaces upstream-test cleanup bugs at the point of
// pollution rather than as mystery stalls in unrelated code.
//
// Usage
//
//	for chainID, conductors := range sys.ConductorSets {
//	    conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)
//	    // ... scenario-specific setup ...
//	}
func RequireHealthyConductorCluster(
	t devtest.T,
	chain *dsl.L2Network,
	chainID stack.L2NetworkID,
	conductors dsl.ConductorSet,
) {
	r := t.Require()
	r.NotEmpty(conductors,
		"chain %s: cluster has no conductors — cannot baseline health",
		chainID)

	// 1. Stable raft leader. We require two consecutive observations of
	//    the same non-empty leader ID to avoid catching a mid-election
	//    window where conductors[0] briefly reports no leader.
	const stabilityBudget = 15 * time.Second
	stableDeadline := time.Now().Add(stabilityBudget)
	var leaderID, prevLeaderID string
	for time.Now().Before(stableDeadline) {
		li := conductors[0].FetchLeader()
		if li != nil && li.ID != "" {
			if li.ID == prevLeaderID {
				leaderID = li.ID
				break
			}
			prevLeaderID = li.ID
		} else {
			prevLeaderID = ""
		}
		time.Sleep(250 * time.Millisecond)
	}
	r.NotEmpty(leaderID,
		"chain %s: raft did not converge on a stable leader within %s",
		chainID, stabilityBudget)

	// 2. Leader's op-node is actively sequencing. SequencerActive is
	//    routed to op-node's RollupAPI; this is the only RPC that gives
	//    a direct answer about whether op-node has called its
	//    driver.StartSequencer (the production SequencerHealthMonitor's
	//    SequencerHealthy reports liveness inputs but is not itself a
	//    proof that the leader is producing blocks).
	leaderCL := CLPairedWithConductor(chain, leaderID)
	r.NotNil(leaderCL,
		"chain %s: cannot resolve L2CL paired with raft leader %s",
		chainID, leaderID)

	const activeBudget = 15 * time.Second
	activeDeadline := time.Now().Add(activeBudget)
	var leaderActive bool
	var lastActiveErr error
	for time.Now().Before(activeDeadline) {
		probeCtx, cancelProbe := context.WithTimeout(t.Ctx(), 2*time.Second)
		leaderActive, lastActiveErr = leaderCL.RollupAPI().
			SequencerActive(probeCtx)
		cancelProbe()
		if lastActiveErr == nil && leaderActive {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	r.NoError(lastActiveErr,
		"chain %s: SequencerActive RPC failed on leader CL %s",
		chainID, leaderCL.ID())
	r.True(leaderActive,
		"chain %s: leader CL %s reports SequencerActive=false at "+
			"test start within %s — the cluster did not reach the "+
			"expected sequencing steady state",
		chainID, leaderCL.ID(), activeBudget)

	// 3. Leader's EL is advancing. Just claiming SequencerActive=true is
	//    insufficient: a misconfigured op-node could accept StartSequencer
	//    and fail to build, or its op-geth could be wedged. The only
	//    falsifiable proof is fresh blocks landing on the EL.
	leaderEL := ELPairedWithConductor(chain, leaderID)
	r.NotNil(leaderEL,
		"chain %s: cannot resolve L2EL paired with raft leader %s",
		chainID, leaderID)

	const advanceWindow = 6 * time.Second
	const expectedAdvanceDelta uint64 = 2
	advanceBaseline := leaderEL.BlockRefByLabel(eth.Unsafe).Number
	time.Sleep(advanceWindow)
	advanceHead := leaderEL.BlockRefByLabel(eth.Unsafe).Number
	r.GreaterOrEqual(advanceHead, advanceBaseline+expectedAdvanceDelta,
		"chain %s: leader EL %s did not advance past baseline+%d "+
			"(baseline=%d, head=%d) within %s — the cluster is not in a "+
			"healthy 'leader sequencing' steady state at test start",
		chainID, leaderEL.Escape().ID(), expectedAdvanceDelta,
		advanceBaseline, advanceHead, advanceWindow)

	// 4. All follower op-nodes have caught up to within 1 block of the
	//    leader. The 1-block tolerance matches startSequencer's
	//    PostUnsafePayload bypass at op-conductor/conductor/service.go:874.
	//    Anything looser would let a >1-block-lagging follower through,
	//    which would then wedge the next leadership transfer.
	const followerCatchupBudget = 60 * time.Second
	catchupDeadline := time.Now().Add(followerCatchupBudget)
	var caughtUp bool
	followerHeads := map[string]uint64{}
	var lastLeaderHead uint64
	for time.Now().Before(catchupDeadline) {
		lastLeaderHead = leaderEL.BlockRefByLabel(eth.Unsafe).Number
		allCaughtUp := true
		for _, cdt := range conductors {
			cid := strings.TrimPrefix(
				cdt.String(), stack.ConductorKind.String()+"-")
			if cid == leaderID {
				continue
			}
			followerEL := ELPairedWithConductor(chain, cid)
			if followerEL == nil {
				allCaughtUp = false
				break
			}
			h := followerEL.BlockRefByLabel(eth.Unsafe).Number
			followerHeads[cid] = h
			if lastLeaderHead > h+1 {
				allCaughtUp = false
				break
			}
		}
		if allCaughtUp {
			caughtUp = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	r.True(caughtUp,
		"chain %s: not every follower op-node caught up to within 1 "+
			"block of the leader within %s (leaderHead=%d, "+
			"followerHeads=%v) — cluster is in a degraded state at test "+
			"start.",
		chainID, followerCatchupBudget, lastLeaderHead, followerHeads)
}

// ELPairedWithConductor finds the L2EL whose key (per the sysgo
// wire-up in op-devstack/sysgo/mantle_system_conductor.go) is paired
// with the given raft conductor ID. Returns nil if no match is found.
func ELPairedWithConductor(
	chain *dsl.L2Network,
	conductorID string,
) *dsl.L2ELNode {
	wantKey, ok := NodeKeyForConductor(conductorID)
	if !ok {
		return nil
	}
	for _, el := range chain.L2ELNodes() {
		if el.Escape().ID().Key() == wantKey {
			return el
		}
	}
	return nil
}

// CLPairedWithConductor mirrors ELPairedWithConductor for L2CL nodes.
// We return the raw stack.L2CLNode (not a dsl wrapper) because the
// caller only needs RollupAPI() / ID() — the dsl wrapper would add
// context plumbing we don't use.
func CLPairedWithConductor(
	chain *dsl.L2Network,
	conductorID string,
) stack.L2CLNode {
	wantKey, ok := NodeKeyForConductor(conductorID)
	if !ok {
		return nil
	}
	for _, cl := range chain.Escape().L2CLNodes() {
		if cl.ID().Key() == wantKey {
			return cl
		}
	}
	return nil
}

// NodeKeyForConductor maps a conductor ID to the expected
// stack.L2{CL,EL}NodeID.Key() of its paired sequencer node, per the
// fixed sysgo topology in NewDefaultMantleConductorSystemIDs:
//
//	conductor "a" ↔ key "sequencer"   (the bootstrap leader)
//	conductor "b" ↔ key "sequencer-b"
//	conductor "c" ↔ key "sequencer-c"
//
// Returns ok=false for unrecognised conductor IDs (e.g. on a
// non-default topology); callers should treat that as a skip / NotNil
// failure rather than crashing.
func NodeKeyForConductor(conductorID string) (string, bool) {
	switch conductorID {
	case "a":
		return "sequencer", true
	case "b":
		return "sequencer-b", true
	case "c":
		return "sequencer-c", true
	default:
		return "", false
	}
}

// AssertChainAdvances verifies the supplied L2 EL is still producing
// blocks in the test's "happy path" final state.
//
// Tests whose success criteria includes "leaving the cluster with an
// active sequencer" should call this last. Asserting that raft has a
// leader, that the conductor is unpaused, and that SequencerHealthy
// returns true is HOLLOW on its own — the production
// SequencerHealthMonitor reports liveness inputs (op-node SyncStatus +
// p2p peer count), not whether op-node actually called StartSequencer.
// A conductor that thinks it's the leader but never called
// StartSequencer (or whose op-node accepted the call but didn't
// actually build blocks) would still pass those checks. The only
// direct, falsifiable proof that the cluster ended in a usable state
// is that fresh blocks are being committed to L2.
//
// The 12s window / +3 delta is ~5 blocks at the 2s L2 block time,
// plus margin for the conductor action-loop tick that drives
// startSequencer after a leader transition.
func AssertChainAdvances(t devtest.T, activeEL *dsl.L2ELNode, label string) {
	const observationWindow = 12 * time.Second
	const expectedDelta uint64 = 3
	baseline := activeEL.BlockRefByLabel(eth.Unsafe).Number
	time.Sleep(observationWindow)
	head := activeEL.BlockRefByLabel(eth.Unsafe).Number
	t.Require().GreaterOrEqual(head, baseline+expectedDelta,
		"%s: active sequencer EL %s did not advance past baseline+%d "+
			"(baseline=%d, head=%d) — the test left the cluster in a "+
			"supposedly-healthy 'leader+sequencing' state but no new "+
			"blocks are being produced",
		label, activeEL.Escape().ID(),
		expectedDelta, baseline, head)
}

// ConductorStatus is a small status snapshot of a single conductor.
// "active" here means "not paused" (op-conductor's RPC API has both
// Active and Paused; we use !Paused to keep the existing dsl.FetchPaused
// helper). "healthy" is what the conductor's own SequencerHealthy
// reports — under the production SequencerHealthMonitor wired in
// op-devstack/sysgo/l2_conductor.go, this reflects op-node SyncStatus
// + p2p peer count + (optionally) EL P2P. Under the sysgo static mesh
// (A↔B, A↔C, B↔C only), killing 2 of 3 op-nodes leaves the survivor
// with 0 peers and trips MinPeerCount=1 → SequencerHealthy=false.
type ConductorStatus struct {
	ID       stack.ConductorID
	IsLeader bool
	Active   bool
	Healthy  bool
}

// SnapshotConductorStatuses captures (IsLeader, Active, Healthy) for
// every conductor in the set.
func SnapshotConductorStatuses(conductors dsl.ConductorSet) []ConductorStatus {
	out := make([]ConductorStatus, 0, len(conductors))
	for _, c := range conductors {
		out = append(out, ConductorStatus{
			ID:       c.Escape().ID(),
			IsLeader: c.IsLeader(),
			Active:   !c.FetchPaused(),
			Healthy:  c.FetchSequencerHealthy(),
		})
	}
	return out
}

// AssertExpectedSteadyState asserts the (leader, active, healthy) shape
// expected of a freshly bootstrapped 3-node conductor cluster:
//
//   - exactly one node reports IsLeader == true, and its raft server ID
//     matches expectedLeaderID (the test's reference leader from FetchLeader);
//   - every node is active (not paused) — bootstrap conductor was Resume'd
//     after raft converged, followers were Resume'd after AddVoter
//     (see DefaultMantleConductorSystem's bootstrap Finally);
//   - every node reports SequencerHealthy. Under the production
//     SequencerHealthMonitor wired in op-devstack/sysgo/l2_conductor.go,
//     this requires op-node SyncStatus to succeed AND p2p peer count
//     ≥ MinPeerCount. In a freshly bootstrapped cluster all three
//     op-nodes are connected via the static mesh, so all three report
//     healthy.
func AssertExpectedSteadyState(
	r *testreq.Assertions,
	chainID stack.L2NetworkID,
	phase string,
	statuses []ConductorStatus,
	expectedLeaderID string,
) {
	leaderCount := 0
	var actualLeaderID stack.ConductorID
	for _, s := range statuses {
		r.True(s.Active,
			"chain %s, %s: conductor %s should be active (not paused)",
			chainID, phase, s.ID)
		r.True(s.Healthy,
			"chain %s, %s: conductor %s should report SequencerHealthy",
			chainID, phase, s.ID)
		if s.IsLeader {
			leaderCount++
			actualLeaderID = s.ID
		}
	}
	r.Equal(1, leaderCount,
		"chain %s, %s: expected exactly 1 leader across %d conductors, got %d",
		chainID, phase, len(statuses), leaderCount)
	r.Equal(stack.ConductorID(expectedLeaderID), actualLeaderID,
		"chain %s, %s: leader from per-conductor IsLeader (%s) "+
			"disagrees with cluster-level FetchLeader (%s)",
		chainID, phase, actualLeaderID, expectedLeaderID)
}

// ELNodeByID looks up an L2EL by ID inside a chain and returns its dsl
// wrapper. Fails the test if no matching node is found.
func ELNodeByID(t devtest.T, chain *dsl.L2Network, id stack.L2ELNodeID) *dsl.L2ELNode {
	for _, el := range chain.L2ELNodes() {
		if el.ID() == id {
			return el
		}
	}
	t.Require().Failf("L2EL not found",
		"no L2EL with ID %s on chain %s", id, chain.ChainID())
	return nil
}
