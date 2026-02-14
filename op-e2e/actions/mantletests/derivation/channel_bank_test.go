package derivation

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"

	batcherFlags "github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	upgradesHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/log"
)

// TestChannelBankArsiaSizeIncrease tests that the channel bank size limit
// increases from 100MB (Bedrock) to 1GB (Arsia) when Arsia activates.
// This test verifies:
// 1. ChannelBankSize returns correct value before and after Arsia
// 2. System continues to work normally across the fork boundary
// 3. Verifier can derive blocks both before and after Arsia activation
func TestChannelBankArsiaSizeIncrease(gt *testing.T) {
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

	// Arsia will activate at L1 timestamp = 48 (after 4 L1 blocks of 12s each)
	arsiaTimeOffset := hexutil.Uint64(48)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)

	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)
	log := testlog.Logger(t, log.LevelInfo)

	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, log)

	// Setup verifier to test derivation pipeline behavior
	_, verifier := helpers.SetupVerifier(t, sd, log, miner.L1Client(t, sd.RollupCfg),
		miner.BlobStore(), &sync.Config{})

	rollupSeqCl := sequencer.RollupClient()
	batcher := helpers.NewL2Batcher(log, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:              0,
		MaxL1TxSize:              128_000,
		BatcherKey:               dp.Secrets.Batcher,
		DataAvailabilityType:     batcherFlags.BlobsType,
		ForceSubmitSingularBatch: true,
		EnableCellProofs:         true,
	}, rollupSeqCl, miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	// PHASE 1: Before Arsia activation

	genesisTime := sd.L1Cfg.Timestamp
	t.Logf("=== PHASE 1: Before Arsia (genesis time: %d, Arsia time: %d) ===",
		genesisTime, arsiaTimeOffset)

	// Create ChainSpec to access channel bank size
	chainSpec := rollup.NewChainSpec(sd.RollupCfg)

	// Verify channel bank size is 100MB (Bedrock)
	bedrockSize := chainSpec.MaxChannelBankSize(genesisTime)
	require.Equal(t, uint64(100_000_000), bedrockSize,
		"Before Arsia, channel bank size should be 100MB (Bedrock)")
	t.Logf("Channel bank size before Arsia: %d bytes (100MB)", bedrockSize)

	// Create L2 blocks before Arsia
	miner.ActEmptyBlock(t) // L1 block 1
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	preArsiaL2Block := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
	require.Greater(t, preArsiaL2Block, uint64(0), "should have created L2 blocks before Arsia")
	t.Logf("Created %d L2 blocks before Arsia", preArsiaL2Block)

	// Submit batch before Arsia
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantle(t)
	batchTX := batcher.LastSubmitted

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTX.Hash())(t)
	miner.ActL1EndBlock(t)
	preArsiaL1Block := miner.L1Chain().CurrentBlock().Number.Uint64()
	t.Logf("Submitted batch at L1 block %d (time %d)",
		preArsiaL1Block, miner.L1Chain().CurrentBlock().Time)

	// Verifier derives blocks before Arsia
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	verifierSafeBeforeArsia := verifier.L2Safe()
	require.Equal(t, preArsiaL2Block, verifierSafeBeforeArsia.Number,
		"verifier should derive all L2 blocks before Arsia")
	t.Logf("Verifier synced to L2 block %d before Arsia", verifierSafeBeforeArsia.Number)

	// PHASE 2: Arsia activation

	t.Log("=== PHASE 2: Activating Arsia ===")

	// Calculate actual Arsia activation time (genesis + offset)
	arsiaActivationTime := genesisTime + uint64(arsiaTimeOffset)

	// CRITICAL: Push L1 time far ahead (Arsia activation + large margin)
	// This ensures all L2 blocks built in Phase 3 will have L1 origins after Arsia activation
	// and avoids batch format confusion at the fork boundary
	targetL1Time := arsiaActivationTime + 120 // 120 seconds ahead
	currentTime := miner.L1Chain().CurrentBlock().Time
	t.Logf("Current L1 time: %d, pushing to: %d (Arsia: %d)", currentTime, targetL1Time, arsiaActivationTime)

	blocksToArsia := 0
	for miner.L1Chain().CurrentBlock().Time < targetL1Time {
		miner.ActEmptyBlock(t)
		blocksToArsia++
	}
	t.Logf("Mined %d L1 blocks to reach Arsia + margin", blocksToArsia)

	arsiaBlock := miner.L1Chain().CurrentBlock()
	t.Logf("Arsia activated at L1 block %d, time %d (margin: %d seconds)",
		arsiaBlock.Number, arsiaBlock.Time, arsiaBlock.Time-arsiaActivationTime)

	// Signal to sequencer and verifier so they see the new L1 head
	sequencer.ActL1HeadSignal(t)
	verifier.ActL1HeadSignal(t)

	// Verify channel bank size increased to 1GB
	arsiaSize := chainSpec.MaxChannelBankSize(arsiaBlock.Time)
	require.Equal(t, uint64(1_000_000_000), arsiaSize,
		"After Arsia, channel bank size should be 1GB")
	t.Logf("Channel bank size after Arsia: %d bytes (1GB)", arsiaSize)

	// Verify the size actually increased
	require.Equal(t, bedrockSize*10, arsiaSize,
		"Arsia should increase channel bank size by 10x")

	// PHASE 3: After Arsia activation

	t.Log("=== PHASE 3: After Arsia ===")

	// Create more L2 blocks after Arsia
	// Now all L2 blocks will have L1 origins after Arsia activation
	sequencer.ActBuildToL1Head(t)

	postArsiaL2Block := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
	require.Greater(t, postArsiaL2Block, preArsiaL2Block,
		"should have created more L2 blocks after Arsia")
	t.Logf("Created L2 blocks up to %d after Arsia", postArsiaL2Block)

	// Submit batch after Arsia
	// Use ActL2BatchSubmitMantleAtTime with the current L1 block time
	// This ensures the batcher uses the correct format (OP Stack format after Arsia)
	// instead of using time.Now() which may not match the simulated L1 time
	currentL1Time := miner.L1Chain().CurrentBlock().Time
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantleAtTime(t, currentL1Time)
	batchTX1 := batcher.LastSubmitted

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTX1.Hash())(t)
	miner.ActL1EndBlock(t)
	postArsiaL1Block := miner.L1Chain().CurrentBlock().Number.Uint64()
	t.Logf("Submitted batch at L1 block %d (after Arsia)", postArsiaL1Block)

	// Sequencer processes its own batch
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)
	seqSafeHead := sequencer.L2Safe()
	require.Equal(t, postArsiaL2Block, seqSafeHead.Number,
		"sequencer safe head should advance after Arsia")
	t.Logf("Sequencer safe head at L2 block %d", seqSafeHead.Number)

	// Verifier derives blocks after Arsia
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	verifierSafeAfterArsia := verifier.L2Safe()
	require.Equal(t, postArsiaL2Block, verifierSafeAfterArsia.Number,
		"verifier should derive all L2 blocks after Arsia")
	t.Logf("Verifier synced to L2 block %d after Arsia", verifierSafeAfterArsia.Number)

	// Final verification

	t.Log("=== Final Verification ===")

	// Verify sequencer and verifier are in sync
	require.Equal(t, seqSafeHead.Number, verifierSafeAfterArsia.Number,
		"sequencer and verifier should have same safe head")
	require.Equal(t, seqSafeHead.Hash, verifierSafeAfterArsia.Hash,
		"sequencer and verifier should have same safe head hash")

	// Verify we processed blocks both before and after Arsia
	require.Greater(t, preArsiaL2Block, uint64(0),
		"should have processed blocks before Arsia")
	require.Greater(t, postArsiaL2Block, preArsiaL2Block,
		"should have processed more blocks after Arsia")

	t.Logf("Successfully verified channel bank size increase from 100MB to 1GB")
	t.Logf("Processed %d L2 blocks before Arsia, %d after Arsia",
		preArsiaL2Block, postArsiaL2Block-preArsiaL2Block)
	t.Logf("Verifier successfully derived blocks across Arsia fork boundary")
}

// TestMaxRLPBytesPerChannelArsiaIncrease tests that the max RLP bytes per channel limit
// increases from 10MB (Bedrock) to 100MB (Arsia/Fjord) when Arsia activates.
// This limit controls the maximum size of a single channel's decompressed RLP data,
// allowing channels to contain more L2 blocks and transactions after Arsia.
func TestMaxRLPBytesPerChannelArsiaIncrease(gt *testing.T) {
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

	// Arsia will activate at L1 timestamp = 48 (after 4 L1 blocks of 12s each)
	arsiaTimeOffset := hexutil.Uint64(48)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)

	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)
	log := testlog.Logger(t, log.LevelInfo)

	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, log)

	// Setup verifier
	_, verifier := helpers.SetupVerifier(t, sd, log, miner.L1Client(t, sd.RollupCfg),
		miner.BlobStore(), &sync.Config{})

	rollupSeqCl := sequencer.RollupClient()
	batcher := helpers.NewL2Batcher(log, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:              0,
		MaxL1TxSize:              128_000,
		BatcherKey:               dp.Secrets.Batcher,
		DataAvailabilityType:     batcherFlags.BlobsType,
		ForceSubmitSingularBatch: true,
		EnableCellProofs:         true,
	}, rollupSeqCl, miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	// PHASE 1: Before Arsia activation

	genesisTime := sd.L1Cfg.Timestamp
	t.Logf("=== PHASE 1: Before Arsia (genesis time: %d, Arsia time: %d) ===",
		genesisTime, arsiaTimeOffset)

	chainSpec := rollup.NewChainSpec(sd.RollupCfg)

	// Verify MaxRLPBytesPerChannel is 10MB (Bedrock)
	bedrockRLPLimit := chainSpec.MaxRLPBytesPerChannel(genesisTime)
	require.Equal(t, uint64(10_000_000), bedrockRLPLimit,
		"Before Arsia, MaxRLPBytesPerChannel should be 10MB (Bedrock)")
	t.Logf("MaxRLPBytesPerChannel before Arsia: %d bytes (10MB)", bedrockRLPLimit)

	// Create L2 blocks before Arsia
	miner.ActEmptyBlock(t)
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	preArsiaL2Block := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
	require.Greater(t, preArsiaL2Block, uint64(0), "should have created L2 blocks before Arsia")
	t.Logf("Created %d L2 blocks before Arsia", preArsiaL2Block)

	// Submit batch before Arsia
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
	t.Logf("Verifier synced to L2 block %d before Arsia", verifierSafeBeforeArsia.Number)

	// PHASE 2: Arsia activation

	t.Log("=== PHASE 2: Activating Arsia ===")

	arsiaActivationTime := genesisTime + uint64(arsiaTimeOffset)
	targetL1Time := arsiaActivationTime + 120
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

	// Verify MaxRLPBytesPerChannel increased to 100MB
	arsiaRLPLimit := chainSpec.MaxRLPBytesPerChannel(arsiaBlock.Time)
	require.Equal(t, uint64(100_000_000), arsiaRLPLimit,
		"After Arsia, MaxRLPBytesPerChannel should be 100MB")
	t.Logf("MaxRLPBytesPerChannel after Arsia: %d bytes (100MB)", arsiaRLPLimit)

	// Verify the size actually increased by 10x
	require.Equal(t, bedrockRLPLimit*10, arsiaRLPLimit,
		"Arsia should increase MaxRLPBytesPerChannel by 10x")

	// PHASE 3: After Arsia activation

	t.Log("=== PHASE 3: After Arsia ===")

	sequencer.ActBuildToL1Head(t)

	postArsiaL2Block := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
	require.Greater(t, postArsiaL2Block, preArsiaL2Block,
		"should have created more L2 blocks after Arsia")
	t.Logf("Created L2 blocks up to %d after Arsia", postArsiaL2Block)

	// Submit batch after Arsia
	currentL1Time := miner.L1Chain().CurrentBlock().Time
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantleAtTime(t, currentL1Time)
	batchTX1 := batcher.LastSubmitted

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTX1.Hash())(t)
	miner.ActL1EndBlock(t)

	// Sequencer processes batch
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)
	seqSafeHead := sequencer.L2Safe()
	require.Equal(t, postArsiaL2Block, seqSafeHead.Number,
		"sequencer safe head should advance after Arsia")

	// Verifier derives blocks after Arsia
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	verifierSafeAfterArsia := verifier.L2Safe()
	require.Equal(t, postArsiaL2Block, verifierSafeAfterArsia.Number,
		"verifier should derive all L2 blocks after Arsia")
	t.Logf("Verifier synced to L2 block %d after Arsia", verifierSafeAfterArsia.Number)

	// Final verification

	t.Log("=== Final Verification ===")

	require.Equal(t, seqSafeHead.Number, verifierSafeAfterArsia.Number,
		"sequencer and verifier should have same safe head")
	require.Equal(t, seqSafeHead.Hash, verifierSafeAfterArsia.Hash,
		"sequencer and verifier should have same safe head hash")

	t.Logf("Successfully verified MaxRLPBytesPerChannel increase from 10MB to 100MB")
	t.Logf("Processed %d L2 blocks before Arsia, %d after Arsia",
		preArsiaL2Block, postArsiaL2Block-preArsiaL2Block)
	t.Logf("Verifier successfully derived blocks across Arsia fork boundary")
}
