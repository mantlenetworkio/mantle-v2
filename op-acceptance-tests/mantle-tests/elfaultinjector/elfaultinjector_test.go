package elfaultinjector

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	elfi "github.com/ethereum-optimism/optimism/op-service/testutils/elfaultinjector"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

// TestEngineFaultInjector_VerifierStallsOnInvalidPayload reproduces the
// wire-level outcome of the op-conductor split-brain at unsafe head case
// study (see op-conductor/INTEGRATION.md):
//
//   - The sequencer's local op-geth accepts every payload it builds.
//   - The verifier's op-geth, fronted by the Engine API fault-injection
//     proxy, rejects every engine_newPayloadV{3,4} for blocks at or
//     above a fixed threshold with INVALID.
//   - Result: the sequencer EL's unsafe head advances normally; the
//     verifier EL's unsafe head stalls at threshold-1.
//
// In a full HA cluster (which sysgo cannot reproduce today, see
// #16418), this same wire-level pattern is exactly what causes:
//   - the conductor raft FSM to advance on every node (validity-blind),
//   - while followers' local UnsafeL2 stays behind,
//   - making compareUnsafeHead refuse to fail over.
//
// This acceptance test pins down the EL-side of that scenario as a
// regression target. Future mitigations (cross-EL pre-commit validation,
// shadow EL, recovery flow) can be validated against the same setup by
// extending this test.
func TestEngineFaultInjector_VerifierStallsOnInvalidPayload(gt *testing.T) {
	t := devtest.SerialT(gt)
	logger := testlog.Logger(t, log.LevelInfo).With("Test", "TestEngineFaultInjector_VerifierStallsOnInvalidPayload")

	sys := presets.NewMantleFaultyMultiNode(t)
	r := t.Require()

	// 1. Sanity: both ELs must have an injector wired (config asserted by
	//    the preset, but verify the test setup matches expectations).
	seqInj, ok := sys.EngineFaultInjectors[sys.L2EL.ID()]
	r.True(ok, "expected fault injector for sequencer EL %s", sys.L2EL.ID())
	r.NotNil(seqInj, "sequencer fault injector is nil")

	verInj, ok := sys.EngineFaultInjectors[sys.L2ELB.ID()]
	r.True(ok, "expected fault injector for verifier EL %s", sys.L2ELB.ID())
	r.NotNil(verInj, "verifier fault injector is nil")

	// 2. Pre-condition: in pass-through mode (no Activate yet), both ELs
	//    must be advancing in lock-step. NewMantleSingleChainMultiNode
	//    already waited for L2CLB to match L2CL on LocalUnsafe; do a
	//    second confirmation directly on the EL block heads to make this
	//    test self-contained.
	delta := uint64(3)
	dsl.CheckAll(t,
		sys.L2EL.AdvancedFn(eth.Unsafe, delta),
		sys.L2ELB.AdvancedFn(eth.Unsafe, delta),
	)

	baseline := sys.L2ELB.BlockRefByLabel(eth.Unsafe).Number
	logger.Info("Captured baseline verifier EL unsafe head", "block", baseline)

	// 3. Activate the verifier-side injector to reject every newPayload
	//    for blocks >= baseline+3. The leader (sequencer) keeps building
	//    blocks; the verifier's op-geth will start refusing them.
	const rejectFromOffset uint64 = 3
	rejectFromBlock := baseline + rejectFromOffset
	verInj.Activate(elfi.Rule{RejectFromBlock: rejectFromBlock})
	logger.Info("Activated fault injection on verifier EL",
		"verifier", sys.L2ELB.ID(),
		"rejectFromBlock", rejectFromBlock)

	// 4. Wait long enough that the sequencer has clearly advanced past
	//    rejectFromBlock — at 2s block time, ~5 blocks is plenty.
	const observationWindow = 15 * time.Second
	time.Sleep(observationWindow)

	// 5. Assert the divergence:
	//    a. Sequencer EL has advanced past rejectFromBlock.
	//    b. Verifier EL is stuck at rejectFromBlock-1 (the last payload
	//       that pre-dated the rejection threshold).
	seqHead := sys.L2EL.BlockRefByLabel(eth.Unsafe)
	verHead := sys.L2ELB.BlockRefByLabel(eth.Unsafe)
	logger.Info("Heads after fault-injection window",
		"seqHead", seqHead.Number,
		"verHead", verHead.Number,
		"rejectFromBlock", rejectFromBlock)

	r.GreaterOrEqual(seqHead.Number, rejectFromBlock,
		"sequencer EL should have advanced past rejectFromBlock=%d (got %d)",
		rejectFromBlock, seqHead.Number)
	r.Less(verHead.Number, rejectFromBlock,
		"verifier EL should be stuck below rejectFromBlock=%d (got %d)",
		rejectFromBlock, verHead.Number)

	r.Greater(seqHead.Number, verHead.Number,
		"sequencer EL must be ahead of verifier EL while injector is active")
	r.Greater(verInj.InjectionCount(), int64(0),
		"verifier injector should have synthesized at least one INVALID")

	logger.Info("Confirmed split EL divergence",
		"injectionCount", verInj.InjectionCount(),
		"divergence", seqHead.Number-verHead.Number)

	// 6. Disable the rule. The verifier's op-geth should now resume
	//    accepting payloads, and op-node should drive its EL to catch up
	//    with the sequencer. We don't assert exact catch-up here, just
	//    that the verifier resumes advancing — proving the injector did
	//    not corrupt geth state, only blocked acceptance while active.
	verInj.Deactivate()
	logger.Info("Deactivated fault injection on verifier EL", "verifier", sys.L2ELB.ID())

	dsl.CheckAll(t,
		sys.L2ELB.AdvancedFn(eth.Unsafe, 1),
	)

	// 7. Final sanity: the L2CL/L2ELB sync subsystem should have
	//    re-converged on the leader's chain post-deactivation.
	dsl.CheckAll(t,
		sys.L2CLB.MatchedFn(sys.L2CL, types.LocalUnsafe, 30),
	)
}
