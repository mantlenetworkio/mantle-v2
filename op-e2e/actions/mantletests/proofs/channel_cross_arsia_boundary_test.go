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

// TestChannelCrossArsiaActivationBoundary verifies that when a single channel's frames
// span across the Arsia activation boundary (Frame 0 before Arsia with RLP format,
// Frame 1 after Arsia with standard OP Stack format), the channel cannot be completed
// and the safe head does not progress.
//
// This test documents a known issue: if the batcher submits a channel that crosses
// the Arsia activation boundary, the format mismatch will cause frame drops.
// In production, the batcher should be stopped before Arsia activation to ensure
// all channels complete before the fork.
func TestChannelCrossArsiaActivationBoundary(gt *testing.T) {

	runTest := func(gt *testing.T, testCfg *helpers.TestCfg[any]) {
		t := actionsHelpers.NewDefaultTesting(gt)

		// Define override to activate MantleArsia 14 seconds after genesis
		var setMantleArsiaTime = func(dc *genesis.DeployConfig) {
			now := uint64(time.Now().Unix())
			dc.L1GenesisBlockTimestamp = hexutil.Uint64(now)
			fourteen := uint64(14)
			dc.ActivateMantleForkAtOffset(rollup.MantleForkName("MantleArsia"), fourteen)
			// Set tokenRatio to 1 to avoid gas calculation issues in pre-Arsia blocks
			dc.GasPriceOracleTokenRatio = 1
		}

		env := helpers.NewL2ProofEnv(t, testCfg, helpers.NewTestParams(), helpers.NewBatcherCfg(), setMantleArsiaTime)

		// Get genesis time and Arsia activation time for reference
		genesisTime := env.Sequencer.RollupCfg.Genesis.L2Time
		arsiaTime := *env.Sequencer.RollupCfg.MantleArsiaTime
		t.Logf("L2 Genesis Time: %d, MantleArsiaTime: %d", genesisTime, arsiaTime)

		// Build the L2 chain until the MantleArsia activation time
		for env.Engine.L2Chain().CurrentBlock().Time < *env.Sequencer.RollupCfg.MantleArsiaTime {
			env.Sequencer.ActL2StartBlock(t)
			env.Alice.L2.ActResetTxOpts(t)
			nonce := env.Alice.L2.PendingNonce(t)

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
				Gas:       55_000, // Increased to cover L1 cost in pre-Arsia blocks
				To:        &toAddr,
				Value:     big.NewInt(0),
				Data:      []byte{},
			})

			require.NoError(t, env.Engine.EthClient().SendTransaction(t.Ctx(), tx))
			env.Engine.ActL2IncludeTx(env.Alice.Address())(t)
			env.Sequencer.ActL2EndBlock(t)
		}

		// Calculate the expected L1 block times
		l1BlockTime1 := genesisTime + 12 // 12s after genesis, before Arsia (14s)
		l1BlockTime2 := genesisTime + 24 // 24s after genesis, after Arsia (14s)

		t.Logf("L1 Block 1 time: %d (Arsia active: %v)", l1BlockTime1, l1BlockTime1 >= arsiaTime)
		t.Logf("L1 Block 2 time: %d (Arsia active: %v)", l1BlockTime2, l1BlockTime2 >= arsiaTime)

		// === Create a SINGLE channel with 2 frames that will cross the Arsia boundary ===
		orderedFrames := make([][]byte, 0, 2)

		env.Batcher.ActCreateChannel(t, false)
		for i, blockNum := range []uint{1, 2} {
			env.Batcher.ActAddBlockByNumber(t, int64(blockNum), actionsHelpers.BlockLogger(t))
			if i == 1 {
				env.Batcher.ActL2ChannelClose(t)
			}
			frame := env.Batcher.ReadNextOutputFrame(t)
			require.NotEmpty(t, frame, "frame %d", i)
			orderedFrames = append(orderedFrames, frame)
		}

		includeBatchTx := func() {
			env.Miner.ActL1StartBlock(12)(t)
			env.Miner.ActL1IncludeTxByHash(env.Batcher.LastSubmitted.Hash())(t)
			env.Miner.ActL1EndBlock(t)
		}

		// Submit Frame 0 BEFORE Arsia activation (uses RLP format)
		env.Batcher.ActL2BatchSubmitMantleRawAtTime(t, orderedFrames[0], l1BlockTime1)
		includeBatchTx()

		// Submit Frame 1 AFTER Arsia activation (uses standard OP Stack format)
		// This creates a format mismatch within the same channel!
		env.Batcher.ActL2BatchSubmitMantleRawAtTime(t, orderedFrames[1], l1BlockTime2)
		includeBatchTx()

		// Derive the L2 chain
		env.Sequencer.ActL1HeadSignal(t)
		env.Sequencer.ActL2PipelineFull(t)

		l2SafeHead := env.Sequencer.L2Safe()
		t.Logf("L2 Safe Head: Number=%d, Time=%d", l2SafeHead.Number, l2SafeHead.Time)

		// === Verify: Safe head should NOT progress due to channel format mismatch ===
		require.EqualValues(t, uint64(0), l2SafeHead.Number,
			"safe head should remain at 0 because channel frames have mismatched formats")

		// Verify that frame dropping occurred
		droppingLogs := env.Logs.FindLogs(
			testlog.NewMessageContainsFilter("dropping non-first frame"),
			testlog.NewAttributesFilter("role", "sequencer"),
		)
		require.Greater(t, len(droppingLogs), 0,
			"should have 'dropping non-first frame' log due to format mismatch")

		t.Log("Test passed: Channel crossing Arsia boundary correctly fails to complete")
	}

	matrix := helpers.NewMatrix[any]()
	defer matrix.Run(gt)

	matrix.AddTestCase(
		"ChannelCrossArsiaActivationBoundary",
		nil,
		helpers.NewForkMatrix(helpers.MantleArsia),
		runTest,
		helpers.ExpectNoError(),
	)
}
