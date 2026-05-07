package split_brain

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	elfi "github.com/ethereum-optimism/optimism/op-service/testutils/elfaultinjector"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
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

		// Suite-wide baseline.
		conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)

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
		baselineStatuses := conductorhelpers.SnapshotConductorStatuses(conductors)
		conductorhelpers.AssertExpectedSteadyState(r, chainID, "pre-injection",
			baselineStatuses, leaderInfo.ID)
		logger.Info("Pre-injection per-conductor status",
			"chain", chainID, "statuses", baselineStatuses)

		baseline := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		logger.Info("Captured baseline sequencer EL unsafe head",
			"chain", chainID, "block", baseline)

		// 2. Activate fault injectors on every EL EXCEPT the sequencer's.
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

		postStatuses := conductorhelpers.SnapshotConductorStatuses(conductors)
		conductorhelpers.AssertExpectedSteadyState(r, chainID, "post-injection",
			postStatuses, leaderAfter.ID)
		logger.Info("Post-injection per-conductor status",
			"chain", chainID, "statuses", postStatuses)

		// Direct FSM-state assertion. Skip if the active backend doesn't
		// expose in-process conductors (kurtosis/persistent today).
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
		seqHead := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		r.GreaterOrEqual(seqHead, rejectFromBlock,
			"chain %s: sequencer EL should have advanced past %d (got %d)",
			chainID, rejectFromBlock, seqHead)

		var totalInjections int64
		stalledHeads := make(map[string]uint64, len(injectedIDs))
		for _, id := range injectedIDs {
			elNode := conductorhelpers.ELNodeByID(t, sys.L2Chain, id)
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
