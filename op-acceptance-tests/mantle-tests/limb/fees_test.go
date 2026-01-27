package ecotone

import (
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

func TestFees(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMinimal(t)
	require := t.Require()

	// Limb is the last fork that validates the same fees as Ecotone.
	// And we're not going to support acceptance tests for forks before Limb.
	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleLimb), "Limb fork must be active for this test")
	if sys.L2Chain.IsMantleForkActive(forks.MantleArsia) {
		t.Skip("skipping since Arsia introduces a new fee component: OperatorFee")
	}

	alice := sys.FunderL2.NewFundedEOA(eth.OneEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)

	// In Limb, L1 fee is distributed to BaseFeeVault (not L1FeeVault like in Ecotone/Arsia)
	// And there is no OperatorFee in Limb.
	limbFees := dsl.NewLimbFees(t, sys.L2Chain)

	result := limbFees.ValidateTransaction(alice, bob, big.NewInt(42000000000))

	limbFees.LogResults(result)

	t.Log("Limb fees test completed successfully",
		"gasUsed", result.TransactionReceipt.GasUsed,
		"l1Fee", result.L1Fee.String(),
		"l2Fee", result.L2Fee.String(),
		"baseFee", result.BaseFee.String(),
		"priorityFee", result.PriorityFee.String(),
		"totalFee", result.TotalFee.String(),
		"walletBalanceDiff", result.WalletBalanceDiff.String())
}
