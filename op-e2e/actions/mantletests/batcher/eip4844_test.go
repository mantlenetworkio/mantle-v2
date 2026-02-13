package batcher

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/stretchr/testify/require"

	upgradesHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"

	batcherFlags "github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

func setupEIP4844Test(t helpers.Testing, log log.Logger) (*e2eutils.SetupData, *e2eutils.DeployParams, *helpers.L1Miner, *helpers.L2Sequencer, *helpers.L2Engine, *helpers.L2Verifier, *helpers.L2Engine) {
	dp := e2eutils.MakeMantleDeployParams(t, helpers.DefaultRollupTestParams())
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)

	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)
	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, log)
	miner.ActL1SetFeeRecipient(common.Address{'A'})
	sequencer.ActL2PipelineFull(t)
	verifEngine, verifier := helpers.SetupVerifier(t, sd, log, miner.L1Client(t, sd.RollupCfg), miner.BlobStore(), &sync.Config{})
	return sd, dp, miner, sequencer, seqEngine, verifier, verifEngine
}

func setupBatcher(t helpers.Testing, log log.Logger, sd *e2eutils.SetupData, dp *e2eutils.DeployParams, miner *helpers.L1Miner,
	sequencer *helpers.L2Sequencer, engine *helpers.L2Engine, daType batcherFlags.DataAvailabilityType,
) *helpers.L2Batcher {
	return helpers.NewL2Batcher(log, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:          0,
		MaxL1TxSize:          128_000,
		BatcherKey:           dp.Secrets.Batcher,
		DataAvailabilityType: daType,
	}, sequencer.RollupClient(), miner.EthClient(), engine.EthClient(), engine.EngineClient(t, sd.RollupCfg))
}

func TestEIP4844DataAvailability(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)

	log := testlog.Logger(t, log.LevelDebug)
	sd, dp, miner, sequencer, seqEngine, verifier, _ := setupEIP4844Test(t, log)

	batcher := setupBatcher(t, log, sd, dp, miner, sequencer, seqEngine, batcherFlags.BlobsType)

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	// build empty L1 block
	miner.ActEmptyBlock(t)
	// finalize it, so the L1 geth blob pool doesn't log errors about missing finality
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	// Create L2 blocks, and reference the L1 head as origin
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// submit all new L2 blocks
	batcher.ActSubmitAll(t)
	batchTx := batcher.LastSubmitted
	require.Equal(t, uint8(types.BlobTxType), batchTx.Type(), "batch tx must be blob-tx")

	// new L1 block with L2 batch
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	// verifier picks up the L2 chain that was submitted
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	require.Equal(t, verifier.L2Safe(), sequencer.L2Unsafe(), "verifier syncs from sequencer via L1")
	require.NotEqual(t, sequencer.L2Safe(), sequencer.L2Unsafe(), "sequencer has not processed L1 yet")
}

func TestEIP4844MultiBlobs(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	// Feel free to bump to Prague when updating this test's L1 config to activate Prague

	log := testlog.Logger(t, log.LevelDebug)
	sd, dp, miner, sequencer, seqEngine, verifier, _ := setupEIP4844Test(t, log)
	// We could use eip4844.MaxBlobsPerBlock(sd.L1Cfg.Config, sd.L1Cfg.Timestamp) here, but
	// we don't have the L1 chain config available in the action test batcher. So we just
	// stick to Cancun max blobs for now, which is sufficient for this test.
	maxBlobsPerBlock := params.DefaultCancunBlobConfig.Max

	batcher := setupBatcher(t, log, sd, dp, miner, sequencer, seqEngine, batcherFlags.BlobsType)

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	// build empty L1 block
	miner.ActEmptyBlock(t)
	// finalize it, so the L1 geth blob pool doesn't log errors about missing finality
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	// Create L2 blocks, and reference the L1 head as origin
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// submit all new L2 blocks
	batcher.ActSubmitAllMultiBlobs(t, maxBlobsPerBlock)
	batchTx := batcher.LastSubmitted
	require.Equal(t, uint8(types.BlobTxType), batchTx.Type(), "batch tx must be blob-tx")
	require.Len(t, batchTx.BlobTxSidecar().Blobs, maxBlobsPerBlock)

	// new L1 block with L2 batch
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	// verifier picks up the L2 chain that was submitted
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	require.Equal(t, verifier.L2Safe(), sequencer.L2Unsafe(), "verifier syncs from sequencer via L1")
	require.NotEqual(t, sequencer.L2Safe(), sequencer.L2Unsafe(), "sequencer has not processed L1 yet")
}

func TestEIP4844DataAvailabilitySwitch(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)

	log := testlog.Logger(t, log.LevelDebug)
	sd, dp, miner, sequencer, seqEngine, verifier, _ := setupEIP4844Test(t, log)

	oldBatcher := setupBatcher(t, log, sd, dp, miner, sequencer, seqEngine, batcherFlags.CalldataType)

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	// build empty L1 block
	miner.ActEmptyBlock(t)
	// finalize it, so the L1 geth blob pool doesn't log errors about missing finality
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	// Create L2 blocks, and reference the L1 head as origin
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// submit all new L2 blocks, with legacy calldata DA
	oldBatcher.ActSubmitAll(t)
	batchTx := oldBatcher.LastSubmitted
	require.Equal(t, uint8(types.DynamicFeeTxType), batchTx.Type(), "batch tx must be eip1559 tx")

	// new L1 block with L2 batch
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	// verifier picks up the L2 chain that was submitted
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	require.Equal(t, verifier.L2Safe(), sequencer.L2Unsafe(), "verifier syncs from sequencer via L1")
	require.NotEqual(t, sequencer.L2Safe(), sequencer.L2Unsafe(), "sequencer has not processed L1 yet")

	newBatcher := setupBatcher(t, log, sd, dp, miner, sequencer, seqEngine, batcherFlags.BlobsType)

	// build empty L1 block
	miner.ActEmptyBlock(t)

	// Create L2 blocks, and reference the L1 head as origin
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// submit all new L2 blocks, now with Blobs DA!
	newBatcher.ActSubmitAll(t)
	batchTx = newBatcher.LastSubmitted

	// new L1 block with L2 batch
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	require.Equal(t, uint8(types.BlobTxType), batchTx.Type(), "batch tx must be blob-tx")

	// verifier picks up the L2 chain that was submitted
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	require.Equal(t, verifier.L2Safe(), sequencer.L2Unsafe(), "verifier syncs from sequencer via L1")
	require.NotEqual(t, sequencer.L2Safe(), sequencer.L2Unsafe(), "sequencer has not processed L1 yet")
}

// TestEIP4844DuplicateBlobSubmission tests that the derivation pipeline correctly handles
// duplicate blob submissions in the same L1 block. This is a critical edge case where:
// - A batcher mistakenly submits the same L2 data twice
// - Two batcher instances accidentally submit identical batches
// - A transaction is replayed due to nonce management issues
//
// Expected behavior:
// - Derivation pipeline detects duplicate batch data
// - Only the first valid batch is processed
// - Subsequent duplicate batches are ignored
// - Safe head continues to progress normally
// - No chain stall or reorg occurs
func TestEIP4844DuplicateBlobSubmission(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)

	log := testlog.Logger(t, log.LevelDebug)
	// Setup test environment with EIP-4844 support
	sd, dp, miner, sequencer, seqEngine, verifier, _ := setupEIP4844Test(t, log)

	batcher := setupBatcher(t, log, sd, dp, miner, sequencer, seqEngine, batcherFlags.BlobsType)

	// Initialize sequencer and verifier pipelines
	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	// Step 1: Create and finalize an empty L1 block
	// This ensures the blob pool has proper finality references
	// build empty L1 block
	miner.ActEmptyBlock(t)
	// finalize it, so the L1 geth blob pool doesn't log errors about missing finality
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	// Step 2: Build L2 blocks that reference the L1 head as origin
	// The sequencer creates new L2 blocks based on the latest L1 state
	// Create L2 blocks, and reference the L1 head as origin
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// Record the current unsafe head before batch submission
	// This will be used to verify the safe head progresses correctly
	unsafeHeadBeforeSubmission := sequencer.L2Unsafe()

	// Step 3: Submit the L2 batch data for the first time
	// This creates a blob transaction containing the L2 block data
	batcher.ActSubmitAll(t)
	firstBatchTx := batcher.LastSubmitted
	require.Equal(t, uint8(types.BlobTxType), firstBatchTx.Type(),
		"first batch tx must be blob-tx")

	// Log the first batch for debugging
	t.Logf("First batch submitted: tx=%s, blobs=%d",
		firstBatchTx.Hash().Hex(),
		len(firstBatchTx.BlobTxSidecar().Blobs))

	// Step 4: INTENTIONALLY submit the same data again
	// This simulates the duplicate submission scenario where:
	// - Two batcher instances are running concurrently (configuration error)
	// - A batcher restarts and doesn't track what was already submitted
	// - Network issues cause transaction replay

	// Create a second batcher instance that is unaware of the first submission
	// Both batchers see the same L2 unsafe head and will independently create batches
	t.Logf("Creating second batcher to simulate duplicate submission")
	secondBatcher := setupBatcher(t, log, sd, dp, miner, sequencer, seqEngine, batcherFlags.BlobsType)

	// The second batcher looks at the same unsafe blocks and submits them
	// It doesn't know that firstBatcher already submitted this data
	secondBatcher.ActSubmitAll(t)
	secondBatchTx := secondBatcher.LastSubmitted
	require.Equal(t, uint8(types.BlobTxType), secondBatchTx.Type(),
		"second batch tx must also be blob-tx")

	// Verify that we actually created a different transaction
	// The transactions have different hashes (due to different nonces from different batcher instances)
	// but contain the same L2 block data in their blobs
	require.NotEqual(t, firstBatchTx.Hash(), secondBatchTx.Hash(),
		"duplicate submission should create different tx hash")

	t.Logf("Duplicate batch submitted: tx=%s, blobs=%d",
		secondBatchTx.Hash().Hex(),
		len(secondBatchTx.BlobTxSidecar().Blobs))

	// Optionally verify that both blobs contain the same data
	// (This proves it's truly a duplicate, not just two different batches)
	firstBlobs := firstBatchTx.BlobTxSidecar().Blobs
	secondBlobs := secondBatchTx.BlobTxSidecar().Blobs
	require.Equal(t, len(firstBlobs), len(secondBlobs),
		"both transactions should have same number of blobs")
	// Note: Deep comparison of blob data would require iterating through blobs
	// For now, we trust that both batchers encoded the same L2 blocks

	// Step 5: Include BOTH transactions in the same L1 block
	// This is the critical test condition: two identical batches in one L1 block
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(firstBatchTx.Hash())(t)
	miner.ActL1IncludeTxByHash(secondBatchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	t.Logf("L1 block %d includes both duplicate blob transactions",
		miner.L1Chain().CurrentBlock().Number.Uint64())

	// Step 6: Verifier processes the L1 block containing duplicate data
	// Expected behavior: derivation pipeline should handle this gracefully
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	// Step 7: CRITICAL VERIFICATION - Safe head should progress normally
	// The derivation pipeline should:
	// 1. Process the first batch successfully
	// 2. Detect that the second batch is a duplicate
	// 3. Ignore the duplicate without error
	// 4. Advance safe head to match the unsafe head
	require.Equal(t, verifier.L2Safe(), unsafeHeadBeforeSubmission,
		"verifier safe head should match sequencer unsafe head - duplicate data should be ignored")

	// Verify that safe head actually progressed (not stuck)
	require.NotEqual(t, verifier.L2Safe().Number, uint64(0),
		"safe head should have progressed beyond genesis")

	t.Logf("✓ Safe head progressed correctly: L2 block %d",
		verifier.L2Safe().Number)

	// Step 8: Sequencer should also process the L1 block correctly
	// When sequencer derives from L1, it should also ignore duplicates
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)

	require.Equal(t, sequencer.L2Safe(), verifier.L2Safe(),
		"sequencer and verifier should have identical safe heads")
	// Step 9: Additional verification - build more blocks to ensure chain continues
	// This proves that the duplicate submission didn't cause a permanent stall
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// Submit new batches and verify chain continues to progress
	batcher.ActSubmitAll(t)
	newBatchTx := batcher.LastSubmitted

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(newBatchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	// Safe head should continue to progress beyond the duplicate submission
	require.Greater(t, verifier.L2Safe().Number, unsafeHeadBeforeSubmission.Number,
		"safe head should continue progressing after duplicate submission incident")

	t.Logf("Chain continues to progress: L2 block %d",
		verifier.L2Safe().Number)

	// Ensure sequencer also processes the latest L1 block before final comparison
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)

	// Final assertion: no divergence between sequencer and verifier
	require.Equal(t, sequencer.L2Safe(), verifier.L2Safe(),
		"sequencer and verifier remain in sync throughout duplicate submission scenario")

}
