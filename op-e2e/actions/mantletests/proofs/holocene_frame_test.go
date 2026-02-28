package proofs

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/mantletests/proofs/helpers"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

func Test_ProgramAction_ArsiaFrames(gt *testing.T) {
	type testCase struct {
		name   string
		frames []uint
		arsiaExpectations
	}

	// An ordered list of frames to read from the channel and submit
	// on L1. We expect a different progression of the safe head under Arsia
	// derivation rules, compared with pre Arsia.
	testCases := []testCase{
		// Standard frame submission,
		{
			name: "ordered", frames: []uint{0, 1, 2},
			arsiaExpectations: arsiaExpectations{
				preArsia: expectations{safeHead: 3},
				arsia:    expectations{safeHead: 3},
			},
		},

		// Non-standard frame submission
		{
			name: "disordered-a", frames: []uint{2, 1, 0},
			arsiaExpectations: arsiaExpectations{
				preArsia: expectations{safeHead: 3}, // frames are buffered, so ordering does not matter
				arsia:    expectations{safeHead: 0}, // non-first frames will be dropped b/c it is the first seen with that channel Id. The safe head won't move until the channel is closed/completed.
			},
		},
		{
			name: "disordered-b", frames: []uint{0, 1, 0, 2},
			arsiaExpectations: arsiaExpectations{
				preArsia: expectations{safeHead: 3}, // frames are buffered, so ordering does not matter
				arsia:    expectations{safeHead: 0}, // non-first frames will be dropped b/c it is the first seen with that channel Id. The safe head won't move until the channel is closed/completed.
			},
		},
		{
			name: "duplicates", frames: []uint{0, 1, 1, 2},
			arsiaExpectations: arsiaExpectations{
				preArsia: expectations{safeHead: 3}, // frames are buffered, so ordering does not matter
				arsia:    expectations{safeHead: 3}, // non-contiguous frames are dropped. So this reduces to case-0.
			},
		},
	}

	runArsiaDerivationTest := func(gt *testing.T, testCfg *helpers.TestCfg[testCase]) {
		t := actionsHelpers.NewDefaultTesting(gt)
		env := helpers.NewL2ProofEnv(t, testCfg, helpers.NewTestParams(), helpers.NewBatcherCfg(), func(dc *genesis.DeployConfig) {
			// Set tokenRatio to 1 to avoid gas calculation issues in MantleLimb
			dc.GasPriceOracleTokenRatio = 1
		})

		blocks := []uint{1, 2, 3}
		targetHeadNumber := 3
		nonce := uint64(0)
		for env.Engine.L2Chain().CurrentBlock().Number.Uint64() < uint64(targetHeadNumber) {
			// Get gas price parameters
			latestHeader, err := env.Engine.EthClient().HeaderByNumber(t.Ctx(), nil)
			require.NoError(t, err)
			gasTipCap := big.NewInt(2 * params.GWei)
			gasFeeCap := new(big.Int).Add(gasTipCap, new(big.Int).Mul(latestHeader.BaseFee, big.NewInt(2)))

			// Send an L2 tx with fixed gas limit
			toAddr := env.Dp.Addresses.Bob
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

			// Send transaction to tx pool
			err = env.Engine.EthClient().SendTransaction(t.Ctx(), tx)
			require.NoError(t, err, "must send tx")

			// Include transaction in block
			env.Sequencer.ActL2StartBlock(t)
			env.Engine.ActL2IncludeTx(env.Alice.Address())(t)
			env.Sequencer.ActL2EndBlock(t)
			nonce++
		}

		// Build up a local list of frames
		orderedFrames := make([][]byte, 0, len(testCfg.Custom.frames))
		// Buffer the blocks in the batcher and populat orderedFrames list
		env.Batcher.ActCreateChannel(t, false)
		for i, blockNum := range blocks {
			env.Batcher.ActAddBlockByNumber(t, int64(blockNum), actionsHelpers.BlockLogger(t))
			if i == len(blocks)-1 {
				env.Batcher.ActL2ChannelClose(t)
			}
			frame := env.Batcher.ReadNextOutputFrame(t)
			require.NotEmpty(t, frame, "frame %d", i)
			orderedFrames = append(orderedFrames, frame)
		}

		includeBatchTx := func() {
			// Include the last transaction submitted by the batcher.
			env.Miner.ActL1StartBlock(12)(t)
			env.Miner.ActL1IncludeTxByHash(env.Batcher.LastSubmitted.Hash())(t)
			env.Miner.ActL1EndBlock(t)

			// Finalize the block with the first channel frame on L1.
			env.Miner.ActL1SafeNext(t)
			env.Miner.ActL1FinalizeNext(t)
		}

		// Submit frames in specified order order
		for _, j := range testCfg.Custom.frames {
			env.Batcher.ActL2BatchSubmitMantleRaw(t, orderedFrames[j])
			includeBatchTx()
		}

		// Instruct the sequencer to derive the L2 chain from the data on L1 that the batcher just posted.
		env.Sequencer.ActL1HeadSignal(t)
		env.Sequencer.ActL2PipelineFull(t)

		l2SafeHead := env.Sequencer.L2Safe()

		isArsia := testCfg.Hardfork.Precedence >= helpers.MantleArsia.Precedence
		testCfg.Custom.RequireExpectedProgressAndLogs(t, l2SafeHead, isArsia, env.Engine, env.Logs)
		t.Log("Safe head progressed as expected", "l2SafeHeadNumber", l2SafeHead.Number)

		//env.RunFaultProofProgramFromGenesis(t, l2SafeHead.Number, testCfg.CheckResult, testCfg.InputParams...)
	}

	matrix := helpers.NewMatrix[testCase]()
	defer matrix.Run(gt)

	for _, ordering := range testCases {
		matrix.AddTestCase(
			fmt.Sprintf("HonestClaim-%s", ordering.name),
			ordering,
			helpers.NewForkMatrix(helpers.MantleLimb, helpers.MantleArsia, helpers.MantleLatestFork),
			runArsiaDerivationTest,
			helpers.ExpectNoError(),
		)
	}
}
