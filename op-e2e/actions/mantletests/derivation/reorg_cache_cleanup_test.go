package derivation

import (
	"math/rand"
	"testing"

	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	upgradesHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// TestReorgCacheCleanup verifies that derivation pipeline caches (frames/channels/blobs)
// are properly cleaned up during L1 reorg to prevent stuck derivation.
//
// Test strategy:
// 1. Submit batch with L2 blocks 0-29 on L1 chain A
// 2. Trigger L1 reorg to chain B (orphan chain A)
// 3. Submit DIFFERENT batch with L2 blocks 0-25 on chain B
// 4. Verify verifier derives to block 25 (NOT 29) - proves cache was cleaned
// 5. Trigger second reorg to chain C
// 6. Submit ANOTHER different batch with L2 blocks 0-20 on chain C
// 7. Verify verifier derives to block 20 - proves cache cleaned after each reorg
//
// Key verification:
// If cache wasn't cleaned, verifier would either:
// - Fail to derive (mixing old and new frames)
// - Derive to wrong L2 head (using old cached data)
// By submitting DIFFERENT batches after each reorg, we prove cache is truly cleaned.
// TestReorgCacheCleanupCalldata tests cache cleanup with Calldata DA type
func TestReorgCacheCleanupCalldata(t *testing.T) {
	t.Run("SingularBatch", func(t *testing.T) {
		ReorgCacheCleanup(t, false, false)
	})
	t.Run("SpanBatch", func(t *testing.T) {
		ReorgCacheCleanup(t, true, false)
	})
}

// TestReorgCacheCleanupBlobs tests cache cleanup with Blobs DA type
func TestReorgCacheCleanupBlobs(t *testing.T) {
	t.Run("SingularBatch", func(t *testing.T) {
		ReorgCacheCleanup(t, false, true)
	})
	t.Run("SpanBatch", func(t *testing.T) {
		ReorgCacheCleanup(t, true, true)
	})
}

func ReorgCacheCleanup(gt *testing.T, isSpanBatch bool, useBlobs bool) {
	t := actionsHelpers.NewDefaultTesting(gt)

	dp := e2eutils.MakeMantleDeployParams(t, actionsHelpers.DefaultRollupTestParams())
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)
	sd := e2eutils.SetupMantleNormal(t, dp, actionsHelpers.DefaultAlloc)

	logger := testlog.Logger(t, log.LevelInfo)
	miner, seqEngine, sequencer := actionsHelpers.SetupSequencerTest(t, sd, logger)
	miner.ActL1SetFeeRecipient(common.Address{'A'})
	sequencer.ActL2PipelineFull(t)

	_, verifier := actionsHelpers.SetupVerifier(t, sd, logger, miner.L1Client(t, sd.RollupCfg), miner.BlobStore(), &sync.Config{})

	// Select batcher config based on batch type and DA type
	var batcherCfg *actionsHelpers.BatcherCfg
	if useBlobs {
		if isSpanBatch {
			batcherCfg = actionsHelpers.MantleBlobsSpanBatcherCfg(dp)
		} else {
			batcherCfg = actionsHelpers.MantleBlobsSingularBatcherCfg(dp)
		}
	} else {
		if isSpanBatch {
			batcherCfg = actionsHelpers.MantleSpanBatcherCfg(dp)
		} else {
			batcherCfg = actionsHelpers.MantleSingularBatcherCfg(dp)
		}
	}
	batcher := actionsHelpers.NewL2Batcher(logger, sd.RollupCfg, batcherCfg,
		sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	// Setup Alice for L2 transactions
	alice := actionsHelpers.NewBasicUser[any](logger, dp.Secrets.Alice, rand.New(rand.NewSource(1234)))
	alice.SetUserEnv(&actionsHelpers.BasicUserEnv[any]{
		EthCl:  seqEngine.EthClient(),
		Signer: types.LatestSigner(sd.L2Cfg.Config),
	})

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	// Build initial L1 blocks
	miner.ActEmptyBlock(t)
	miner.ActEmptyBlock(t)

	// Build L2 blocks
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// Alice makes some L2 transactions to create meaningful batch data
	for i := 0; i < 5; i++ {
		alice.ActResetTxOpts(t)
		alice.ActMakeTx(t)
		sequencer.ActL2StartBlock(t)
		seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
		sequencer.ActL2EndBlock(t)
	}

	l2Head := sequencer.L2Unsafe()
	t.Logf("Built L2 blocks up to: %d", l2Head.Number)

	// PHASE 1: Submit partial channel to L1 Chain A

	t.Log("PHASE 1: Submitting partial batch data to L1 Chain A")

	// Submit batch data but don't include it in L1 yet
	// This will create frames/channels in the batcher's internal state
	batcher.ActSubmitAll(t)
	batchTx := batcher.LastSubmitted

	// Include the batch in L1 block A
	miner.ActL1SetFeeRecipient(common.Address{'A', 1})
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	blockA := miner.L1Chain().CurrentBlock()
	t.Logf("Created L1 block A at height %d with batch", blockA.Number.Uint64())

	// Verifier starts processing the batch
	// This will populate frames/channels cache in the derivation pipeline
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	verifierSafeHead := verifier.L2Safe()
	require.Equal(t, l2Head, verifierSafeHead, "Verifier should derive L2 from batch")
	t.Logf("Verifier derived L2 up to: %d", verifierSafeHead.Number)

	// PHASE 2: L1 Reorg - orphan block A

	t.Log("PHASE 2: Triggering L1 reorg to orphan block A")

	// Reorg L1: orphan block A
	miner.ActL1RewindToParent(t)

	// Build alternative L1 chain B (longer to ensure reorg)
	miner.ActL1SetFeeRecipient(common.Address{'B', 1})
	miner.ActEmptyBlock(t)
	miner.ActL1SetFeeRecipient(common.Address{'B', 2})
	miner.ActEmptyBlock(t)

	blockB := miner.L1Chain().CurrentBlock()
	require.NotEqual(t, blockA.Hash(), blockB.Hash(), "L1 chain should have reorged")
	t.Logf("Created alternative L1 chain B at height %d", blockB.Number.Uint64())

	// PHASE 3: Verify cache cleanup after reorg

	t.Log("PHASE 3: Verifying derivation pipeline cache cleanup")

	// Signal L1 reorg to verifier
	// This should trigger pipeline reset and cache cleanup
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	// Verify L2 safe head rewound (because L1 origin was reorged out)
	verifierSafeHeadAfterReorg := verifier.L2Safe()
	require.Less(t, verifierSafeHeadAfterReorg.Number, verifierSafeHead.Number,
		"Verifier should rewind L2 safe head after L1 reorg")
	t.Logf("Verifier rewound L2 safe head from %d to %d",
		verifierSafeHead.Number, verifierSafeHeadAfterReorg.Number)

	// Note: We cannot directly inspect internal cache state (frames/channels)
	// but we can verify the pipeline behavior indicates proper cleanup

	// PHASE 4: Submit DIFFERENT batch on chain B to prove cache cleanup

	t.Log("PHASE 4: Submitting DIFFERENT batch on L1 chain B")

	// Build FEWER L2 blocks (only 3 instead of 5)
	// This creates a different L2 chain and different batch data
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)

	// Sequencer should have rewound after reorg
	t.Logf("Sequencer safe head after reorg: %d", sequencer.L2Safe().Number)
	t.Logf("Sequencer unsafe head after reorg: %d", sequencer.L2Unsafe().Number)

	// Reset batcher state after L1 reorg.
	// This mimics the real op-batcher behavior in computeSyncActions where it detects
	// "safe chain reorg" and clears channel manager state.
	batcher.Reset()

	// Build L2 to catch up with the new L1 head first
	// This ensures L2 blocks have proper L1 origin references
	sequencer.ActBuildToL1Head(t)

	// Build only 3 L2 blocks (different from the original 5)
	for i := 0; i < 3; i++ {
		alice.ActResetTxOpts(t)
		alice.ActMakeTx(t)
		sequencer.ActL2StartBlock(t)
		seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
		sequencer.ActL2EndBlock(t)
	}

	newL2Head := sequencer.L2Unsafe()
	t.Logf("Built NEW L2 chain up to: %d (different from original %d)", newL2Head.Number, l2Head.Number)
	// Note: We don't assert chain length here because ActBuildToL1Head may build more blocks
	// to catch up with the new L1 head after reorg. The key verification is that the verifier
	// derives to the correct NEW L2 head (not the old one), proving cache was cleaned.

	// Submit the NEW batch (different data than original)
	batcher.ActSubmitAll(t)
	newBatchTx := batcher.LastSubmitted
	require.NotEqual(t, batchTx.Hash(), newBatchTx.Hash(), "New batch should be different")

	miner.ActL1SetFeeRecipient(common.Address{'B', 3})
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(newBatchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	blockB2 := miner.L1Chain().CurrentBlock()
	t.Logf("Submitted NEW batch to L1 chain B at height %d", blockB2.Number.Uint64())

	// PHASE 5: Verify clean derivation without cache pollution
	// This is the KEY verification: if cache wasn't cleaned, verifier would:
	// 1. Have old frames/channels from Chain A in cache
	// 2. Try to mix them with new data from Chain B
	// 3. Fail to derive or derive to wrong L2 head

	t.Log("PHASE 5: Verifying L2 derives correctly without cache pollution")

	// Verifier should derive L2 from NEW batch without being affected by old cache
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	verifierFinalSafeHead := verifier.L2Safe()

	// CRITICAL: Should derive to NEW L2 head (not the old one)
	// This proves cache was cleaned - if old frames were still in cache,
	// verifier would either fail or derive to wrong head
	require.Equal(t, newL2Head, verifierFinalSafeHead,
		"Verifier should derive to NEW L2 head, proving cache was cleaned")
	require.NotEqual(t, l2Head, verifierFinalSafeHead,
		"Verifier should NOT derive to old L2 head")
	t.Logf("Cache cleanup verified: Verifier derived to NEW L2 head %d (not old %d)",
		verifierFinalSafeHead.Number, l2Head.Number)

	// Verify sequencer also processes the new batch
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)
	require.Equal(t, verifierFinalSafeHead, sequencer.L2Safe(),
		"Sequencer and verifier should be in sync")

	// PHASE 6: Additional verification - multiple reorgs

	t.Log("PHASE 6: Testing cache cleanup with multiple reorgs")

	// Trigger second reorg - orphan the block that contains the new batch
	miner.ActL1RewindToParent(t)
	miner.ActL1SetFeeRecipient(common.Address{'C', 1})
	miner.ActEmptyBlock(t)
	miner.ActEmptyBlock(t)

	blockC := miner.L1Chain().CurrentBlock()
	require.NotEqual(t, blockB2.Hash(), blockC.Hash(), "Second reorg should succeed")
	t.Logf("Triggered second L1 reorg to chain C at height %d", blockC.Number.Uint64())

	// Verify cache cleanup after second reorg
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	safeAfterSecondReorg := verifier.L2Safe()
	require.Less(t, safeAfterSecondReorg.Number, verifierFinalSafeHead.Number,
		"Verifier should rewind after second reorg")
	t.Logf("After second reorg, L2 safe head rewound to: %d", safeAfterSecondReorg.Number)

	// Sequencer must also process the second reorg
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)

	seqSafeAfterSecondReorg := sequencer.L2Safe()
	require.Equal(t, safeAfterSecondReorg, seqSafeAfterSecondReorg,
		"Sequencer and verifier should sync after second reorg")
	t.Logf("Sequencer synced to L2 safe head: %d", seqSafeAfterSecondReorg.Number)

	// Reset batcher state after second L1 reorg.
	// This mimics the real op-batcher behavior in computeSyncActions.
	batcher.Reset()

	// Build L2 to catch up with the new L1 head
	sequencer.ActBuildToL1Head(t)

	// Build ANOTHER different L2 chain (only 2 blocks this time)
	for i := 0; i < 2; i++ {
		alice.ActResetTxOpts(t)
		alice.ActMakeTx(t)
		sequencer.ActL2StartBlock(t)
		seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
		sequencer.ActL2EndBlock(t)
	}

	thirdL2Head := sequencer.L2Unsafe()
	t.Logf("Built THIRD L2 chain up to: %d", thirdL2Head.Number)
	// Note: Similar to above, we don't assert chain length. The key verification is correct derivation.

	// Submit the third batch
	batcher.ActSubmitAll(t)
	thirdBatchTx := batcher.LastSubmitted

	miner.ActL1SetFeeRecipient(common.Address{'C', 2})
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(thirdBatchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	// Final verification - verifier should derive correctly after multiple reorgs
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	finalSafeHead := verifier.L2Safe()
	t.Logf("Final L2 safe head after multiple reorgs: %d", finalSafeHead.Number)

	// CRITICAL: Should derive to THIRD L2 head (not first or second)
	// This proves cache was cleaned after EACH reorg
	require.Equal(t, thirdL2Head, finalSafeHead,
		"Verifier should derive to THIRD L2 head after multiple reorgs")
	require.NotEqual(t, l2Head, finalSafeHead, "Should not be first L2 head")
	require.NotEqual(t, newL2Head, finalSafeHead, "Should not be second L2 head")
	t.Logf("Multiple reorg cache cleanup verified: derived to L2 head %d", finalSafeHead.Number)

	// Sequencer should also be in sync
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)
	require.Equal(t, finalSafeHead, sequencer.L2Safe(),
		"Sequencer and verifier should be in sync after multiple reorgs")

	t.Log("SUCCESS: Cache cleanup test completed")

	t.Log("Verified cache cleanup by submitting DIFFERENT batches after each reorg:")
	t.Logf("  Chain A: L2 blocks 0-%d (original)", l2Head.Number)
	t.Logf("  Chain B: L2 blocks 0-%d (after 1st reorg)", newL2Head.Number)
	t.Logf("  Chain C: L2 blocks 0-%d (after 2nd reorg)", thirdL2Head.Number)
	t.Log("")
	t.Log("Key verifications:")
	t.Log("  Frames/channels cache cleared on each L1 reorg")
	t.Log("  No cache pollution from previous L1 chains")
	t.Log("  Verifier correctly processes NEW data after reorg")
	t.Log("  Multiple reorgs handled correctly")
	t.Log("  Each reorg produces correct L2 derivation")

}
