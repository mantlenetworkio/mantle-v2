package dsl

import (
	"math/big"

	"github.com/ethereum-optimism/optimism/op-devstack/stack/match"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/core/types"
)

// ValidateTransaction validates the transaction on Mantle and returns the validation result
func (ff *FjordFees) ValidateMantleTransaction(from *EOA, to *EOA, amount *big.Int, tokenRatio *big.Int) FjordFeesValidationResult {
	client := ff.l2Network.inner.L2ELNode(match.FirstL2EL).EthClient()

	startBalance := from.GetBalance()
	vaultsBefore := ff.getVaultBalances(client)
	coinbaseStartBalance := ff.getCoinbaseBalance(client)

	tx := from.Transfer(to.Address(), eth.WeiBig(amount))
	receipt, err := tx.Included.Eval(ff.ctx)
	ff.require.NoError(err)
	ff.require.Equal(types.ReceiptStatusSuccessful, receipt.Status)

	endBalance := from.GetBalance()
	vaultsAfter := ff.getVaultBalances(client)
	vaultIncreases := ff.calculateVaultIncreases(vaultsBefore, vaultsAfter)
	coinbaseEndBalance := ff.getCoinbaseBalance(client)
	coinbaseDiff := new(big.Int).Sub(coinbaseEndBalance, coinbaseStartBalance)

	l1Fee := big.NewInt(0)
	if receipt.L1Fee != nil {
		l1Fee = receipt.L1Fee
	}

	block, err := client.InfoByHash(ff.ctx, receipt.BlockHash)
	ff.require.NoError(err)

	baseFee := new(big.Int).Mul(block.BaseFee(), big.NewInt(int64(receipt.GasUsed)))
	totalGasFee := new(big.Int).Mul(receipt.EffectiveGasPrice, big.NewInt(int64(receipt.GasUsed)))
	priorityFee := new(big.Int).Sub(totalGasFee, baseFee)

	l2Fee := new(big.Int).Set(priorityFee)

	operatorFee := vaultIncreases.OperatorVault

	ff.validateVaultIncreaseFees(l2Fee, baseFee, priorityFee, l1Fee, operatorFee, coinbaseDiff, vaultsAfter, vaultsBefore)

	totalFee := new(big.Int).Add(l1Fee, l2Fee)
	totalFee.Add(totalFee, baseFee)
	totalFee.Add(totalFee, operatorFee)

	walletBalanceDiff := new(big.Int).Sub(startBalance.ToBig(), endBalance.ToBig())
	walletBalanceDiff.Sub(walletBalanceDiff, amount)

	fastLzSize, estimatedBrotliSize := ff.validateArsiaFeatures(receipt, l1Fee, tokenRatio)
	ff.validateFeeDistribution(l1Fee, baseFee, priorityFee, operatorFee, vaultIncreases)
	ff.validateTotalBalance(walletBalanceDiff, totalFee, vaultIncreases)

	return FjordFeesValidationResult{
		TransactionReceipt:  receipt,
		L1Fee:               l1Fee,
		L2Fee:               l2Fee,
		BaseFee:             baseFee,
		PriorityFee:         priorityFee,
		TotalFee:            totalFee,
		VaultBalances:       vaultIncreases,
		WalletBalanceDiff:   walletBalanceDiff,
		TransferAmount:      amount,
		FastLzSize:          fastLzSize,
		EstimatedBrotliSize: estimatedBrotliSize,
		OperatorFee:         operatorFee,
		CoinbaseDiff:        coinbaseDiff,
	}
}

// validateArsiaFeatures validates that the features of the Arsia transaction are correct
func (ff *FjordFees) validateArsiaFeatures(receipt *types.Receipt, l1Fee *big.Int, tokenRatio *big.Int) (uint64, *big.Int) {
	ff.require.NotNil(receipt.L1Fee, "L1 fee should be present in Fjord")
	ff.require.True(l1Fee.Cmp(big.NewInt(0)) > 0, "L1 fee should be greater than 0 in Fjord")

	client := ff.l2Network.inner.L2ELNode(match.FirstL2EL).EthClient()

	_, txs, err := client.InfoAndTxsByHash(ff.ctx, receipt.BlockHash)
	ff.require.NoError(err)

	var signedTx *types.Transaction
	for _, tx := range txs {
		if tx.Hash() == receipt.TxHash {
			signedTx = tx
			break
		}
	}
	ff.require.NotNil(signedTx, "should find the signed transaction")

	unsignedTx := types.NewTx(&types.DynamicFeeTx{
		Nonce:     signedTx.Nonce(),
		To:        signedTx.To(),
		Value:     signedTx.Value(),
		Gas:       signedTx.Gas(),
		GasFeeCap: signedTx.GasFeeCap(),
		GasTipCap: signedTx.GasTipCap(),
		Data:      signedTx.Data(),
	})

	txUnsigned, err := unsignedTx.MarshalBinary()
	ff.require.NoError(err)
	txSigned, err := signedTx.MarshalBinary()
	ff.require.NoError(err)

	fastLzSizeUnsigned := uint64(types.FlzCompressLen(txUnsigned) + 68) // overhead used by the original test
	fastLzSizeSigned := uint64(types.FlzCompressLen(txSigned))

	// Validate that FastLZ compression produces reasonable results
	ff.require.Greater(fastLzSizeUnsigned, uint64(0), "FastLZ size should be positive")
	ff.require.Greater(fastLzSizeSigned, uint64(0), "FastLZ size should be positive")

	txLenGPO := len(txUnsigned) + 68
	flzUpperBound := uint64(txLenGPO + txLenGPO/255 + 16)
	ff.require.LessOrEqual(fastLzSizeUnsigned, flzUpperBound, "Compressed size should not exceed upper bound")

	signedUpperBound := uint64(len(txSigned) + len(txSigned)/255 + 16)
	ff.require.LessOrEqual(fastLzSizeSigned, signedUpperBound, "Compressed size should not exceed upper bound")

	receiptL1Fee := receipt.L1Fee
	if receiptL1Fee == nil {
		ff.t.Logf("L1 fee is nil in receipt, skipping L1 fee validation")
		return fastLzSizeSigned, nil
	}

	expectedFee, err := CalculateFjordL1Cost(ff.ctx, client, signedTx.RollupCostData(), receipt.BlockHash)
	ff.require.NoError(err, "should calculate L1 fee")

	// Mantle L1 fee is multiplied by token ratio
	expectedFee = expectedFee.Mul(expectedFee, tokenRatio)

	ff.require.Equalf(expectedFee, receiptL1Fee, "Calculated L1 fee should match receipt L1 fee (expected=%s actual=%s)", expectedFee.String(), receiptL1Fee.String())

	ff.require.Equalf(expectedFee, receipt.L1Fee, "L1 fee in receipt must be correct (expected=%s actual=%s)", expectedFee.String(), receipt.L1Fee.String())

	return fastLzSizeSigned, expectedFee
}
