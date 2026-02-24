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
			dc.GasPriceOracleTokenRatio = 1
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

		// === Channel 1: Submit block 1 before Arsia activation (RLP format) ===
		env.Batcher.ActCreateChannel(t, false)
		env.Batcher.ActAddBlockByNumber(t, 1, actionsHelpers.BlockLogger(t))
		env.Batcher.ActL2ChannelClose(t)
		frame1 := env.Batcher.ReadNextOutputFrame(t)
		require.NotEmpty(t, frame1, "frame 1 should not be empty")

		// Submit channel 1 before Arsia activation
		env.Batcher.ActL2BatchSubmitMantleRawAtTime(t, frame1, l1BlockTime1)
		includeBatchTx() // L1 block should have a timestamp of 12s after genesis

		// === Channel 2: Submit block 2 after Arsia activation (standard OP Stack format) ===
		env.Batcher.ActCreateChannel(t, false)
		env.Batcher.ActAddBlockByNumber(t, 2, actionsHelpers.BlockLogger(t))
		env.Batcher.ActL2ChannelClose(t)
		frame2 := env.Batcher.ReadNextOutputFrame(t)
		require.NotEmpty(t, frame2, "frame 2 should not be empty")

		// Submit channel 2 after Arsia activation
		env.Batcher.ActL2BatchSubmitMantleRawAtTime(t, frame2, l1BlockTime2)
		includeBatchTx() // L1 block should have a timestamp of 24s after genesis

		// Instruct the sequencer to derive the L2 chain from the data on L1 that the batcher just posted.
		env.Sequencer.ActL1HeadSignal(t)
		env.Sequencer.ActL2PipelineFull(t)

		l2SafeHead := env.Sequencer.L2Safe()
		t.Logf("L2 Safe Head: Number=%d, Time=%d", l2SafeHead.Number, l2SafeHead.Time)

		// With separate channels using correct format for each, both should be processed successfully
		// and the safe head should progress to block 2
		require.EqualValues(t, uint64(2), l2SafeHead.Number, "safe head should progress to block 2 with correct format")

		t.Log("Safe head progressed as expected", "l2SafeHeadNumber", l2SafeHead.Number)

		droppingLogs := env.Logs.FindLogs(testlog.NewMessageContainsFilter("dropping non-first frame"), testlog.NewAttributesFilter("role", "sequencer"))
		require.Len(t, droppingLogs, 0, "should not have any dropping frame logs with separate channels")
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
