package derivation

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"

	batcherFlags "github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	upgradesHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/blobstore"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// TestMantleUsesZlibNotBrotliAcrossArsia tests that Mantle continues to use
// legacy zlib compression (not Brotli) both before and after Arsia activation.
//
// This test verifies:
// 1. Blob data submitted to L1 before Arsia uses zlib compression (starts with 0x78)
// 2. Blob data submitted to L1 after Arsia still uses zlib compression (starts with 0x78)
// 3. Verifier can successfully decompress and derive L2 blocks in both cases
// 4. Mantle does NOT switch to Brotli compression (unlike OP Stack's Fjord upgrade)
func TestMantleUsesZlibNotBrotliAcrossArsia(gt *testing.T) {
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

	// Create logger with capture to verify zlib compression is logged
	verifierLog, logHandler := testlog.CaptureLogger(t, log.LevelInfo)
	seqLog := testlog.Logger(t, log.LevelInfo)

	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, seqLog)

	// Setup verifier with capturing logger to verify zlib usage in logs
	_, verifier := helpers.SetupVerifier(t, sd, verifierLog, miner.L1Client(t, sd.RollupCfg),
		miner.BlobStore(), &sync.Config{})

	rollupSeqCl := sequencer.RollupClient()
	batcher := helpers.NewL2Batcher(seqLog, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:              0,
		MaxL1TxSize:              128_000,
		BatcherKey:               dp.Secrets.Batcher,
		DataAvailabilityType:     batcherFlags.BlobsType,
		ForceSubmitSingularBatch: true,
		EnableCellProofs:         true,
	}, rollupSeqCl, miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	// PHASE 1: Before Arsia - Verify zlib compression

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

	preArsiaL2Block := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
	require.Greater(t, preArsiaL2Block, uint64(0), "should have created L2 blocks before Arsia")
	t.Logf("Created %d L2 blocks before Arsia", preArsiaL2Block)

	// Submit batch before Arsia (will use blobs with zlib compression)
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantle(t)
	batchTX := batcher.LastSubmitted

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTX.Hash())(t)
	miner.ActL1EndBlock(t)
	preArsiaL1Block := miner.L1Chain().CurrentBlock()
	t.Logf("Submitted batch at L1 block %d (time %d, before Arsia %d)",
		preArsiaL1Block.Number.Uint64(), preArsiaL1Block.Time, arsiaActivationTime)

	// Get blob hashes from the transaction
	preArsiaBlobHashes := batchTX.BlobHashes()
	require.Greater(t, len(preArsiaBlobHashes), 0, "batch tx should have blob hashes")
	t.Logf("Batch has %d blobs", len(preArsiaBlobHashes))

	// Verify blob data uses zlib compression (before Arsia)
	for i, blobHash := range preArsiaBlobHashes {
		verifyZlibCompression(t, blobHash, miner.BlobStore(), preArsiaL1Block.Time, uint64(i), "Before Arsia")
	}

	// Verifier derives blocks before Arsia (proves zlib decompression works)
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	verifierSafeBeforeArsia := verifier.L2Safe()
	require.Equal(t, preArsiaL2Block, verifierSafeBeforeArsia.Number,
		"verifier should derive all L2 blocks before Arsia using zlib decompression")
	t.Logf("Verifier synced to L2 block %d before Arsia (zlib decompression successful)",
		verifierSafeBeforeArsia.Number)

	// Verify that logs contain "zlib" compression references (before Arsia)
	// compression_algo is a log attribute, not part of the message
	zlibLogsBeforeArsia := logHandler.FindLogs(testlog.NewAttributesFilter("compression_algo", "zlib"))
	require.Greater(t, len(zlibLogsBeforeArsia), 0,
		"verifier logs should contain compression_algo=zlib attribute before Arsia")
	t.Logf("Found %d log entries with compression_algo=zlib before Arsia", len(zlibLogsBeforeArsia))

	// PHASE 2: Arsia activation

	t.Log("=== PHASE 2: Activating Arsia ===")

	// Push L1 time past Arsia activation
	targetL1Time := arsiaActivationTime + 120
	currentTime := miner.L1Chain().CurrentBlock().Time
	t.Logf("Current L1 time: %d, pushing to: %d (Arsia: %d)",
		currentTime, targetL1Time, arsiaActivationTime)

	blocksToArsia := 0
	for miner.L1Chain().CurrentBlock().Time < targetL1Time {
		miner.ActEmptyBlock(t)
		blocksToArsia++
	}
	t.Logf("Mined %d L1 blocks to reach Arsia + margin", blocksToArsia)

	arsiaBlock := miner.L1Chain().CurrentBlock()
	t.Logf("Arsia activated at L1 block %d, time %d",
		arsiaBlock.Number, arsiaBlock.Time)

	sequencer.ActL1HeadSignal(t)
	verifier.ActL1HeadSignal(t)

	// PHASE 3: After Arsia - Verify still using zlib (not Brotli)

	t.Log("=== PHASE 3: After Arsia - Verify still using zlib ===")

	// Create more L2 blocks after Arsia
	sequencer.ActBuildToL1Head(t)

	postArsiaL2Block := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
	require.Greater(t, postArsiaL2Block, preArsiaL2Block,
		"should have created more L2 blocks after Arsia")
	t.Logf("Created L2 blocks up to %d after Arsia", postArsiaL2Block)

	// Submit batch after Arsia (should still use zlib, not Brotli)
	currentL1Time := miner.L1Chain().CurrentBlock().Time
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantleAtTime(t, currentL1Time)
	batchTX2 := batcher.LastSubmitted

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTX2.Hash())(t)
	miner.ActL1EndBlock(t)
	postArsiaL1Block := miner.L1Chain().CurrentBlock()
	t.Logf("Submitted batch at L1 block %d (time %d, after Arsia %d)",
		postArsiaL1Block.Number.Uint64(), postArsiaL1Block.Time, arsiaActivationTime)

	// Get blob hashes from the transaction after Arsia
	postArsiaBlobHashes := batchTX2.BlobHashes()
	require.Greater(t, len(postArsiaBlobHashes), 0, "batch tx should have blob hashes")
	t.Logf("Batch has %d blobs after Arsia", len(postArsiaBlobHashes))

	// CRITICAL: Verify blob data STILL uses zlib compression (after Arsia)
	// This proves Mantle does NOT switch to Brotli like OP Stack's Fjord
	for i, blobHash := range postArsiaBlobHashes {
		verifyZlibCompression(t, blobHash, miner.BlobStore(), postArsiaL1Block.Time, uint64(i), "After Arsia")
	}

	// Sequencer processes batch
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)
	seqSafeHead := sequencer.L2Safe()
	require.Equal(t, postArsiaL2Block, seqSafeHead.Number,
		"sequencer safe head should advance after Arsia")

	// Verifier derives blocks after Arsia (proves zlib decompression still works)
	logHandler.Clear() // Clear logs from before Arsia to only capture post-Arsia logs
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	verifierSafeAfterArsia := verifier.L2Safe()
	require.Equal(t, postArsiaL2Block, verifierSafeAfterArsia.Number,
		"verifier should derive all L2 blocks after Arsia using zlib decompression")
	t.Logf("Verifier synced to L2 block %d after Arsia (zlib decompression successful)",
		verifierSafeAfterArsia.Number)

	// Verify that logs STILL contain "zlib" compression references (after Arsia)
	zlibLogsAfterArsia := logHandler.FindLogs(testlog.NewAttributesFilter("compression_algo", "zlib"))
	require.Greater(t, len(zlibLogsAfterArsia), 0,
		"verifier logs should STILL contain compression_algo=zlib attribute after Arsia (not brotli)")
	t.Logf("Found %d log entries with compression_algo=zlib after Arsia", len(zlibLogsAfterArsia))

	// Final verification

	t.Log("=== Final Verification ===")

	// Verify sequencer and verifier are in sync
	require.Equal(t, seqSafeHead.Number, verifierSafeAfterArsia.Number,
		"sequencer and verifier should have same safe head")
	require.Equal(t, seqSafeHead.Hash, verifierSafeAfterArsia.Hash,
		"sequencer and verifier should have same safe head hash")

	t.Log("SUCCESS: Mantle uses zlib compression both before and after Arsia")
	t.Log("All blobs start with 0x78 (zlib magic number)")
	t.Log("Verifier successfully decompressed and derived blocks in both phases")
	t.Log("Verifier logs confirm 'compression_algo=zlib' in both phases")
	t.Log("CONFIRMED: Mantle does NOT use Brotli compression (OP Stack difference)")
	t.Logf("Processed %d L2 blocks before Arsia, %d after Arsia",
		preArsiaL2Block, postArsiaL2Block-preArsiaL2Block)
}

// verifyZlibCompression verifies that blob data uses zlib compression format
// by checking the RFC 1950 zlib header (first 2 bytes should be 0x78XX)
func verifyZlibCompression(t helpers.StatefulTesting, blobHash common.Hash, blobStore *blobstore.Store,
	l1Time uint64, blobIndex uint64, phase string) {
	// Get blob data from store using GetBlobs API
	indexedHash := eth.IndexedBlobHash{
		Index: blobIndex,
		Hash:  blobHash,
	}
	ref := eth.L1BlockRef{Time: l1Time}
	blobs, err := blobStore.GetBlobs(t.Ctx(), ref, []eth.IndexedBlobHash{indexedHash})
	require.NoError(t, err, "%s blob %d: failed to get blob from store", phase, blobIndex)
	require.Len(t, blobs, 1, "%s blob %d: expected 1 blob", phase, blobIndex)

	blob := blobs[0]
	require.NotNil(t, blob, "%s blob %d: blob data should exist", phase, blobIndex)

	// Decode blob to get the raw data
	blobData, err := blob.ToData()
	require.NoError(t, err, "%s blob %d: failed to decode blob data", phase, blobIndex)
	require.Greater(t, len(blobData), 2, "%s blob %d: blob data should have at least 2 bytes",
		phase, blobIndex)

	// Mantle uses two formats depending on Arsia activation:
	// - Before Arsia (Limb format): RLP([[version_byte][frame_binary]])
	// - After Arsia (OP Stack format): [version_byte][frame_binary]
	// Try to detect and handle both formats

	var framePayload []byte
	// Check if data starts with RLP list prefix (0xC0-0xFF indicates RLP list/array)
	if blobData[0] >= 0xC0 {
		// Likely RLP-encoded (Limb format)
		var frameDataArray []eth.Data
		err = rlp.DecodeBytes(blobData, &frameDataArray)
		if err == nil && len(frameDataArray) > 0 {
			// Successfully RLP-decoded, use first element
			framePayload = frameDataArray[0]
			t.Logf("%s blob %d: Detected Limb format (RLP-encoded), extracted %d bytes",
				phase, blobIndex, len(framePayload))
		} else {
			// RLP decode failed, treat as raw frame data
			framePayload = blobData
			t.Logf("%s blob %d: RLP decode failed, treating as raw frame data", phase, blobIndex)
		}
	} else {
		// Starts with version byte (OP Stack format)
		framePayload = blobData
		t.Logf("%s blob %d: Detected OP Stack format (raw frame data)", phase, blobIndex)
	}

	// Parse frames from the payload
	frames, err := derive.ParseFrames(framePayload)
	require.NoError(t, err, "%s blob %d: failed to parse frames", phase, blobIndex)
	require.Greater(t, len(frames), 0, "%s blob %d: expected at least one frame", phase, blobIndex)

	// Get the first frame's data (this contains the compressed channel data)
	frameData := frames[0].Data
	require.Greater(t, len(frameData), 2, "%s blob %d: frame data should have at least 2 bytes for compression header",
		phase, blobIndex)

	// Verify zlib magic number (RFC 1950)
	// First byte (CMF) should be 0x78 for Deflate with 32KB window
	require.Equal(t, byte(0x78), frameData[0],
		"%s blob %d: First byte should be 0x78 (zlib CMF - Compression Method and Flags)", phase, blobIndex)

	// Second byte (FLG) varies by compression level but has valid values
	// Common values: 0x9C (no compression), 0xDA (default), 0x5E (best compression), 0x01 (fast)
	secondByte := frameData[1]
	t.Logf("%s blob %d: Confirmed zlib compression (header: 0x%02x%02x)",
		phase, blobIndex, frameData[0], secondByte)

	// Additional validation: check that it's NOT Brotli
	// Brotli has no fixed magic number, but zlib's 0x78 start is distinctive
	// If we see 0x78, it's definitely zlib, not Brotli
	require.True(t, frameData[0] == 0x78,
		"%s blob %d: frame data should use zlib compression (0x78 prefix), not Brotli", phase, blobIndex)

	// Log first few bytes of compressed data for debugging
	headerPreview := frameData
	if len(headerPreview) > 16 {
		headerPreview = headerPreview[:16]
	}
	t.Logf("%s blob %d: First bytes of compressed data: % x", phase, blobIndex, headerPreview)
}
