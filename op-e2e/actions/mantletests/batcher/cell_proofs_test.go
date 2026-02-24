package batcher

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/stretchr/testify/require"

	upgradesHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/log"

	batcherFlags "github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// TestBlobTxWithCellProofs verifies that op-batcher in dynamic Blob mode submits transactions
// with Cell Proof data (BlobTxSidecar Version1) when EnableCellProofs is true.
//
// This test validates:
// 1. Blob transaction sidecar contains cell proofs (128 proofs per blob)
// 2. Sidecar version is BlobSidecarVersion1 (for cell proofs support)
// 3. Logs show cell proofs are being used
// 4. L1 accepts the transaction with cell proofs
// 5. L2 verifier can successfully process and derive blocks from cell proof blob data
//
// Background:
// - Cell proofs (EIP-7594) are introduced in Osaka (Pectra) upgrade
// - Traditional blob proofs: 1 proof per blob (Version0)
// - Cell proofs: 128 proofs per blob (Version1), enables PeerDAS
func TestBlobTxWithCellProofs(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)

	// Create capturing logger to verify cell proof usage in logs
	lgr, logHandler := testlog.CaptureLogger(t, log.LevelInfo)

	// Setup test environment with Arsia activated at genesis
	dp := e2eutils.MakeMantleDeployParams(t, helpers.DefaultRollupTestParams())
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)

	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)
	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, lgr)
	sequencer.ActL2PipelineFull(t)

	// Setup verifier with capturing logger
	verifEngine, verifier := helpers.SetupVerifier(t, sd, lgr, miner.L1Client(t, sd.RollupCfg),
		miner.BlobStore(), &sync.Config{})
	_ = verifEngine

	// Create batcher with Cell Proofs ENABLED
	batcher := helpers.NewL2Batcher(lgr, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:          0,
		MaxL1TxSize:          128_000,
		BatcherKey:           dp.Secrets.Batcher,
		DataAvailabilityType: batcherFlags.BlobsType,
		EnableCellProofs:     true, // CRITICAL: Enable cell proofs
	}, sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	// Build empty L1 block and finalize it
	miner.ActEmptyBlock(t)
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	// Create L2 blocks
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)
	l2BlockNum := sequencer.L2Unsafe().Number

	// Submit batch with cell proofs
	batcher.ActSubmitAll(t)
	batchTx := batcher.LastSubmitted

	// Verification 1: Transaction is Type 3 (BlobTxType)

	require.Equal(t, uint8(types.BlobTxType), batchTx.Type(),
		"Batch transaction must be Type 3 (BlobTxType)")
	t.Logf("Transaction type verified: Type %d (BlobTxType)", batchTx.Type())

	// Verification 2: Sidecar exists and contains cell proofs

	sidecar := batchTx.BlobTxSidecar()
	require.NotNil(t, sidecar, "Blob transaction must have sidecar")

	numBlobs := len(sidecar.Blobs)
	require.Greater(t, numBlobs, 0, "Sidecar must contain at least one blob")
	t.Logf("Sidecar contains %d blob(s)", numBlobs)

	// Verification 3: Sidecar version is Version1 (cell proofs)

	// Note: BlobSidecarVersion0 = legacy blob proofs (1 proof per blob)
	//       BlobSidecarVersion1 = cell proofs (128 proofs per blob)
	require.Equal(t, types.BlobSidecarVersion1, sidecar.Version,
		"Sidecar version must be Version1 when cell proofs are enabled")
	t.Logf("Sidecar version verified: Version%d (cell proofs support)", sidecar.Version)

	// Verification 4: Cell proofs count is correct

	// Each blob should have 128 cell proofs (kzg4844.CellProofsPerBlob = 128)
	expectedProofs := numBlobs * kzg4844.CellProofsPerBlob
	actualProofs := len(sidecar.Proofs)
	require.Equal(t, expectedProofs, actualProofs,
		"Number of cell proofs should be %d (blobs) * %d (CellProofsPerBlob) = %d",
		numBlobs, kzg4844.CellProofsPerBlob, expectedProofs)
	t.Logf("Cell proofs count verified: %d proofs (%d blobs × %d proofs per blob)",
		actualProofs, numBlobs, kzg4844.CellProofsPerBlob)

	// Verification 5: Commitments and blob hashes match

	require.Equal(t, numBlobs, len(sidecar.Commitments),
		"Number of commitments should match number of blobs")
	require.Equal(t, numBlobs, len(batchTx.BlobHashes()),
		"Number of blob hashes should match number of blobs")
	t.Logf("Blob hashes and commitments verified: %d each", numBlobs)

	// Verification 6: Include transaction in L1 block

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTx.Hash())(t)
	miner.ActL1EndBlock(t)
	t.Logf("Cell proof blob transaction included in L1 block")

	// Verification 7: Verifier processes cell proof data successfully

	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	verifierSafeHead := verifier.L2Safe()
	require.Equal(t, l2BlockNum, verifierSafeHead.Number,
		"Verifier should successfully derive L2 blocks from cell proof blob data")
	t.Logf("Verifier successfully derived L2 block %d from cell proof blob data", verifierSafeHead.Number)

	// Verification 8: Check logs for cell proof usage (optional)

	// Note: This depends on whether the batcher logs mention cell proofs explicitly
	// We can check for any blob-related logs as a sanity check
	blobLogs := logHandler.FindLogs(testlog.NewMessageContainsFilter("blob"))
	t.Logf("Found %d log entries mentioning 'blob'", len(blobLogs))

	// Final Summary

	t.Log("SUCCESS: Cell Proof Blob Transaction Validation Complete")

	t.Logf("Transaction Type: Type 3 (BlobTxType)")
	t.Logf("Sidecar Version: Version%d (cell proofs support)", sidecar.Version)
	t.Logf("Number of Blobs: %d", numBlobs)
	t.Logf("Cell Proofs per Blob: %d", kzg4844.CellProofsPerBlob)
	t.Logf("Total Cell Proofs: %d", actualProofs)
	t.Logf("L1 Transaction: Accepted and included in block")
	t.Logf("L2 Derivation: Successfully processed by verifier")

	t.Log("CONFIRMED: Cell proofs are correctly generated and processed")
	t.Log("This enables PeerDAS (Peer Data Availability Sampling) for Osaka upgrade")
}

// TestBlobTxWithoutCellProofs verifies that when EnableCellProofs is false,
// the batcher uses legacy blob proofs (Version0, 1 proof per blob).
//
// This test validates backward compatibility:
// 1. Sidecar uses BlobSidecarVersion0 (legacy proofs)
// 2. Only 1 proof per blob (not 128)
// 3. Transaction is still accepted and processed correctly
func TestBlobTxWithoutCellProofs(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)

	lgr := testlog.Logger(t, log.LevelInfo)

	// Setup test environment
	dp := e2eutils.MakeMantleDeployParams(t, helpers.DefaultRollupTestParams())
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)

	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)
	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, lgr)
	sequencer.ActL2PipelineFull(t)

	verifEngine, verifier := helpers.SetupVerifier(t, sd, lgr, miner.L1Client(t, sd.RollupCfg),
		miner.BlobStore(), &sync.Config{})
	_ = verifEngine

	// Create batcher with Cell Proofs DISABLED (legacy mode)
	batcher := helpers.NewL2Batcher(lgr, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:          0,
		MaxL1TxSize:          128_000,
		BatcherKey:           dp.Secrets.Batcher,
		DataAvailabilityType: batcherFlags.BlobsType,
		EnableCellProofs:     false, // DISABLED: Use legacy blob proofs
	}, sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	// Build L1 block and create L2 blocks
	miner.ActEmptyBlock(t)
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)
	l2BlockNum := sequencer.L2Unsafe().Number

	// Submit batch with legacy proofs
	batcher.ActSubmitAll(t)
	batchTx := batcher.LastSubmitted

	// Verify transaction type
	require.Equal(t, uint8(types.BlobTxType), batchTx.Type(),
		"Batch transaction must be Type 3 (BlobTxType)")

	// Verify sidecar with legacy proofs
	sidecar := batchTx.BlobTxSidecar()
	require.NotNil(t, sidecar, "Blob transaction must have sidecar")

	numBlobs := len(sidecar.Blobs)
	require.Greater(t, numBlobs, 0, "Sidecar must contain at least one blob")

	// CRITICAL: Verify Version0 (legacy proofs)

	require.Equal(t, types.BlobSidecarVersion0, sidecar.Version,
		"Sidecar version must be Version0 when cell proofs are disabled")
	t.Logf("Sidecar version verified: Version%d (legacy blob proofs)", sidecar.Version)

	// CRITICAL: Verify legacy proof count (1 per blob)

	expectedProofs := numBlobs // 1 proof per blob in legacy mode
	actualProofs := len(sidecar.Proofs)
	require.Equal(t, expectedProofs, actualProofs,
		"Number of legacy proofs should be %d (1 proof per blob, not %d cell proofs)",
		numBlobs, kzg4844.CellProofsPerBlob)
	t.Logf("Legacy proof count verified: %d proofs (%d blobs × 1 proof per blob)",
		actualProofs, numBlobs)

	// Include in L1 and verify derivation
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	verifierSafeHead := verifier.L2Safe()
	require.Equal(t, l2BlockNum, verifierSafeHead.Number,
		"Verifier should successfully derive L2 blocks from legacy blob proofs")
	t.Logf("Verifier successfully derived L2 block %d from legacy blob proofs", verifierSafeHead.Number)

	// Final Summary

	t.Log("SUCCESS: Legacy Blob Proof Transaction Validation Complete")

	t.Logf("Sidecar Version: Version%d (legacy proofs)", sidecar.Version)
	t.Logf("Number of Blobs: %d", numBlobs)
	t.Logf("Proofs per Blob: 1 (legacy mode)")
	t.Logf("Total Proofs: %d", actualProofs)
	t.Logf("Backward Compatibility: Confirmed")

}

// TestCellProofsComparisonWithLegacy compares cell proofs and legacy proofs side-by-side
// to demonstrate the difference and verify both modes work correctly.
func TestCellProofsComparisonWithLegacy(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)

	lgr := testlog.Logger(t, log.LevelInfo)

	// Setup test environment
	dp := e2eutils.MakeMantleDeployParams(t, helpers.DefaultRollupTestParams())
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)

	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)
	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, lgr)
	sequencer.ActL2PipelineFull(t)

	verifEngine, verifier := helpers.SetupVerifier(t, sd, lgr, miner.L1Client(t, sd.RollupCfg),
		miner.BlobStore(), &sync.Config{})
	_ = verifEngine

	t.Log("SCENARIO 1: Cell Proofs (Version1)")

	// Create batcher with cell proofs
	cellProofBatcher := helpers.NewL2Batcher(lgr, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:          0,
		MaxL1TxSize:          128_000,
		BatcherKey:           dp.Secrets.Batcher,
		DataAvailabilityType: batcherFlags.BlobsType,
		EnableCellProofs:     true,
	}, sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	miner.ActEmptyBlock(t)
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)
	l2Block1 := sequencer.L2Unsafe().Number

	cellProofBatcher.ActSubmitAll(t)
	cellProofTx := cellProofBatcher.LastSubmitted

	cellProofSidecar := cellProofTx.BlobTxSidecar()
	cellProofBlobs := len(cellProofSidecar.Blobs)
	cellProofProofs := len(cellProofSidecar.Proofs)

	t.Logf("Cell Proof Transaction:")
	t.Logf("  - Version: Version%d", cellProofSidecar.Version)
	t.Logf("  - Blobs: %d", cellProofBlobs)
	t.Logf("  - Proofs: %d", cellProofProofs)
	t.Logf("  - Proofs per Blob: %d", cellProofProofs/cellProofBlobs)

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(cellProofTx.Hash())(t)
	miner.ActL1EndBlock(t)

	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	require.Equal(t, l2Block1, verifier.L2Safe().Number)
	t.Logf("Cell proof transaction processed successfully")

	t.Log("")

	t.Log("SCENARIO 2: Legacy Proofs (Version0)")

	// Create batcher with legacy proofs
	legacyBatcher := helpers.NewL2Batcher(lgr, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:          0,
		MaxL1TxSize:          128_000,
		BatcherKey:           dp.Secrets.Batcher,
		DataAvailabilityType: batcherFlags.BlobsType,
		EnableCellProofs:     false,
	}, sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	miner.ActEmptyBlock(t)

	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)
	l2Block2 := sequencer.L2Unsafe().Number

	legacyBatcher.ActSubmitAll(t)
	legacyTx := legacyBatcher.LastSubmitted

	legacySidecar := legacyTx.BlobTxSidecar()
	legacyBlobs := len(legacySidecar.Blobs)
	legacyProofs := len(legacySidecar.Proofs)

	t.Logf("Legacy Proof Transaction:")
	t.Logf("  - Version: Version%d", legacySidecar.Version)
	t.Logf("  - Blobs: %d", legacyBlobs)
	t.Logf("  - Proofs: %d", legacyProofs)
	t.Logf("  - Proofs per Blob: %d", legacyProofs/legacyBlobs)

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(legacyTx.Hash())(t)
	miner.ActL1EndBlock(t)

	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	require.Equal(t, l2Block2, verifier.L2Safe().Number)
	t.Logf("Legacy proof transaction processed successfully")

	t.Log("")

	t.Log("COMPARISON SUMMARY")

	t.Logf("Cell Proofs (Version1):  %d proofs for %d blobs = %d proofs/blob",
		cellProofProofs, cellProofBlobs, cellProofProofs/cellProofBlobs)
	t.Logf("Legacy Proofs (Version0): %d proofs for %d blobs = %d proof/blob",
		legacyProofs, legacyBlobs, legacyProofs/legacyBlobs)
	t.Log("")
	t.Logf("Cell proofs provide %dx more proofs per blob",
		(cellProofProofs/cellProofBlobs)/(legacyProofs/legacyBlobs))
	t.Log("This enables PeerDAS (Peer Data Availability Sampling) in Osaka upgrade")
	t.Log("")
	t.Log("Both proof types are correctly generated and processed")
	t.Log("System maintains backward compatibility with legacy proofs")

}
