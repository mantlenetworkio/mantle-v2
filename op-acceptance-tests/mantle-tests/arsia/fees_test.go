package arsia

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	dsl "github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	txib "github.com/ethereum-optimism/optimism/op-service/txintent/bindings"
	"github.com/ethereum-optimism/optimism/op-service/txintent/contractio"
)

func TestFees(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))
	operatorFee := dsl.NewOperatorFee(t, sys.L2Chain, sys.L1EL)
	operatorFee.SetOperatorFee(100000000, 500)
	operatorFee.WaitForL2SyncWithCurrentL1State()

	alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)

	l2Client := sys.L2EL.Escape().EthClient()
	gpo := txib.NewGasPriceOracle(
		txib.WithClient(l2Client),
		txib.WithTo(predeploys.GasPriceOracleAddr),
		txib.WithTest(t),
	)
	tokenRatio, err := contractio.Read(gpo.TokenRatio(), ctx)
	require.NoError(err)

	fjordFees := dsl.NewFjordFees(t, sys.L2Chain)
	result := fjordFees.ValidateMantleTransaction(alice, bob, eth.OneHundredthEther.ToBig(), tokenRatio)

	signedTx, err := dsl.FindSignedTransactionFromReceipt(ctx, l2Client, result.TransactionReceipt)
	require.NoError(err)
	require.NotNil(signedTx)

	unsignedTx, err := dsl.CreateUnsignedTransactionFromSigned(signedTx)
	require.NoError(err)

	txUnsigned, err := unsignedTx.MarshalBinary()
	require.NoError(err)

	gpoL1Fee, err := dsl.ReadGasPriceOracleL1FeeAt(ctx, l2Client, gpo, txUnsigned, result.TransactionReceipt.BlockHash)
	require.NoError(err)
	// Mantle L1 fee is multiplied by token ratio
	gpoL1Fee = gpoL1Fee.Mul(gpoL1Fee, tokenRatio)
	dsl.ValidateL1FeeMatches(t, result.L1Fee, gpoL1Fee)
}
