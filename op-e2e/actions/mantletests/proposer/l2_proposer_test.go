package proposer

import (
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-e2e/config"

	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	upgradesHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	mantlebindings "github.com/ethereum-optimism/optimism/op-e2e/mantlebindings/bindings"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// TestProposerBatchType run each proposer-related test case in singular batch mode and span batch mode.
func TestProposerBatchType(t *testing.T) {
	t.Run("SingularBatch/L2output", func(t *testing.T) {
		runProposerTest(t, false, config.DefaultAllocType)
	})
	t.Run("SpanBatch/L2output", func(t *testing.T) {
		runProposerTest(t, true, config.DefaultAllocType)
	})
}

func runProposerTest(gt *testing.T, isSpanBatch bool, allocType config.AllocType) {
	t := actionsHelpers.NewDefaultTesting(gt)
	params := actionsHelpers.DefaultRollupTestParams()
	params.AllocType = allocType
	dp := e2eutils.MakeMantleDeployParams(t, params)
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)
	sd := e2eutils.SetupMantleNormal(t, dp, actionsHelpers.DefaultAlloc)
	log := testlog.Logger(t, log.LevelDebug)
	miner, seqEngine, sequencer := actionsHelpers.SetupSequencerTest(t, sd, log)

	rollupSeqCl := sequencer.RollupClient()
	batcher := actionsHelpers.NewL2Batcher(log, sd.RollupCfg, actionsHelpers.MantleSpanBatcherCfg(dp),
		rollupSeqCl, miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))
	if !isSpanBatch {
		batcher = actionsHelpers.NewL2Batcher(log, sd.RollupCfg, actionsHelpers.MantleSingularBatcherCfg(dp),
			rollupSeqCl, miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))
	}

	var proposer *actionsHelpers.MantleL2Proposer

	proposer = actionsHelpers.NewMantleL2Proposer(t, log, &actionsHelpers.MantleProposerCfg{
		OutputOracleAddr:      &sd.DeploymentsL1.L2OutputOracleProxy,
		ProposerKey:           dp.Secrets.Proposer,
		ProposalRetryInterval: 3 * time.Second,
		AllowNonFinalized:     false,
		AllocType:             allocType,
		ChainID:               eth.ChainIDFromBig(sd.L1Cfg.Config.ChainID),
	}, miner.EthClient(), rollupSeqCl)

	// L1 block
	miner.ActEmptyBlock(t)
	// L2 block
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)
	sequencer.ActBuildToL1Head(t)
	// submit and include in L1
	batcher.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)
	// finalize the first and second L1 blocks, including the batch
	miner.ActL1SafeNext(t)
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)
	miner.ActL1FinalizeNext(t)
	// derive and see the L2 chain fully finalize
	sequencer.ActL2PipelineFull(t)
	sequencer.ActL1SafeSignal(t)
	sequencer.ActL1FinalizedSignal(t)
	sequencer.ActL2PipelineFull(t)
	require.Equal(t, sequencer.SyncStatus().UnsafeL2, sequencer.SyncStatus().FinalizedL2)
	require.True(t, proposer.CanMantlePropose(t))

	maxProposals := 5
	proposalCount := 0
	// make proposals until there is nothing left to propose
	for proposer.CanMantlePropose(t) && proposalCount < maxProposals {
		proposer.ActMantleMakeProposalTx(t)
		// include proposal on L1
		miner.ActL1StartBlock(12)(t)
		miner.ActL1IncludeTx(dp.Addresses.Proposer)(t)
		miner.ActL1EndBlock(t)
		// Check proposal was successful
		receipt, err := miner.EthClient().TransactionReceipt(t.Ctx(), proposer.LastMantleProposalTx())
		require.NoError(t, err)
		require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status, "proposal failed")
		proposalCount++
		sequencer.ActL1HeadSignal(t)
		sequencer.ActL2PipelineFull(t)
	}

	// check that L1 stored the expected output root

	outputOracleContract, err := mantlebindings.NewL2OutputOracle(sd.DeploymentsL1.L2OutputOracleProxy, miner.EthClient())
	require.NoError(t, err)
	blockNumber, err := outputOracleContract.LatestBlockNumber(&bind.CallOpts{})
	require.NoError(t, err)
	require.Greater(t, int64(blockNumber.Uint64()), int64(0), "latest block number must be greater than 0")
	block, err := seqEngine.EthClient().BlockByNumber(t.Ctx(), blockNumber)
	require.NoError(t, err)
	outputOnL1, err := outputOracleContract.GetL2OutputAfter(&bind.CallOpts{}, blockNumber)
	require.NoError(t, err)
	require.Less(t, block.Time(), outputOnL1.Timestamp.Uint64(), "output is registered with L1 timestamp of proposal tx, past L2 block")
	outputComputed, err := sequencer.RollupClient().OutputAtBlock(t.Ctx(), blockNumber.Uint64())
	require.NoError(t, err)
	require.Equal(t, eth.Bytes32(outputOnL1.OutputRoot), outputComputed.OutputRoot, "output roots must match")

}
