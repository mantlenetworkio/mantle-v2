package conductor

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testreq"
	elfi "github.com/ethereum-optimism/optimism/op-service/testutils/elfaultinjector"
)

// TestConductorSplitBrainAtUnsafeHead reproduces the conductor split-brain
// at unsafe head case study (op-conductor/INTEGRATION.md):
//
// In a 3-conductor cluster, activate the Engine API fault-injection proxy
// on 2 of the 3 EL nodes (specifically, every EL EXCEPT the current
// sequencer's). Every engine_newPayloadV{3,4} for blocks at or above a
// fixed threshold returns INVALID. The expected outcome:
//
//   - All 3 op-conductor instances stay alive and continue to agree on
//     cluster membership; the raft FSM keeps committing payloads —
//     "the conductor's fsm keeps advancing".
//   - The 2 ELs with active injectors stall at threshold-1 — "underlying
//     geth stops".
//   - The sequencer's EL keeps advancing past the threshold, proving the
//     leader is still building blocks even though 2 followers' op-geth
//     refuse to apply them.
//
// This test is a forward-looking regression target. It currently SKIPs
// unless the active backend supplies BOTH:
//   - a hydrated conductor cluster with at least 3 voters, AND
//   - an Engine API fault-injection proxy on each L2 EL.
//
// Today, sysgo provides per-EL fault injectors but does not wire
// op-conductor (#16418); kurtosis/persistent backends wire conductors but
// do not yet wire the fault injector. The test activates the moment
// either gap closes.
func TestConductorSplitBrainAtUnsafeHead(gt *testing.T) {
	t := devtest.SerialT(gt)
	logger := testlog.Logger(t, log.LevelInfo).With("Test", "TestConductorSplitBrainAtUnsafeHead")

	sys := presets.NewMantleMinimalWithFaultyConductors(t)
	r := t.Require()

	if len(sys.EngineFaultInjectors) < 2 {
		t.Skipf("split-brain replay needs >= 2 EL fault injectors, got %d "+
			"(sysgo wires injectors but not conductors per #16418; "+
			"kurtosis/persistent wire conductors but not injectors)",
			len(sys.EngineFaultInjectors))
	}

	// The MantleMinimal preset selects sys.L2EL via match.WithSequencerActive,
	// so this is the EL whose op-node is currently the raft leader's
	// sequencer. We will leave its injector inactive.
	sequencerELID := sys.L2EL.Escape().ID()
	logger.Info("Identified sequencer EL (left in passthrough mode)",
		"sequencerELID", sequencerELID)

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; split-brain replay needs >= 3",
				chainID, len(conductors))
			continue
		}

		// 1. Pre-injection state: capture cluster topology, current leader,
		//    and the sequencer EL's baseline unsafe head.
		membership := conductors[0].FetchClusterMembership()
		r.GreaterOrEqual(len(membership.Servers), 3,
			"chain %s: cluster has %d servers; need >= 3",
			chainID, len(membership.Servers))

		leaderInfo := conductors[0].FetchLeader()
		r.NotNil(leaderInfo, "chain %s: no current leader", chainID)
		logger.Info("Pre-injection cluster state",
			"chain", chainID,
			"leaderID", leaderInfo.ID,
			"members", len(membership.Servers))

		// Per-conductor baseline: (isLeader, !paused as "active", healthy).
		// Sanity-check the cluster is in the expected steady state before
		// we start injecting: exactly one leader, every node un-paused,
		// every node reporting healthy. The sysgo backend uses a no-op
		// health monitor (Conductor.Start in op-devstack/sysgo/l2_conductor.go),
		// so SequencerHealthy stays true regardless of EL divergence —
		// the assertion "all healthy" is a steady-state property here,
		// not a claim about real EL health.
		baselineStatuses := snapshotConductorStatuses(conductors)
		assertExpectedSteadyState(r, chainID, "pre-injection",
			baselineStatuses, leaderInfo.ID)
		logger.Info("Pre-injection per-conductor status",
			"chain", chainID, "statuses", baselineStatuses)

		baseline := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		logger.Info("Captured baseline sequencer EL unsafe head",
			"chain", chainID, "block", baseline)

		// 2. Activate fault injectors on every EL EXCEPT the sequencer's.
		//    This sets up the canonical case-study scenario: leader builds
		//    blocks, 2 followers' op-geth reject every engine_newPayload
		//    >= rejectFromBlock with INVALID.
		const rejectFromOffset uint64 = 3
		rejectFromBlock := baseline + rejectFromOffset

		var injectedIDs []stack.L2ELNodeID
		for id, inj := range sys.EngineFaultInjectors {
			if id == sequencerELID {
				continue
			}
			inj.Activate(elfi.Rule{RejectFromBlock: rejectFromBlock})
			injectedIDs = append(injectedIDs, id)
			if len(injectedIDs) == 2 {
				break
			}
		}
		r.GreaterOrEqual(len(injectedIDs), 2,
			"chain %s: needed 2 non-sequencer ELs with injectors, got %d",
			chainID, len(injectedIDs))
		logger.Info("Activated fault injection on 2 follower ELs",
			"chain", chainID,
			"rejectFromBlock", rejectFromBlock,
			"injectedIDs", injectedIDs)

		// Always restore passthrough mode, regardless of assertion outcome.
		t.Cleanup(func() {
			for _, inj := range sys.EngineFaultInjectors {
				inj.Deactivate()
			}
		})

		// 3. Observation window: 15s at ~2s block time = ~7 blocks past
		//    rejectFromBlock, well clear of any race with the threshold.
		const observationWindow = 15 * time.Second
		time.Sleep(observationWindow)

		// 4. Raft FSM still healthy → "the conductor's fsm keeps advancing".
		//    The public conductor RPC API only exposes CommitUnsafePayload
		//    (write side); LatestUnsafePayload (read side) is a Go-only
		//    method on *conductor.OpConductor. In sysgo we hold the
		//    in-process conductor wrapper and can call it directly to
		//    prove the FSM has advanced past the EL-rejection threshold
		//    on EVERY raft member — including the followers whose op-geth
		//    has stalled. This is the direct proof that "raft is
		//    validity-blind: it commits payloads on monotonic block
		//    number alone, and the FSM applies on every member".
		membershipAfter := conductors[0].FetchClusterMembership()
		r.Equal(len(membership.Servers), len(membershipAfter.Servers),
			"chain %s: cluster membership shrank during injection (%d -> %d)",
			chainID, len(membership.Servers), len(membershipAfter.Servers))

		leaderAfter := conductors[0].FetchLeader()
		r.NotNil(leaderAfter,
			"chain %s: cluster lost the leader during injection", chainID)
		r.Equal(leaderInfo.ID, leaderAfter.ID,
			"chain %s: leadership changed during injection (%s -> %s); "+
				"split-brain replay assumes a stable leader",
			chainID, leaderInfo.ID, leaderAfter.ID)

		// Per-conductor status snapshot *during* injection. This is the
		// core observability artifact for the case study: every conductor
		// reports the SAME (leader, active, healthy) shape it had before
		// injection started, even though 2 of 3 ELs have stopped applying
		// payloads. That's exactly the "conductor doesn't notice" story
		// from op-conductor/INTEGRATION.md — the split-brain is invisible
		// at the conductor layer.
		postStatuses := snapshotConductorStatuses(conductors)
		assertExpectedSteadyState(r, chainID, "post-injection",
			postStatuses, leaderAfter.ID)
		logger.Info("Post-injection per-conductor status",
			"chain", chainID, "statuses", postStatuses)

		// Direct FSM-state assertion. Skip if the active backend doesn't
		// expose in-process conductors (kurtosis/persistent today).
		//
		// LatestUnsafePayload is leader-only: RaftConsensus.LatestUnsafePayload
		// (op-conductor/consensus/raft.go) issues raft.Barrier() to get a
		// strongly-consistent FSM read, and Barrier requires leadership.
		// That's fine for our claim: by raft's quorum invariant, the
		// leader's FSM only advances when a quorum of voters have applied
		// the log entry — so a leader-FSM head past rejectFromBlock proves
		// AT LEAST one follower's FSM has also applied that entry,
		// despite that follower's op-geth rejecting every newPayload.
		if len(sys.SysgoConductors) > 0 {
			leaderConductor, ok := sys.SysgoConductors[stack.ConductorID(leaderAfter.ID)]
			r.True(ok, "chain %s: in-process conductor for leader %s missing",
				chainID, leaderAfter.ID)
			env, perr := leaderConductor.LatestUnsafePayload(t.Ctx())
			r.NoError(perr,
				"chain %s, leader conductor %s: LatestUnsafePayload failed",
				chainID, leaderAfter.ID)
			r.NotNil(env, "chain %s, leader conductor %s: FSM has no unsafe payload",
				chainID, leaderAfter.ID)
			fsmHead := uint64(env.ExecutionPayload.BlockNumber)
			r.GreaterOrEqual(fsmHead, rejectFromBlock,
				"chain %s, leader conductor %s: FSM unsafe head %d should have "+
					"advanced past rejectFromBlock=%d (raft is validity-blind: "+
					"leader's FSM only advances on quorum-applied log entries, "+
					"so this also proves a follower's FSM applied it)",
				chainID, leaderAfter.ID, fsmHead, rejectFromBlock)
			logger.Info("Confirmed raft FSM advanced past rejectFromBlock on leader",
				"chain", chainID,
				"leaderID", leaderAfter.ID,
				"rejectFromBlock", rejectFromBlock,
				"fsmHead", fsmHead)
		}

		// 5. EL divergence → "underlying geth stops".
		//    The sequencer EL has advanced past rejectFromBlock; every
		//    EL with an active injector is below it.
		seqHead := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		r.GreaterOrEqual(seqHead, rejectFromBlock,
			"chain %s: sequencer EL should have advanced past %d (got %d)",
			chainID, rejectFromBlock, seqHead)

		var totalInjections int64
		stalledHeads := make(map[string]uint64, len(injectedIDs))
		for _, id := range injectedIDs {
			elNode := elNodeByID(t, sys.L2Chain, id)
			head := elNode.BlockRefByLabel(eth.Unsafe).Number
			stalledHeads[id.String()] = head
			r.Less(head, rejectFromBlock,
				"chain %s, EL %s: should be stuck below rejectFromBlock=%d (got %d)",
				chainID, id, rejectFromBlock, head)

			totalInjections += sys.EngineFaultInjectors[id].InjectionCount()
		}
		r.Greater(totalInjections, int64(0),
			"chain %s: at least one INVALID PayloadStatusV1 should have been "+
				"synthesized across the injected ELs", chainID)

		logger.Info("Confirmed conductor split-brain at unsafe head",
			"chain", chainID,
			"raftMembers", len(membershipAfter.Servers),
			"leaderID", leaderAfter.ID,
			"leaderEL.head", seqHead,
			"rejectFromBlock", rejectFromBlock,
			"stalledHeads", stalledHeads,
			"totalInjections", totalInjections)
	}
}

// elNodeByID looks up an L2EL by ID inside a chain and returns its dsl
// wrapper. Fails the test if no matching node is found.
func elNodeByID(t devtest.T, chain *dsl.L2Network, id stack.L2ELNodeID) *dsl.L2ELNode {
	for _, el := range chain.L2ELNodes() {
		if el.ID() == id {
			return el
		}
	}
	t.Require().Failf("L2EL not found",
		"no L2EL with ID %s on chain %s", id, chain.ChainID())
	return nil
}

// conductorStatus is a small status snapshot of a single conductor.
// "active" here means "not paused" (op-conductor's RPC API has both
// Active and Paused; we use !Paused to keep the existing dsl.FetchPaused
// helper). "healthy" is what the conductor's own SequencerHealthy reports
// — which on sysgo is wired to a no-op health monitor that never flips,
// so it stays true even under EL divergence (see Conductor.Start in
// op-devstack/sysgo/l2_conductor.go for the rationale).
type conductorStatus struct {
	ID       stack.ConductorID
	IsLeader bool
	Active   bool
	Healthy  bool
}

func snapshotConductorStatuses(conductors dsl.ConductorSet) []conductorStatus {
	out := make([]conductorStatus, 0, len(conductors))
	for _, c := range conductors {
		out = append(out, conductorStatus{
			ID:       c.Escape().ID(),
			IsLeader: c.IsLeader(),
			Active:   !c.FetchPaused(),
			Healthy:  c.FetchSequencerHealthy(),
		})
	}
	return out
}

// assertExpectedSteadyState asserts the (leader, active, healthy) shape
// expected of a freshly bootstrapped 3-node conductor cluster:
//
//   - exactly one node reports IsLeader == true, and its raft server ID
//     matches expectedLeaderID (the test's reference leader from FetchLeader);
//   - every node is active (not paused) — bootstrap conductor was Resume'd
//     after raft converged, followers were Resume'd after AddVoter
//     (see DefaultMantleConductorSystem's bootstrap Finally);
//   - every node reports SequencerHealthy. On sysgo this is trivially
//     true via the no-op health monitor; on a real backend it's a real
//     claim about op-node + op-geth liveness.
//
// This same shape must hold both before injection (steady state) AND
// during injection — the case study's punchline is that the conductor
// layer DOES NOT notice EL divergence.
func assertExpectedSteadyState(
	r *testreq.Assertions,
	chainID stack.L2NetworkID,
	phase string,
	statuses []conductorStatus,
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
