package actions

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

func Test_Operator_Fee_Constistency(gt *testing.T) {

	const testOperatorFeeScalar = uint64(20000)
	const testOperatorFeeConstant = uint64(500)

	t := NewDefaultTesting(gt)
	dp := e2eutils.MakeDeployParams(t, defaultRollupTestParams)
	dp.DeployConfig.GasPriceOracleOperatorFeeConstant = testOperatorFeeConstant
	dp.DeployConfig.GasPriceOracleOperatorFeeScalar = testOperatorFeeScalar

	sd := e2eutils.Setup(t, dp, defaultAlloc)
	log := testlog.Logger(t, log.LvlDebug)
	miner, seqEngine, sequencer := setupSequencerTest(t, sd, log)

	alice := NewBasicUser[any](log, dp.Secrets.Alice, rand.New(rand.NewSource(1234)))
	alice.SetUserEnv(&BasicUserEnv[any]{
		EthCl:  seqEngine.EthClient(),
		Signer: types.LatestSigner(sd.L2Cfg.Config),
	})

	balanceAt := func(a common.Address) *big.Int {
		t.Helper()
		bal, err := seqEngine.EthClient().BalanceAt(t.Ctx(), a, nil)
		require.NoError(t, err)
		return bal
	}
	aliceInitialBalance := balanceAt(dp.Addresses.Alice)

	sequencer.ActL2PipelineFull(t)

	// new L1 block, with new L2 chain
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// Send an L2 tx
	alice.ActResetTxOpts(t)
	alice.ActSetTxToAddr(&dp.Addresses.Bob)
	alice.ActMakeTx(t)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
	sequencer.ActL2EndBlock(t)

	receipt := alice.LastTxReceipt(t)

	// Check that the operator fee was applied
	// TODO: uncomment this once the op-geth is updated to include the operator fee
	// require.Equal(t, testOperatorFeeScalar, uint64(*receipt.OperatorFeeScalar))
	// require.Equal(t, testOperatorFeeConstant, *receipt.OperatorFeeConstant)

	l1FeeVaultBalance := balanceAt(predeploys.L1FeeVaultAddr)
	baseFeeVaultBalance := balanceAt(predeploys.BaseFeeVaultAddr)
	sequencerFeeVaultBalance := balanceAt(predeploys.SequencerFeeVaultAddr)
	aliceFinalBalance := balanceAt(dp.Addresses.Alice)

	require.True(t, aliceFinalBalance.Cmp(aliceInitialBalance) < 0, "Alice's balance should decrease")

	// Check that the operator fee sent to the vault is correct
	l2GasUsed := new(big.Int).Sub(new(big.Int).SetUint64(receipt.GasUsed), receipt.L1GasUsed)
	baseFee := seqEngine.l2Chain.CurrentBlock().BaseFee
	require.Equal(t,
		new(big.Int).Add(
			new(big.Int).Div(
				new(big.Int).Mul(
					l2GasUsed,
					new(big.Int).Add(
						new(big.Int).SetUint64(uint64(testOperatorFeeScalar)),
						new(big.Int).Mul(baseFee, new(big.Int).SetUint64(1e6)),
					),
				),
				new(big.Int).SetUint64(1e6),
			),
			new(big.Int).SetUint64(testOperatorFeeConstant),
		),
		baseFeeVaultBalance,
	)

	// Check that no Ether has been minted or burned
	// All vault balances are 0 at the beginning of the test
	finalTotalBalance := new(big.Int).Add(
		aliceFinalBalance,
		new(big.Int).Add(
			new(big.Int).Add(l1FeeVaultBalance, sequencerFeeVaultBalance),
			baseFeeVaultBalance,
		),
	)

	require.Equal(t, aliceInitialBalance, finalTotalBalance)
}
