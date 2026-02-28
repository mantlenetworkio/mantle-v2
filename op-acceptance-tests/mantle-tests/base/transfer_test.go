package base

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	txib "github.com/ethereum-optimism/optimism/op-service/txintent/bindings"
	"github.com/ethereum-optimism/optimism/op-service/txintent/contractio"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

// Transfer: simple L2 native transfer smoke test with balance delta assertions.
func TestTransfer(gt *testing.T) {
	// Create a test environment using op-devstack
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)

	// Create two L2 wallets
	alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
	aliceBalance := alice.GetBalance()
	bob := sys.Wallet.NewEOA(sys.L2EL)
	bobBalance := bob.GetBalance()
	t.Logger().Info("L2 balances before transfer", "alice", alice.Address(), "aliceBalance", aliceBalance, "bob", bob.Address(), "bobBalance", bobBalance)

	depositAmount := eth.OneHundredthEther
	bobAddr := bob.Address()
	receipt := alice.Transfer(bobAddr, depositAmount).Included.Value()
	bob.WaitForBalance(bobBalance.Add(depositAmount))

	if sys.L2Chain.IsMantleForkActive(forks.MantleArsia) {
		l2Client := sys.L2EL.Escape().EthClient()
		gpo := txib.NewGasPriceOracle(
			txib.WithClient(l2Client),
			txib.WithTo(predeploys.GasPriceOracleAddr),
			txib.WithTest(t),
		)
		tokenRatio, err := contractio.Read(gpo.TokenRatio(), t.Ctx())
		t.Require().NoError(err)

		fees := dsl.NewArsiaFees(t, sys.L2Chain, tokenRatio)
		fees.ValidateReceipt(receipt, depositAmount.ToBig())
	} else {
		fees := dsl.NewLimbFees(t, sys.L2Chain)
		fees.ValidateReceipt(receipt, depositAmount.ToBig())
	}
}
