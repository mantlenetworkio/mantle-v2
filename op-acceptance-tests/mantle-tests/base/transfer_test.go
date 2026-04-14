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
// Uses ValidateTransaction (not ValidateReceipt) to isolate fee measurement to
// a single transaction, reducing interference from other transactions that may
// share the same block — especially when reth is the sequencer, which batches
// more aggressively than geth. Note: vault snapshots still use "latest" block,
// so complete isolation is only guaranteed when the test is the sole tx source.
func TestTransfer(gt *testing.T) {
	// Create a test environment using op-devstack
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)

	// Create two L2 wallets
	alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)
	t.Logger().Info("L2 balances before transfer", "alice", alice.Address(), "aliceBalance", alice.GetBalance(), "bob", bob.Address(), "bobBalance", bob.GetBalance())

	depositAmount := eth.OneHundredthEther

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
		fees.ValidateTransaction(alice, bob, depositAmount.ToBig())
	} else {
		fees := dsl.NewLimbFees(t, sys.L2Chain)
		fees.ValidateTransaction(alice, bob, depositAmount.ToBig())
	}

	// Verify recipient received the transfer amount (ValidateTransaction only
	// asserts the sender side; this closes the recipient-side coverage gap).
	bob.VerifyBalanceExact(depositAmount)
}
