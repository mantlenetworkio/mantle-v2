package dsl

import (
	"math/big"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txintent/contractio"
	"github.com/ethereum/go-ethereum/core/types"
)

func (of *OperatorFee) ValidateMantleTransactionFees(from *EOA, to *EOA, amount *big.Int, expectedScalar uint32, expectedConstant uint64) OperatorFeeValidationResult {
	// Ensure there is at least one user transaction, to trigger flow of operator fees to vault.
	tx := from.Transfer(to.Address(), eth.WeiBig(amount))
	receipt, err := tx.Included.Eval(of.ctx)
	of.require.NoError(err)
	of.require.Equal(types.ReceiptStatusSuccessful, receipt.Status)

	blockHash := receipt.BlockHash
	_, txs, err := from.el.stackEL().EthClient().InfoAndTxsByHash(of.ctx, blockHash)
	of.require.NoError(err)

	// Verify GPO upgraded when arsia is active
	isArsiainGPO, err := contractio.Read(of.gasPriceOracle.IsArsia(), of.ctx)
	of.require.NoError(err)
	of.require.True(isArsiainGPO)

	// Get updated balance in operator fee vault to compute delta
	vaultAfter, err := from.el.stackEL().EthClient().BalanceAt(of.ctx, predeploys.OperatorFeeVaultAddr, receipt.BlockNumber)
	of.require.NoError(err)
	vaultBefore, err := from.el.stackEL().EthClient().BalanceAt(of.ctx, predeploys.OperatorFeeVaultAddr, big.NewInt(0).Sub(receipt.BlockNumber, big.NewInt(1)))
	of.require.NoError(err)
	vaultIncrease := new(big.Int).Sub(vaultAfter, vaultBefore)

	// Loop through transactions in block to compute expected operator fee vault increase
	expectedOperatorFeeVaultIncrease := big.NewInt(0)
	if !(expectedScalar == 0 && expectedConstant == 0) {
		// The test submits one user transaction but we loop over all user transactions
		// to make the test robust to any other traffic on the chain.
		for _, tx := range txs {
			if tx.Type() == types.DepositTxType {
				continue
			}
			receipt, err := from.el.stackEL().EthClient().TransactionReceipt(of.ctx, tx.Hash())
			of.require.NoError(err)

			operatorFee := new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), big.NewInt(int64(expectedScalar)))
			// Arsia(Jovian) formula: (gasUsed * operatorFeeScalar * 100) + operatorFeeConstant
			operatorFee.Mul(operatorFee, big.NewInt(100))
			operatorFee.Add(operatorFee, big.NewInt(int64(expectedConstant)))
			expectedOperatorFeeVaultIncrease =
				expectedOperatorFeeVaultIncrease.Add(expectedOperatorFeeVaultIncrease, operatorFee)
		}
	}

	// Use Cmp for big.Int comparison to avoid representation issues
	of.require.Equal(0, expectedOperatorFeeVaultIncrease.Cmp(vaultIncrease),
		"operator fee vault balance mismatch: expected %s, got %s",
		expectedOperatorFeeVaultIncrease.String(), vaultIncrease.String())

	actualTotalFee := new(big.Int).Mul(receipt.EffectiveGasPrice, big.NewInt(int64(receipt.GasUsed)))
	if receipt.L1Fee != nil {
		actualTotalFee.Add(actualTotalFee, receipt.L1Fee)
	}

	if expectedScalar != 0 || expectedConstant != 0 {
		of.require.NotNil(receipt.OperatorFeeScalar)
		of.require.NotNil(receipt.OperatorFeeConstant)

		of.require.Equal(expectedScalar, uint32(*receipt.OperatorFeeScalar))
		of.require.Equal(expectedConstant, *receipt.OperatorFeeConstant)
	}

	return OperatorFeeValidationResult{
		TransactionReceipt:   receipt,
		ExpectedOperatorFee:  expectedOperatorFeeVaultIncrease,
		ActualTotalFee:       actualTotalFee,
		VaultBalanceIncrease: vaultIncrease,
	}
}

func RunMantleOperatorFeeTest(t devtest.T, l2Chain *L2Network, l1EL *L1ELNode, funderL1, funderL2 *Funder) {
	fundAmount := eth.OneTenthEther
	alice := funderL2.NewFundedEOA(fundAmount)
	alice.WaitForBalance(fundAmount)
	bob := funderL2.NewFundedEOA(eth.ZeroWei)

	operatorFee := NewOperatorFee(t, l2Chain, l1EL)
	operatorFee.CheckCompatibility()
	systemOwner := operatorFee.GetSystemOwner()
	funderL1.FundAtLeast(systemOwner, fundAmount)

	// First, ensure L2 is synced with current L1 state before starting tests
	t.Log("Ensuring L2 is synced with current L1 state...")
	operatorFee.WaitForL2SyncWithCurrentL1State()

	testCases := []struct {
		name     string
		scalar   uint32
		constant uint64
	}{
		{"ZeroFees", 0, 0},
		{"NonZeroFees", 300, 400},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t devtest.T) {
			operatorFee.SetOperatorFee(tc.scalar, tc.constant)
			operatorFee.WaitForL2Sync(tc.scalar, tc.constant)
			operatorFee.VerifyL2Config(tc.scalar, tc.constant)

			result := operatorFee.ValidateMantleTransactionFees(alice, bob, big.NewInt(1000), tc.scalar, tc.constant)

			t.Log("Test completed successfully:",
				"testCase", tc.name,
				"gasUsed", result.TransactionReceipt.GasUsed,
				"actualTotalFee", result.ActualTotalFee.String(),
				"expectedOperatorFee", result.ExpectedOperatorFee.String(),
				"vaultBalanceIncrease", result.VaultBalanceIncrease.String())
		})
	}

	operatorFee.RestoreOriginalConfig()
}
