package proofs

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/mantletests/proofs/helpers"
	mantlebindings "github.com/ethereum-optimism/optimism/op-e2e/mantlebindings/bindings"
	"github.com/stretchr/testify/require"
)

// Run a test that proves a deposit-only block generated due to sequence window expiry.
func runSequenceWindowExpireTest(gt *testing.T, testCfg *helpers.TestCfg[any]) {
	t := actionsHelpers.NewDefaultTesting(gt)
	tp := helpers.NewTestParams()
	env := helpers.NewL2ProofEnv(t, testCfg, tp, helpers.NewBatcherCfg())

	// Mine an empty block for gas estimation purposes.
	env.Miner.ActEmptyBlock(t)

	// Create Mantle OptimismPortal binding for 7-parameter deposit transaction
	portalAddr := env.Dp.DeployConfig.OptimismPortalProxy
	mantlePortal, err := mantlebindings.NewOptimismPortal(portalAddr, env.Miner.EthClient())
	require.NoError(t, err, "failed to create Mantle OptimismPortal binding")

	// Expire the sequence window by building `SequenceWindow + 1` empty blocks on L1.
	for i := 0; i < int(tp.SequencerWindowSize)+1; i++ {
		env.Alice.L1.ActResetTxOpts(t)

		// Use Mantle 7-parameter depositTransaction instead of standard 5-parameter version
		// Solidity: function depositTransaction(uint256 _ethTxValue, uint256 _mntValue, address _to, uint256 _mntTxValue, uint64 _gasLimit, bool _isCreation, bytes _data) payable returns()
		aliceTxOpts, err := bind.NewKeyedTransactorWithChainID(env.Dp.Secrets.Alice, env.Sd.RollupCfg.L1ChainID)
		require.NoError(t, err)
		aliceTxOpts.Value = big.NewInt(0)
		aliceTxOpts.NoSend = true

		tx, err := mantlePortal.DepositTransaction(
			aliceTxOpts,
			big.NewInt(0),        // _ethTxValue: ETH value in L2 transaction
			big.NewInt(0),        // _mntValue: MNT to transfer from L1 to portal
			env.Dp.Addresses.Bob, // _to: target address on L2
			big.NewInt(0),        // _mntTxValue: MNT value to send in L2 transaction
			uint64(100000),       // _gasLimit: minimum L2 gas limit
			false,                // _isCreation: not a contract creation
			[]byte{},             // _data: empty calldata
		)
		require.NoError(t, err, "failed to create deposit tx")

		err = env.Miner.EthClient().SendTransaction(t.Ctx(), tx)
		require.NoError(t, err, "must send deposit tx")

		env.Miner.ActL1StartBlock(12)(t)
		env.Miner.ActL1IncludeTx(env.Alice.Address())(t)
		env.Miner.ActL1EndBlock(t)

		env.Miner.ActL1SafeNext(t)
		env.Miner.ActL1FinalizeNext(t)
	}

	// Ensure the safe head is still 0.
	l2SafeHead := env.Engine.L2Chain().CurrentSafeBlock()
	require.EqualValues(t, 0, l2SafeHead.Number.Uint64())

	// Ask the sequencer to derive the deposit-only L2 chain.
	env.Sequencer.ActL1HeadSignal(t)
	env.Sequencer.ActL2PipelineFull(t)

	// Ensure the safe head advanced forcefully.
	l2SafeHead = env.Engine.L2Chain().CurrentSafeBlock()
	require.Greater(t, l2SafeHead.Number.Uint64(), uint64(0))

	// Run the FPP on one of the auto-derived blocks.
	//env.RunFaultProofProgram(t, l2SafeHead.Number.Uint64()/2, testCfg.CheckResult, testCfg.InputParams...)
}

// Runs a that proves a block in a chain where the batcher opens a channel, the sequence window expires, and then the
// batcher attempts to close the channel afterwards.
func runSequenceWindowExpire_ChannelCloseAfterWindowExpiry_Test(gt *testing.T, testCfg *helpers.TestCfg[any]) {
	t := actionsHelpers.NewDefaultTesting(gt)
	tp := helpers.NewTestParams()
	env := helpers.NewL2ProofEnv(t, testCfg, tp, helpers.NewBatcherCfg())

	// Mine 2 empty blocks on L2.
	for i := 0; i < 2; i++ {
		env.Sequencer.ActL2StartBlock(t)
		env.Alice.L2.ActResetTxOpts(t)
		env.Alice.L2.ActSetTxToAddr(&env.Dp.Addresses.Bob)
		env.Alice.L2.ActMakeTx(t)
		env.Engine.ActL2IncludeTx(env.Alice.Address())(t)
		env.Sequencer.ActL2EndBlock(t)
	}

	// Open the channel on L1.
	env.Batcher.ActL2BatchBuffer(t)
	env.Batcher.ActL2BatchSubmit(t)
	env.Miner.ActL1StartBlock(12)(t)
	env.Miner.ActL1IncludeTxByHash(env.Batcher.LastSubmitted.Hash())(t)
	env.Miner.ActL1EndBlock(t)

	// Finalize the block with the first channel frame on L1.
	env.Miner.ActL1SafeNext(t)
	env.Miner.ActL1FinalizeNext(t)

	// Instruct the sequencer to derive the L2 chain from the data on L1 that the batcher just posted.
	env.Sequencer.ActL1HeadSignal(t)
	env.Sequencer.ActL2PipelineFull(t)

	// Ensure the safe head is still 0.
	l2SafeHead := env.Engine.L2Chain().CurrentSafeBlock()
	require.EqualValues(t, 0, l2SafeHead.Number.Uint64())

	// Cache the next frame data before expiring the sequence window, but don't submit it yet.
	env.Batcher.ActL2BatchBuffer(t)
	env.Batcher.ActL2ChannelClose(t)
	finalFrame := env.Batcher.ReadNextOutputFrame(t)

	// Create Mantle OptimismPortal binding for 7-parameter deposit transaction
	portalAddr := env.Dp.DeployConfig.OptimismPortalProxy
	mantlePortal, err := mantlebindings.NewOptimismPortal(portalAddr, env.Miner.EthClient())
	require.NoError(t, err, "failed to create Mantle OptimismPortal binding")

	// Expire the sequence window by building `SequenceWindow + 1` empty blocks on L1.
	for i := 0; i < int(tp.SequencerWindowSize)+1; i++ {
		env.Alice.L1.ActResetTxOpts(t)

		// Use Mantle 7-parameter depositTransaction instead of standard 5-parameter version
		// Solidity: function depositTransaction(uint256 _ethTxValue, uint256 _mntValue, address _to, uint256 _mntTxValue, uint64 _gasLimit, bool _isCreation, bytes _data) payable returns()
		aliceTxOpts, err := bind.NewKeyedTransactorWithChainID(env.Dp.Secrets.Alice, env.Sd.RollupCfg.L1ChainID)
		require.NoError(t, err)
		aliceTxOpts.Value = big.NewInt(0)
		aliceTxOpts.NoSend = true

		tx, err := mantlePortal.DepositTransaction(
			aliceTxOpts,
			big.NewInt(0),        // _ethTxValue: ETH value in L2 transaction
			big.NewInt(0),        // _mntValue: MNT to transfer from L1 to portal
			env.Dp.Addresses.Bob, // _to: target address on L2
			big.NewInt(0),        // _mntTxValue: MNT value to send in L2 transaction
			uint64(100000),       // _gasLimit: minimum L2 gas limit
			false,                // _isCreation: not a contract creation
			[]byte{},             // _data: empty calldata
		)
		require.NoError(t, err, "failed to create deposit tx")

		err = env.Miner.EthClient().SendTransaction(t.Ctx(), tx)
		require.NoError(t, err, "must send deposit tx")

		env.Miner.ActL1StartBlock(12)(t)
		env.Miner.ActL1IncludeTx(env.Alice.Address())(t)
		env.Miner.ActL1EndBlock(t)

		env.Miner.ActL1SafeNext(t)
		env.Miner.ActL1FinalizeNext(t)
	}

	// Instruct the batcher to closethe channel on L1, after the sequence window + channel timeout has elapsed.
	env.Batcher.ActL2BatchSubmitRaw(t, finalFrame)
	env.Miner.ActL1StartBlock(12)(t)
	env.Miner.ActL1IncludeTxByHash(env.Batcher.LastSubmitted.Hash())(t)
	env.Miner.ActL1EndBlock(t)

	// Finalize the block with the second channel frame on L1.
	env.Miner.ActL1SafeNext(t)
	env.Miner.ActL1FinalizeNext(t)

	// Ensure the safe head is still 0.
	l2SafeHead = env.Engine.L2Chain().CurrentSafeBlock()
	require.EqualValues(t, 0, l2SafeHead.Number.Uint64())

	// Ask the sequencer to derive the deposit-only L2 chain.
	env.Sequencer.ActL1HeadSignal(t)
	env.Sequencer.ActL2PipelineFull(t)

	// Ensure the safe head advanced forcefully.
	l2SafeHead = env.Engine.L2Chain().CurrentSafeBlock()
	require.Greater(t, l2SafeHead.Number.Uint64(), uint64(0))

	// Run the FPP on one of the auto-derived blocks.
	//env.RunFaultProofProgram(t, l2SafeHead.Number.Uint64()/2, testCfg.CheckResult, testCfg.InputParams...)
}

func Test_ProgramAction_SequenceWindowExpired(gt *testing.T) {
	matrix := helpers.NewMatrix[any]()
	defer matrix.Run(gt)

	forks := helpers.ForkMatrix{helpers.MantleArsia, helpers.MantleLatestFork}
	matrix.AddTestCase(
		"HonestClaim",
		nil,
		forks,
		runSequenceWindowExpireTest,
		helpers.ExpectNoError(),
	)

	matrix.AddTestCase(
		"ChannelCloseAfterWindowExpiry-HonestClaim",
		nil,
		forks,
		runSequenceWindowExpire_ChannelCloseAfterWindowExpiry_Test,
		helpers.ExpectNoError(),
	)

}
