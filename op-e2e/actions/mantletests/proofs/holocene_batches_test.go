package proofs

import (
	"fmt"
	"math/big"
	"testing"

	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/mantletests/proofs/helpers"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

func Test_ProgramAction_ArsiaBatches(gt *testing.T) {
	type testCase struct {
		name        string
		blocks      []uint // blocks is an ordered list of blocks (by number) to add to a single channel.
		isSpanBatch bool
		arsiaExpectations
	}

	// Depending on the blocks list,  we expect a different
	// progression of the safe head under Arsia
	// derivation rules, compared with pre Arsia.
	testCases := []testCase{
		// Standard channel composition
		{
			name: "ordered", blocks: []uint{1, 2, 3},
			arsiaExpectations: arsiaExpectations{
				preArsia: expectations{safeHead: 3},
				arsia:    expectations{safeHead: 3},
			},
		},

		// Non-standard channel composition
		{
			name: "disordered-a", blocks: []uint{1, 3, 2},
			arsiaExpectations: arsiaExpectations{
				preArsia: expectations{safeHead: 3}, // batches are buffered, so the block ordering does not matter
				arsia: expectations{safeHead: 1, // batch for block 3 is considered invalid because it is from the future. This batch + remaining channel is dropped.
					logs: append(
						sequencerOnce("dropping future batch"),
						sequencerOnce("Dropping invalid singular batch, flushing channel")...,
					)},
			},
		},
		{
			name: "disordered-b", blocks: []uint{2, 1, 3},
			arsiaExpectations: arsiaExpectations{
				preArsia: expectations{safeHead: 3}, // batches are buffered, so the block ordering does not matter
				arsia: expectations{safeHead: 0, // batch for block 2 is considered invalid because it is from the future. This batch + remaining channel is dropped.
					logs: append(
						sequencerOnce("dropping future batch"),
						sequencerOnce("Dropping invalid singular batch, flushing channel")...,
					)},
			},
		},

		{
			name: "duplicates-a", blocks: []uint{1, 1, 2, 3},
			arsiaExpectations: arsiaExpectations{
				preArsia: expectations{safeHead: 3}, // duplicate batches are dropped, so this reduces to the "ordered" case
				arsia: expectations{safeHead: 3, // duplicate batches are dropped, so this reduces to the "ordered" case
					logs: sequencerOnce("dropping past batch with old timestamp")},
			},
		},
		{
			name: "duplicates-b", blocks: []uint{2, 2, 1, 3},
			arsiaExpectations: arsiaExpectations{
				preArsia: expectations{safeHead: 3}, // duplicate batches are silently dropped, so this reduces to disordered-2b
				arsia: expectations{safeHead: 0, // duplicate batches are silently dropped, so this reduces to disordered-2b
					logs: append(
						sequencerOnce("dropping future batch"),
						sequencerOnce("Dropping invalid singular batch, flushing channel")...,
					)},
			},
		},
	}

	runArsiaDerivationTest := func(gt *testing.T, testCfg *helpers.TestCfg[testCase]) {
		t := actionsHelpers.NewDefaultTesting(gt)
		env := helpers.NewL2ProofEnv(t, testCfg, helpers.NewTestParams(), helpers.NewBatcherCfg())

		includeBatchTx := func() {
			// Include the last transaction submitted by the batcher.
			env.Miner.ActL1StartBlock(12)(t)
			env.Miner.ActL1IncludeTxByHash(env.Batcher.LastSubmitted.Hash())(t)
			env.Miner.ActL1EndBlock(t)
		}

		max := func(input []uint) uint {
			max := uint(0)
			for _, val := range input {
				if val > max {
					max = val
				}
			}
			return max
		}

		targetHeadNumber := max(testCfg.Custom.blocks)
		nonce := uint64(0)
		for env.Engine.L2Chain().CurrentBlock().Number.Uint64() < uint64(targetHeadNumber) {
			// Get gas price parameters
			latestHeader, err := env.Engine.EthClient().HeaderByNumber(t.Ctx(), nil)
			require.NoError(t, err)
			gasTipCap := big.NewInt(2 * params.GWei)
			gasFeeCap := new(big.Int).Add(gasTipCap, new(big.Int).Mul(latestHeader.BaseFee, big.NewInt(2)))

			// Send an L2 tx with fixed gas limit of 50000
			toAddr := env.Dp.Addresses.Bob
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

			err = env.Engine.EthClient().SendTransaction(t.Ctx(), tx)
			require.NoError(t, err, "must send tx")
			env.Sequencer.ActL2StartBlock(t)
			env.Engine.ActL2IncludeTx(env.Alice.Address())(t)
			env.Sequencer.ActL2EndBlock(t)
			nonce++
		}

		// Buffer the blocks in the batcher.
		env.Batcher.ActCreateChannel(t, testCfg.Custom.isSpanBatch)
		for _, blockNum := range testCfg.Custom.blocks {
			env.Batcher.ActAddBlockByNumber(t, int64(blockNum), actionsHelpers.BlockLogger(t))
		}
		env.Batcher.ActL2ChannelClose(t)
		frame := env.Batcher.ReadNextOutputFrame(t)
		require.NotEmpty(t, frame)
		env.Batcher.ActL2BatchSubmitMantleRaw(t, frame)
		includeBatchTx()

		// Instruct the sequencer to derive the L2 chain from the data on L1 that the batcher just posted.
		env.Sequencer.ActL1HeadSignal(t)
		env.Sequencer.ActL2PipelineFull(t)

		l2SafeHead := env.Sequencer.L2Safe()
		isHoloceneBehavior := testCfg.Hardfork.Precedence >= helpers.MantleArsia.Precedence
		testCfg.Custom.RequireExpectedProgressAndLogs(t, l2SafeHead, isHoloceneBehavior, env.Engine, env.Logs)
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
		/*
			matrix.AddTestCase(
				fmt.Sprintf("JunkClaim-%s", ordering.name),
				ordering,
				helpers.NewForkMatrix(helpers.MantleArsia),
				runArsiaDerivationTest,
				helpers.ExpectError(claim.ErrClaimNotValid),
				helpers.WithL2Claim(common.HexToHash("0xdeadbeef")),
			)
		*/
	}
}
