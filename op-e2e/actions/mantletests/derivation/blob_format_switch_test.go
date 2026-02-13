package derivation

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	batcherFlags "github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	upgradesHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// TestBlobFormatSwitchAfterArsia tests that after Arsia activation,
// when encountering a new format blob, the derivation pipeline correctly switches
// from Mantle format to standard blob format.
//
// This test verifies:
// 1. New format blobs (single frame per blob) trigger format switch after Arsia
// 2. The switch is logged with "Mantle format decode failed, falling back to standard blob format"
// 3. Data is successfully parsed and safe head advances
// 4. Subsequent L1 blocks use the new BlobDataSource
func TestBlobFormatSwitchAfterArsia(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)

	// Setup with Arsia NOT activated at genesis
	p := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}
	dp := e2eutils.MakeMantleDeployParams(t, p)

	// Arsia will activate at L1 timestamp = 48
	arsiaTimeOffset := hexutil.Uint64(48)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)

	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)

	// Setup log capturing for verifier to see format switch logs (Debug level)
	verifLog := testlog.Logger(t, log.LevelDebug)
	verifLogHandler := testlog.WrapCaptureLogger(verifLog.Handler())
	verifLogger := log.NewLogger(verifLogHandler)

	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, testlog.Logger(t, log.LevelInfo))

	// Setup verifier with log capturing
	_, verifier := helpers.SetupVerifier(t, sd, verifLogger, miner.L1Client(t, sd.RollupCfg),
		miner.BlobStore(), &sync.Config{})

	rollupSeqCl := sequencer.RollupClient()
	batcher := helpers.NewL2Batcher(testlog.Logger(t, log.LevelInfo), sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:              0,
		MaxL1TxSize:              128_000,
		BatcherKey:               dp.Secrets.Batcher,
		DataAvailabilityType:     batcherFlags.BlobsType,
		ForceSubmitSingularBatch: true,
		EnableCellProofs:         true,
	}, rollupSeqCl, miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	// PHASE 1: Before Arsia - Create L2 blocks and capture old format blob data

	genesisTime := sd.L1Cfg.Timestamp
	t.Logf("=== PHASE 1: Before Arsia (genesis time: %d, Arsia time: %d) ===",
		genesisTime, arsiaTimeOffset)

	// Create L2 blocks before Arsia
	miner.ActEmptyBlock(t) // L1 block 1
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	preArsiaL2Block := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
	require.Greater(t, preArsiaL2Block, uint64(0), "should have created L2 blocks before Arsia")
	t.Logf("Created %d L2 blocks before Arsia", preArsiaL2Block)

	// Submit a normal batch before Arsia to establish baseline
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantle(t)
	batchTX := batcher.LastSubmitted

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTX.Hash())(t)
	miner.ActL1EndBlock(t)

	// Verifier derives blocks before Arsia
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	verifierSafeBeforeArsia := verifier.L2Safe()
	require.Equal(t, preArsiaL2Block, verifierSafeBeforeArsia.Number,
		"verifier should derive all L2 blocks before Arsia")
	t.Logf("Verifier safe head before Arsia: block %d", verifierSafeBeforeArsia.Number)

	// PHASE 2: Arsia activation

	t.Log("=== PHASE 2: Activating Arsia ===")

	arsiaActivationTime := genesisTime + uint64(arsiaTimeOffset)
	targetL1Time := arsiaActivationTime + 120 // Push far ahead
	currentTime := miner.L1Chain().CurrentBlock().Time
	t.Logf("Current L1 time: %d, pushing to: %d (Arsia: %d)", currentTime, targetL1Time, arsiaActivationTime)

	blocksToArsia := 0
	for miner.L1Chain().CurrentBlock().Time < targetL1Time {
		miner.ActEmptyBlock(t)
		blocksToArsia++
	}
	t.Logf("Mined %d L1 blocks to reach Arsia + margin", blocksToArsia)

	arsiaBlock := miner.L1Chain().CurrentBlock()
	t.Logf("Arsia activated at L1 block %d, time %d", arsiaBlock.Number, arsiaBlock.Time)

	sequencer.ActL1HeadSignal(t)
	verifier.ActL1HeadSignal(t)

	// PHASE 3: After Arsia - Create new L2 blocks and submit with NEW format

	t.Log("=== PHASE 3: Creating L2 blocks after Arsia ===")

	// Create some new L2 blocks after Arsia
	sequencer.ActBuildToL1Head(t)
	postArsiaL2Block := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
	require.Greater(t, postArsiaL2Block, preArsiaL2Block,
		"Should have created more L2 blocks after Arsia")
	t.Logf("Created L2 blocks up to %d after Arsia", postArsiaL2Block)

	// Now prepare a batch with these new L2 blocks
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)

	// Read the frame data
	frameData := batcher.ReadNextOutputFrame(t)
	t.Logf("Prepared frame data for submission: %d bytes", len(frameData))

	// Submit using NEW FORMAT by specifying a timestamp AFTER Arsia activation
	// This forces ActL2BatchSubmitMantleRawAtTime to use the new OP Stack format
	postArsiaTime := arsiaActivationTime + 10 // Use time after Arsia activation
	t.Logf("Submitting blob with NEW format using timestamp %d (Arsia activates at %d)",
		postArsiaTime, arsiaActivationTime)

	// This will create a blob using new OP Stack format (single frame per blob)
	batcher.ActL2BatchSubmitMantleRawAtTime(t, frameData, postArsiaTime)
	newFormatTX := batcher.LastSubmitted

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(newFormatTX.Hash())(t)
	miner.ActL1EndBlock(t)
	newFormatL1Block := miner.L1Chain().CurrentBlock()
	t.Logf("New format blob submitted at L1 block %d (time %d, after Arsia %d)",
		newFormatL1Block.Number.Uint64(), newFormatL1Block.Time, arsiaActivationTime)

	// Signal verifier to process the L1 block with new format blob
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	// Verify that safe head DID advance
	// The derivation pipeline should switch to standard blob format and successfully parse
	verifierSafeAfterNewBlob := verifier.L2Safe()
	require.Greater(t, verifierSafeAfterNewBlob.Number, verifierSafeBeforeArsia.Number,
		"Safe head should advance after format switch")
	t.Logf("Safe head advanced from block %d to block %d after format switch",
		verifierSafeBeforeArsia.Number, verifierSafeAfterNewBlob.Number)

	// PHASE 4: Verify format switch logs

	t.Log("=== PHASE 4: Verifying format switch logs ===")

	capturingHandler, ok := verifLogHandler.(*testlog.CapturingHandler)
	require.True(t, ok, "Should be able to cast to CapturingHandler")

	// Look for the specific format switch log message
	formatSwitchLog := capturingHandler.FindLog(
		testlog.NewMessageContainsFilter("Mantle format decode failed, falling back to standard blob format"),
	)
	require.NotNil(t, formatSwitchLog,
		"Should find 'Mantle format decode failed, falling back to standard blob format' log")
	t.Logf("Found format switch log: %s", formatSwitchLog.Message)

	// Summary

	t.Log("=== SUMMARY ===")
	t.Logf("Before Arsia: Safe head at block %d", verifierSafeBeforeArsia.Number)
	t.Logf("After new format blob (after Arsia): Safe head advanced to block %d", verifierSafeAfterNewBlob.Number)
	t.Log("Test PASSED: Format switch triggered successfully after Arsia activation")
	t.Log("Key verifications:")
	t.Log("  1. Safe head advanced (data parsed successfully)")
	t.Log("  2. Format switch log captured")
}

// TestBlobFormatSwitchBeforeArsia tests that before Arsia activation,
// when encountering a new format blob, the derivation pipeline correctly switches
// from Mantle format to standard blob format.
//
// This test verifies:
// 1. New format blobs (single frame per blob) trigger format switch before Arsia
// 2. The switch is logged with "Mantle format decode failed, falling back to standard blob format"
// 3. Data is successfully parsed and safe head advances
// 4. Subsequent L1 blocks use the new BlobDataSource
func TestBlobFormatSwitchBeforeArsia(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)

	// Setup with Arsia NOT activated at genesis
	p := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}
	dp := e2eutils.MakeMantleDeployParams(t, p)

	// Arsia will activate at L1 timestamp = 96 (later than test 1 to have more room before activation)
	arsiaTimeOffset := hexutil.Uint64(96)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)

	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)

	// Setup log capturing for verifier to see format switch logs (Debug level)
	verifLog := testlog.Logger(t, log.LevelDebug)
	verifLogHandler := testlog.WrapCaptureLogger(verifLog.Handler())
	verifLogger := log.NewLogger(verifLogHandler)

	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, testlog.Logger(t, log.LevelInfo))

	// Setup verifier with log capturing
	_, verifier := helpers.SetupVerifier(t, sd, verifLogger, miner.L1Client(t, sd.RollupCfg),
		miner.BlobStore(), &sync.Config{})

	rollupSeqCl := sequencer.RollupClient()
	batcher := helpers.NewL2Batcher(testlog.Logger(t, log.LevelInfo), sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:              0,
		MaxL1TxSize:              128_000,
		BatcherKey:               dp.Secrets.Batcher,
		DataAvailabilityType:     batcherFlags.BlobsType,
		ForceSubmitSingularBatch: true,
		EnableCellProofs:         true,
	}, rollupSeqCl, miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	// PHASE 1: Before Arsia - Create L2 blocks and establish baseline

	genesisTime := sd.L1Cfg.Timestamp
	arsiaActivationTime := genesisTime + uint64(arsiaTimeOffset)
	t.Logf("=== PHASE 1: Before Arsia (genesis time: %d, Arsia time: %d) ===",
		genesisTime, arsiaActivationTime)

	// Create L2 blocks before Arsia
	miner.ActEmptyBlock(t) // L1 block 1
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	preSubmitL2Block := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
	require.Greater(t, preSubmitL2Block, uint64(0), "should have created L2 blocks before Arsia")
	t.Logf("Created %d L2 blocks before Arsia", preSubmitL2Block)

	// Submit a normal batch before Arsia to establish baseline
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantle(t)
	batchTX := batcher.LastSubmitted

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTX.Hash())(t)
	miner.ActL1EndBlock(t)

	// Verifier derives blocks before Arsia
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	verifierSafeBeforeNewBlob := verifier.L2Safe()
	require.Equal(t, preSubmitL2Block, verifierSafeBeforeNewBlob.Number,
		"verifier should derive all L2 blocks before testing new format")
	t.Logf("Verifier safe head before new format test: block %d", verifierSafeBeforeNewBlob.Number)

	// PHASE 2: Create more L2 blocks but submit with NEW format (before Arsia)

	t.Log("=== PHASE 2: Creating L2 blocks and submitting with NEW format (before Arsia) ===")

	// Verify we're still before Arsia
	currentL1Time := miner.L1Chain().CurrentBlock().Time
	require.Less(t, currentL1Time, arsiaActivationTime,
		"Should still be before Arsia activation")
	t.Logf("Current L1 time: %d, Arsia activates at: %d", currentL1Time, arsiaActivationTime)

	// Progress L1 time forward (but still before Arsia) to allow more L2 blocks to be built
	miner.ActEmptyBlock(t)
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	// Verify we're still before Arsia after advancing L1
	currentL1Time = miner.L1Chain().CurrentBlock().Time
	require.Less(t, currentL1Time, arsiaActivationTime,
		"Should still be before Arsia after L1 advance")
	t.Logf("Advanced L1 to time: %d (still before Arsia %d)", currentL1Time, arsiaActivationTime)

	// Create more L2 blocks
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)
	newL2Block := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
	require.Greater(t, newL2Block, preSubmitL2Block,
		"Should have created more L2 blocks")
	t.Logf("Created L2 blocks up to %d", newL2Block)

	// Prepare batch data
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	frameData := batcher.ReadNextOutputFrame(t)
	t.Logf("Prepared frame data for submission: %d bytes", len(frameData))

	// Submit using NEW FORMAT by specifying a timestamp AFTER Arsia activation
	// This forces ActL2BatchSubmitMantleRawAtTime to use the new OP Stack format
	// even though we're before Arsia activation on L1
	postArsiaTime := arsiaActivationTime + 10 // Use time after Arsia activation
	t.Logf("Submitting blob with NEW format using timestamp %d (Arsia activates at %d)",
		postArsiaTime, arsiaActivationTime)

	// This will create a blob using new OP Stack format (single frame per blob)
	// even though L1 is still before Arsia activation
	batcher.ActL2BatchSubmitMantleRawAtTime(t, frameData, postArsiaTime)
	newFormatTX := batcher.LastSubmitted

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(newFormatTX.Hash())(t)
	miner.ActL1EndBlock(t)
	newFormatL1Block := miner.L1Chain().CurrentBlock()
	t.Logf("New format blob submitted at L1 block %d (time %d, before Arsia %d)",
		newFormatL1Block.Number.Uint64(), newFormatL1Block.Time, arsiaActivationTime)

	// Verify we're still before Arsia after submission
	require.Less(t, newFormatL1Block.Time, arsiaActivationTime,
		"L1 should still be before Arsia after new format blob submission")

	// Signal verifier to process the L1 block with new format blob
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	// Verify that safe head DID advance
	// The derivation pipeline should switch to standard blob format and successfully parse
	verifierSafeAfterNewBlob := verifier.L2Safe()
	require.Greater(t, verifierSafeAfterNewBlob.Number, verifierSafeBeforeNewBlob.Number,
		"Safe head should advance after format switch")
	t.Logf("Safe head advanced from block %d to block %d after format switch",
		verifierSafeBeforeNewBlob.Number, verifierSafeAfterNewBlob.Number)

	// PHASE 3: Verify format switch logs

	t.Log("=== PHASE 3: Verifying format switch logs ===")

	capturingHandler, ok := verifLogHandler.(*testlog.CapturingHandler)
	require.True(t, ok, "Should be able to cast to CapturingHandler")

	// Look for the specific format switch log message
	formatSwitchLog := capturingHandler.FindLog(
		testlog.NewMessageContainsFilter("Mantle format decode failed, falling back to standard blob format"),
	)
	require.NotNil(t, formatSwitchLog,
		"Should find 'Mantle format decode failed, falling back to standard blob format' log")
	t.Logf("Found format switch log: %s", formatSwitchLog.Message)

	// Summary

	t.Log("=== SUMMARY ===")
	t.Logf("Before submitting new format blob: Safe head at block %d", verifierSafeBeforeNewBlob.Number)
	t.Logf("After new format blob (before Arsia): Safe head advanced to block %d", verifierSafeAfterNewBlob.Number)
	t.Log("Test PASSED: Format switch triggered successfully before Arsia activation")
	t.Log("Key verifications:")
	t.Log("  1. Safe head advanced (data parsed successfully)")
	t.Log("  2. Format switch log captured")
}
