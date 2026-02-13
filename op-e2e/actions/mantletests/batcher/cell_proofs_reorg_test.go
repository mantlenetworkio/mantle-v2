package batcher

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	batcherFlags "github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	upgradesHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// TestBlobWithCellProofsReorg verifies that blob with cell proofs works correctly through L1 reorg.
// Test flow:
// 1. Submit L2 batch as blob with 128 cell proofs per blob
// 2. Include blob in L1 block A, verifier derives L2
// 3. L1 reorg orphans block A, verifier rewinds L2 safe head
// 4. Re-submit same blob to new L1 chain, verifier re-derives L2
// 5. Verify final state matches original L2 state
func TestBlobWithCellProofsReorg(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)

	// Setup test environment with Arsia (cell proofs support) activated at genesis
	dp := e2eutils.MakeMantleDeployParams(t, helpers.DefaultRollupTestParams())
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)
	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)

	// Setup sequencer
	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, testlog.Logger(t, log.LevelInfo))
	sequencer.ActL2PipelineFull(t)

	// Setup verifier
	verifierEngine, verifier := helpers.SetupVerifier(t, sd, testlog.Logger(t, log.LevelInfo), miner.L1Client(t, sd.RollupCfg), miner.BlobStore(), &sync.Config{})
	verifierEngineClient := verifierEngine.EngineClient(t, sd.RollupCfg)

	// Setup batcher with cell proofs enabled
	batcher := helpers.NewL2Batcher(testlog.Logger(t, log.LevelInfo), sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:          0,
		MaxL1TxSize:          128_000,
		BatcherKey:           dp.Secrets.Batcher,
		DataAvailabilityType: batcherFlags.BlobsType,
		EnableCellProofs:     true, // Enable 128 cell proofs per blob (PeerDAS/EIP-7594)
	}, sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	// Build L1 block
	miner.ActEmptyBlock(t)
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	// Build L2 blocks
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)
	l2UnsafeHead := sequencer.L2Unsafe()

	// Submit L2 batch as blob with cell proofs
	batcher.ActSubmitAll(t)
	batchTx := batcher.LastSubmitted

	// Include blob in L1 block A
	miner.ActL1SetFeeRecipient(common.Address{'A'})
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTx.Hash())(t)
	miner.ActL1EndBlock(t)
	blockA := miner.L1Chain().CurrentBlock()

	// Verify blob transaction has 128 cell proofs per blob
	require.Equal(t, uint8(types.BlobTxType), batchTx.Type())
	sidecar := batchTx.BlobTxSidecar()
	require.NotNil(t, sidecar)
	require.Equal(t, len(sidecar.Blobs)*128, len(sidecar.Proofs))

	// Verifier derives L2 from blob
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	verifierSafeHead := verifier.L2Safe()
	require.Equal(t, l2UnsafeHead, verifierSafeHead)

	// L1 reorg: orphan block A
	miner.ActL1RewindToParent(t)
	miner.ActL1SetFeeRecipient(common.Address{'B'})
	miner.ActEmptyBlock(t)
	miner.ActEmptyBlock(t) // Need longer chain for reorg
	blockB := miner.L1Chain().CurrentBlock()
	require.NotEqual(t, blockA.Hash(), blockB.Hash())

	// Verifier detects L1 reorg and rewinds L2 safe head
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	verifierSafeHeadAfterReorg := verifier.L2Safe()
	require.Less(t, verifierSafeHeadAfterReorg.Number, verifierSafeHead.Number)

	// Verify engine state matches rollup client
	ref, err := verifierEngineClient.L2BlockRefByLabel(t.Ctx(), eth.Safe)
	require.NoError(t, err)
	require.Equal(t, verifierSafeHeadAfterReorg, ref)

	// Re-submit blob to new L1 chain
	miner.ActL1StartBlock(12)(t)
	miner.ActL1SetFeeRecipient(common.Address{'C'})
	require.NoError(t, miner.Eth.TxPool().Add([]*types.Transaction{batchTx}, true)[0])
	miner.ActL1IncludeTxByHash(batchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	// Verifier re-derives L2 from replayed blob
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)
	verifierFinalSafeHead := verifier.L2Safe()
	require.Equal(t, l2UnsafeHead, verifierFinalSafeHead)

	// Verify engine state matches rollup client after re-derivation
	refFinal, err := verifierEngineClient.L2BlockRefByLabel(t.Ctx(), eth.Safe)
	require.NoError(t, err)
	require.Equal(t, verifierFinalSafeHead, refFinal)

	// Verify sequencer syncs to same state
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)
	require.Equal(t, verifierFinalSafeHead, sequencer.L2Safe())
}
