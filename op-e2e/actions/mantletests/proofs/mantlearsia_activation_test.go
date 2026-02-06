package proofs

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/mantletests/proofs/helpers"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

func Test_ProgramAction_MantleArsiaActivation(gt *testing.T) {

	runMantleArsiaDerivationTest := func(gt *testing.T, testCfg *helpers.TestCfg[any]) {
		t := actionsHelpers.NewDefaultTesting(gt)

		// Define override to activate MantleArsia 14 seconds after genesis
		var setMantleArsiaTime = func(dc *genesis.DeployConfig) {
			now := uint64(time.Now().Unix())
			dc.L1GenesisBlockTimestamp = hexutil.Uint64(now)
			fourteen := uint64(14)
			dc.ActivateMantleForkAtOffset(rollup.MantleForkName("MantleArsia"), fourteen)
		}

		env := helpers.NewL2ProofEnv(t, testCfg, helpers.NewTestParams(), helpers.NewBatcherCfg(), setMantleArsiaTime)

		// Get genesis time and Arsia activation time for reference
		genesisTime := env.Sequencer.RollupCfg.Genesis.L2Time
		arsiaTime := *env.Sequencer.RollupCfg.MantleArsiaTime
		t.Logf("L2 Genesis Time: %d, MantleArsiaTime: %d", genesisTime, arsiaTime)
		t.Logf("HoloceneTime: %v", env.Sequencer.RollupCfg.HoloceneTime)
		// Build the L2 chain until the MantleArsia activation time,
		// which for the Execution Engine is an L2 block timestamp
		for env.Engine.L2Chain().CurrentBlock().Time < *env.Sequencer.RollupCfg.MantleArsiaTime {
			b := env.Engine.L2Chain().GetBlockByHash(env.Sequencer.L2Unsafe().Hash)
			require.Equal(t, "", string(b.Extra()), "extra data should be empty before MantleArsia activation")
			env.Sequencer.ActL2StartBlock(t)
			// Send an L2 tx
			// Manually create a transaction to avoid gas estimation issues in activation block
			env.Alice.L2.ActResetTxOpts(t)
			nonce := env.Alice.L2.PendingNonce(t)

			// Get latest header for gas price info
			latestHeader, err := env.Engine.EthClient().HeaderByNumber(t.Ctx(), nil)
			require.NoError(t, err)
			gasTipCap := big.NewInt(2 * params.GWei)
			gasFeeCap := new(big.Int).Add(gasTipCap, new(big.Int).Mul(latestHeader.BaseFee, big.NewInt(2)))

			toAddr := env.Bob.Address()
			tx := types.MustSignNewTx(env.Alice.L2.Secret(), env.Alice.L2.Signer(), &types.DynamicFeeTx{
				ChainID:   env.Alice.L2.Signer().ChainID(),
				Nonce:     nonce,
				GasTipCap: gasTipCap,
				GasFeeCap: gasFeeCap,
				Gas:       50_000, // Manually set gas limit to avoid estimation
				To:        &toAddr,
				Value:     big.NewInt(0),
				Data:      []byte{},
			})

			// Send the transaction to the txpool first
			require.NoError(t, env.Engine.EthClient().SendTransaction(t.Ctx(), tx))

			// Then include it in the block
			env.Engine.ActL2IncludeTx(env.Alice.Address())(t)
			env.Sequencer.ActL2EndBlock(t)
			t.Log("Unsafe block with timestamp %d", b.Time)
		}
		b := env.Engine.L2Chain().GetBlockByHash(env.Sequencer.L2Unsafe().Hash)
		require.Len(t, b.Extra(), 17, "extra data should be 17 bytes after Arsia activation")

		// Build up a local list of frames
		// For Mantle Arsia (like OP Holocene), we need to submit frames with correct format
		orderedFrames := make([][]byte, 0, 2)

		// Submit the first two blocks, this will be enough to trigger MantleArsia _derivation_
		// which is activated by the L1 inclusion block timestamp
		// block 1 will be 12 seconds after genesis, and 2 seconds before MantleArsia activation
		// block 2 will be 24 seconds after genesis, and 10 seconds after MantleArsia activation
		blocksToSubmit := []uint{1, 2}

		// Buffer the blocks in the batcher and collect frames
		env.Batcher.ActCreateChannel(t, false)
		for i, blockNum := range blocksToSubmit {
			env.Batcher.ActAddBlockByNumber(t, int64(blockNum), actionsHelpers.BlockLogger(t))
			if i == len(blocksToSubmit)-1 {
				env.Batcher.ActL2ChannelClose(t)
			}
			frame := env.Batcher.ReadNextOutputFrame(t)
			require.NotEmpty(t, frame, "frame %d", i)
			orderedFrames = append(orderedFrames, frame)
		}

		// Calculate the expected L1 block times
		// Each L1 block is 12 seconds apart
		l1BlockTime1 := genesisTime + 12 // 12s after genesis, before Arsia (14s)
		l1BlockTime2 := genesisTime + 24 // 24s after genesis, after Arsia (14s)

		t.Logf("L1 Block 1 time: %d (Arsia active: %v)", l1BlockTime1, l1BlockTime1 >= arsiaTime)
		t.Logf("L1 Block 2 time: %d (Arsia active: %v)", l1BlockTime2, l1BlockTime2 >= arsiaTime)

		includeBatchTx := func() {
			// Include the last transaction submitted by the batcher.
			env.Miner.ActL1StartBlock(12)(t)
			env.Miner.ActL1IncludeTxByHash(env.Batcher.LastSubmitted.Hash())(t)
			env.Miner.ActL1EndBlock(t)
		}

		// Submit first frame (before Arsia activation)
		// Use l1BlockTime1 to ensure correct format (RLP-encoded for pre-Arsia)
		env.Batcher.ActL2BatchSubmitMantleRawAtTime(t, orderedFrames[0], l1BlockTime1)
		includeBatchTx() // L1 block should have a timestamp of 12s after genesis

		// Arsia should activate 14s after genesis, so that the previous l1 block
		// was before ArsiaTime and the next l1 block is after it

		// Submit second frame (after Arsia activation)
		// Use l1BlockTime2 to ensure correct format (standard OP Stack for post-Arsia)
		env.Batcher.ActL2BatchSubmitMantleRawAtTime(t, orderedFrames[1], l1BlockTime2)
		includeBatchTx() // block should have a timestamp of 24s after genesis

		// Instruct the sequencer to derive the L2 chain from the data on L1 that the batcher just posted.
		env.Sequencer.ActL1HeadSignal(t)
		env.Sequencer.ActL2PipelineFull(t)

		l2SafeHead := env.Sequencer.L2Safe()
		t.Logf("L2 Safe Head: Number=%d, Time=%d", l2SafeHead.Number, l2SafeHead.Time)

		// With correct format matching, the channel should be processed successfully
		// and the safe head should progress to block 2
		require.EqualValues(t, uint64(2), l2SafeHead.Number, "safe head should progress to block 2 with correct format")

		t.Log("Safe head progressed as expected", "l2SafeHeadNumber", l2SafeHead.Number)

		droppingLogs := env.Logs.FindLogs(testlog.NewMessageContainsFilter("dropping non-first frame"), testlog.NewAttributesFilter("role", "sequencer"))
		require.Len(t, droppingLogs, 0, "should not have any dropping frame logs with correct format")
	}

	matrix := helpers.NewMatrix[any]()
	defer matrix.Run(gt)

	matrix.AddTestCase(
		"HonestClaim-MantleArsiaActivation",
		nil,
		helpers.NewForkMatrix(helpers.MantleArsia),
		runMantleArsiaDerivationTest,
		helpers.ExpectNoError(),
	)
}
