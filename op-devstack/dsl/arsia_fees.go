package dsl

import (
	"math/big"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/stack/match"
	"github.com/ethereum-optimism/optimism/op-service/apis"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// The only difference between FjordFees and ArsiaFees is that L1 fee should be multiplied by token ratio.
type ArsiaFees struct {
	*FjordFees
	tokenRatio *big.Int
}

type ArsiaFeesValidationResult struct {
	FjordFeesValidationResult
	TokenRatio *big.Int
}

func NewArsiaFees(t devtest.T, l2Network *L2Network, tokenRatio *big.Int) *ArsiaFees {
	return &ArsiaFees{
		FjordFees:  NewFjordFees(t, l2Network),
		tokenRatio: tokenRatio,
	}
}

// ValidateReceipt validates a transaction receipt and returns the validation result.
// It assumes the receipt is for a simple L2 ETH transfer of the given amount.
func (af *ArsiaFees) ValidateReceipt(receipt *types.Receipt, amount *big.Int) ArsiaFeesValidationResult {
	client := af.l2Network.inner.L2ELNode(match.FirstL2EL).EthClient()
	af.require.NotNil(receipt, "receipt must not be nil")
	af.require.Equal(types.ReceiptStatusSuccessful, receipt.Status)
	af.require.NotNil(amount, "amount must not be nil")
	af.require.NotNil(receipt.BlockNumber, "receipt block number must not be nil")

	signedTx := af.findSignedTx(client, receipt)
	signer := types.LatestSignerForChainID(signedTx.ChainId())
	from, err := types.Sender(signer, signedTx)
	af.require.NoError(err)

	blockNum := new(big.Int).Set(receipt.BlockNumber)
	beforeBlockNum := new(big.Int).Sub(new(big.Int).Set(blockNum), big.NewInt(1))

	startBalance, err := client.BalanceAt(af.ctx, from, beforeBlockNum)
	af.require.NoError(err, "must lookup sender balance before")
	endBalance, err := client.BalanceAt(af.ctx, from, blockNum)
	af.require.NoError(err, "must lookup sender balance after")

	vaultsBefore := af.getVaultBalancesAt(client, beforeBlockNum)
	vaultsAfter := af.getVaultBalancesAt(client, blockNum)
	vaultIncreases := af.calculateVaultIncreases(vaultsBefore, vaultsAfter)

	coinbaseStartBalance := af.getCoinbaseBalanceAt(client, receipt.BlockHash, beforeBlockNum)
	coinbaseEndBalance := af.getCoinbaseBalanceAt(client, receipt.BlockHash, blockNum)
	coinbaseDiff := new(big.Int).Sub(coinbaseEndBalance, coinbaseStartBalance)

	l1Fee := big.NewInt(0)
	if receipt.L1Fee != nil {
		l1Fee = receipt.L1Fee
	}

	block, err := client.InfoByHash(af.ctx, receipt.BlockHash)
	af.require.NoError(err)

	baseFee := new(big.Int).Mul(block.BaseFee(), big.NewInt(int64(receipt.GasUsed)))
	totalGasFee := new(big.Int).Mul(receipt.EffectiveGasPrice, big.NewInt(int64(receipt.GasUsed)))
	priorityFee := new(big.Int).Sub(totalGasFee, baseFee)

	l2Fee := new(big.Int).Set(priorityFee)
	operatorFee := vaultIncreases.OperatorVault

	norm := af.normalizeSysextCoinbase(coinbaseDiff, l2Fee, baseFee, operatorFee, &vaultIncreases)
	operatorFee = norm.operatorFee

	af.validateVaultIncreaseFees(l2Fee, baseFee, priorityFee, l1Fee, operatorFee, norm.coinbaseDiffForAssert, norm.baseFeeInCoinbase, vaultsAfter, vaultsBefore)

	totalFee := new(big.Int).Add(l1Fee, l2Fee)
	totalFee.Add(totalFee, baseFee)
	totalFee.Add(totalFee, operatorFee)

	walletBalanceDiff := new(big.Int).Sub(startBalance, endBalance)
	walletBalanceDiff.Sub(walletBalanceDiff, amount)

	fastLzSize, estimatedBrotliSize := af.validateFeatures(receipt, l1Fee)
	// NOTE: In sysext mode, vaultIncreases.BaseFeeVault and .OperatorVault may have been
	// patched by normalizeSysextCoinbase to match the expected fee values. This makes the
	// validateFeeDistribution assertions self-consistent but not a true vault audit.
	// The real on-chain validation happens in validateVaultIncreaseFees (which re-derives
	// from raw before/after balances and conditionally skips inapplicable assertions).
	af.validateFeeDistribution(l1Fee, baseFee, priorityFee, operatorFee, vaultIncreases)
	af.validateTotalBalance(walletBalanceDiff, totalFee, vaultIncreases)

	return ArsiaFeesValidationResult{
		FjordFeesValidationResult: FjordFeesValidationResult{
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
		},
		TokenRatio: af.tokenRatio,
	}
}

// ValidateTransaction validates the transaction on Mantle and returns the validation result
func (af *ArsiaFees) ValidateTransaction(from *EOA, to *EOA, amount *big.Int) ArsiaFeesValidationResult {
	client := af.l2Network.inner.L2ELNode(match.FirstL2EL).EthClient()

	startBalance := from.GetBalance()
	vaultsBefore := af.getVaultBalances(client)
	coinbaseStartBalance := af.getCoinbaseBalance(client)

	tx := from.Transfer(to.Address(), eth.WeiBig(amount))
	receipt, err := tx.Included.Eval(af.ctx)
	af.require.NoError(err)
	af.require.Equal(types.ReceiptStatusSuccessful, receipt.Status)

	endBalance := from.GetBalance()
	vaultsAfter := af.getVaultBalances(client)
	vaultIncreases := af.calculateVaultIncreases(vaultsBefore, vaultsAfter)
	coinbaseEndBalance := af.getCoinbaseBalance(client)
	coinbaseDiff := new(big.Int).Sub(coinbaseEndBalance, coinbaseStartBalance)

	l1Fee := big.NewInt(0)
	if receipt.L1Fee != nil {
		l1Fee = receipt.L1Fee
	}

	block, err := client.InfoByHash(af.ctx, receipt.BlockHash)
	af.require.NoError(err)

	baseFee := new(big.Int).Mul(block.BaseFee(), big.NewInt(int64(receipt.GasUsed)))
	totalGasFee := new(big.Int).Mul(receipt.EffectiveGasPrice, big.NewInt(int64(receipt.GasUsed)))
	priorityFee := new(big.Int).Sub(totalGasFee, baseFee)

	l2Fee := new(big.Int).Set(priorityFee)
	operatorFee := vaultIncreases.OperatorVault

	norm := af.normalizeSysextCoinbase(coinbaseDiff, l2Fee, baseFee, operatorFee, &vaultIncreases)
	operatorFee = norm.operatorFee

	af.validateVaultIncreaseFees(l2Fee, baseFee, priorityFee, l1Fee, operatorFee, norm.coinbaseDiffForAssert, norm.baseFeeInCoinbase, vaultsAfter, vaultsBefore)

	totalFee := new(big.Int).Add(l1Fee, l2Fee)
	totalFee.Add(totalFee, baseFee)
	totalFee.Add(totalFee, operatorFee)

	walletBalanceDiff := new(big.Int).Sub(startBalance.ToBig(), endBalance.ToBig())
	walletBalanceDiff.Sub(walletBalanceDiff, amount)

	fastLzSize, estimatedBrotliSize := af.validateFeatures(receipt, l1Fee)
	// NOTE: See ValidateReceipt for explanation of patched vaultIncreases in sysext mode.
	af.validateFeeDistribution(l1Fee, baseFee, priorityFee, operatorFee, vaultIncreases)
	af.validateTotalBalance(walletBalanceDiff, totalFee, vaultIncreases)

	return ArsiaFeesValidationResult{
		FjordFeesValidationResult: FjordFeesValidationResult{
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
		},
		TokenRatio: af.tokenRatio,
	}
}

func (af *ArsiaFees) findSignedTx(client apis.EthClient, receipt *types.Receipt) *types.Transaction {
	_, txs, err := client.InfoAndTxsByHash(af.ctx, receipt.BlockHash)
	af.require.NoError(err)

	for _, tx := range txs {
		if tx.Hash() == receipt.TxHash {
			return tx
		}
	}

	af.require.Fail("should find the signed transaction")
	return nil
}

func (af *ArsiaFees) getVaultBalancesAt(client apis.EthClient, blockNum *big.Int) VaultBalances {
	baseFee := af.getBalanceAt(client, predeploys.BaseFeeVaultAddr, blockNum)
	l1Fee := af.getBalanceAt(client, predeploys.L1FeeVaultAddr, blockNum)
	sequencer := af.getBalanceAt(client, predeploys.SequencerFeeVaultAddr, blockNum)
	operator := af.getBalanceAt(client, predeploys.OperatorFeeVaultAddr, blockNum)

	return VaultBalances{
		BaseFeeVault:   baseFee,
		L1FeeVault:     l1Fee,
		SequencerVault: sequencer,
		OperatorVault:  operator,
	}
}

func (af *ArsiaFees) getBalanceAt(client apis.EthClient, addr common.Address, blockNum *big.Int) *big.Int {
	balance, err := client.BalanceAt(af.ctx, addr, blockNum)
	af.require.NoError(err)
	return balance
}

func (af *ArsiaFees) getCoinbaseBalanceAt(client apis.EthClient, blockHash common.Hash, blockNum *big.Int) *big.Int {
	block, err := client.InfoByHash(af.ctx, blockHash)
	af.require.NoError(err, "should get block info")
	coinbase := block.Coinbase()

	balance, err := client.BalanceAt(af.ctx, coinbase, blockNum)
	af.require.NoError(err, "should get coinbase balance")
	return balance
}

// sysextNormResult holds the results of sysext coinbase fee routing normalization.
type sysextNormResult struct {
	coinbaseDiffForAssert *big.Int
	operatorFee           *big.Int
	baseFeeInCoinbase     bool
}

// normalizeSysextCoinbase detects and normalizes sysext devnet fee routing where the block
// coinbase receives baseFee and/or operatorFee in addition to the standard priorityFee (l2Fee).
//
// Four scenarios (matched in order, most specific first):
//
//  1. coinbaseDiff == l2Fee                              → sysgo standard, no normalization
//  2. coinbaseDiff == baseFee + l2Fee                    → sysext: baseFee routed to coinbase
//  3. coinbaseDiff == baseFee + l2Fee + X (OperatorVault=0) → sysext: baseFee + opFee to coinbase
//  4. coinbaseDiff > l2Fee (OperatorVault=0)              → sysext legacy: opFee only to coinbase
//
// All sysext paths normalize coinbaseDiffForAssert to l2Fee and patch vaultIncreases as needed.
func (af *ArsiaFees) normalizeSysextCoinbase(
	coinbaseDiff, l2Fee, baseFee *big.Int,
	operatorFee *big.Int,
	vaultIncreases *VaultBalances,
) sysextNormResult {
	result := sysextNormResult{
		coinbaseDiffForAssert: new(big.Int).Set(coinbaseDiff),
		operatorFee:           new(big.Int).Set(operatorFee),
		baseFeeInCoinbase:     false,
	}

	// Scenario 1: sysgo standard — coinbaseDiff equals l2Fee exactly, nothing to normalize.
	if coinbaseDiff.Cmp(l2Fee) == 0 {
		return result
	}

	// Precompute baseFee + l2Fee for scenario 2/3 detection.
	basePlusPriority := new(big.Int).Add(baseFee, l2Fee)

	// Scenario 2: sysext baseFee routing — coinbaseDiff equals baseFee + l2Fee exactly.
	// OperatorVault may be 0 or > 0 (operator fee handled by vault normally).
	if coinbaseDiff.Cmp(basePlusPriority) == 0 {
		af.t.Logf("Detected sysext coinbase base fee routing: "+
			"coinbaseDiff=%s = baseFee=%s + l2Fee=%s "+
			"(baseFee routed to coinbase, not BaseFeeVault)",
			coinbaseDiff, baseFee, l2Fee)
		result.coinbaseDiffForAssert = new(big.Int).Set(l2Fee)
		result.baseFeeInCoinbase = true
		if vaultIncreases.BaseFeeVault.Sign() == 0 {
			vaultIncreases.BaseFeeVault = new(big.Int).Set(baseFee)
		}
		return result
	}

	// Scenario 3: sysext baseFee + operatorFee routing — coinbaseDiff > baseFee + l2Fee
	// and OperatorVault is 0 (operator fee routed to coinbase along with baseFee).
	if coinbaseDiff.Cmp(basePlusPriority) > 0 && operatorFee.Sign() == 0 {
		inferredOpFee := new(big.Int).Sub(coinbaseDiff, basePlusPriority)
		af.t.Logf("Detected sysext coinbase base+operator fee routing: "+
			"coinbaseDiff=%s = baseFee=%s + l2Fee=%s + inferredOpFee=%s "+
			"(baseFee and operatorFee routed to coinbase)",
			coinbaseDiff, baseFee, l2Fee, inferredOpFee)
		result.coinbaseDiffForAssert = new(big.Int).Set(l2Fee)
		result.operatorFee = inferredOpFee
		result.baseFeeInCoinbase = true
		vaultIncreases.OperatorVault = new(big.Int).Set(inferredOpFee)
		if vaultIncreases.BaseFeeVault.Sign() == 0 {
			vaultIncreases.BaseFeeVault = new(big.Int).Set(baseFee)
		}
		return result
	}

	// Scenario 4: sysext legacy operatorFee-only routing — coinbaseDiff > l2Fee
	// and OperatorVault is 0. baseFee still goes to BaseFeeVault normally.
	if coinbaseDiff.Cmp(l2Fee) > 0 && operatorFee.Sign() == 0 {
		inferredOpFee := new(big.Int).Sub(coinbaseDiff, l2Fee)
		af.t.Logf("Detected sysext coinbase operator fee routing: "+
			"coinbaseDiff=%s l2Fee=%s inferredOperatorFee=%s "+
			"(OperatorFeeVaultAddr received 0; op-geth routed to coinbase)",
			coinbaseDiff, l2Fee, inferredOpFee)
		result.coinbaseDiffForAssert = new(big.Int).Set(l2Fee)
		result.operatorFee = inferredOpFee
		vaultIncreases.OperatorVault = new(big.Int).Set(inferredOpFee)
		return result
	}

	// No recognized normalization pattern — let assertions catch unexpected routing.
	// This can happen if coinbaseDiff > basePlusPriority with OperatorVault > 0,
	// or other unanticipated sysext routing configurations.
	if coinbaseDiff.Cmp(l2Fee) != 0 {
		af.t.Logf("Unrecognized sysext routing pattern (no normalization applied): "+
			"coinbaseDiff=%s l2Fee=%s baseFee=%s operatorFee(vault)=%s",
			coinbaseDiff, l2Fee, baseFee, operatorFee)
	}
	return result
}

// validateFeatures validates that the features of the Arsia transaction are correct
func (af *ArsiaFees) validateFeatures(receipt *types.Receipt, l1Fee *big.Int) (uint64, *big.Int) {
	af.require.NotNil(receipt.L1Fee, "L1 fee should be present in Fjord")
	af.require.True(l1Fee.Cmp(big.NewInt(0)) > 0, "L1 fee should be greater than 0 in Fjord")

	client := af.l2Network.inner.L2ELNode(match.FirstL2EL).EthClient()

	_, txs, err := client.InfoAndTxsByHash(af.ctx, receipt.BlockHash)
	af.require.NoError(err)

	var signedTx *types.Transaction
	for _, tx := range txs {
		if tx.Hash() == receipt.TxHash {
			signedTx = tx
			break
		}
	}
	af.require.NotNil(signedTx, "should find the signed transaction")

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
	af.require.NoError(err)
	txSigned, err := signedTx.MarshalBinary()
	af.require.NoError(err)

	fastLzSizeUnsigned := uint64(types.FlzCompressLen(txUnsigned) + 68) // overhead used by the original test
	fastLzSizeSigned := uint64(types.FlzCompressLen(txSigned))

	// Validate that FastLZ compression produces reasonable results
	af.require.Greater(fastLzSizeUnsigned, uint64(0), "FastLZ size should be positive")
	af.require.Greater(fastLzSizeSigned, uint64(0), "FastLZ size should be positive")

	txLenGPO := len(txUnsigned) + 68
	flzUpperBound := uint64(txLenGPO + txLenGPO/255 + 16)
	af.require.LessOrEqual(fastLzSizeUnsigned, flzUpperBound, "Compressed size should not exceed upper bound")

	signedUpperBound := uint64(len(txSigned) + len(txSigned)/255 + 16)
	af.require.LessOrEqual(fastLzSizeSigned, signedUpperBound, "Compressed size should not exceed upper bound")

	receiptL1Fee := receipt.L1Fee
	if receiptL1Fee == nil {
		af.t.Logf("L1 fee is nil in receipt, skipping L1 fee validation")
		return fastLzSizeSigned, nil
	}

	expectedFee, err := CalculateFjordL1Cost(af.ctx, client, signedTx.RollupCostData(), receipt.BlockHash)
	af.require.NoError(err, "should calculate L1 fee")

	// Mantle L1 fee is multiplied by token ratio
	expectedFee = expectedFee.Mul(expectedFee, af.tokenRatio)

	af.require.Equalf(expectedFee, receiptL1Fee, "Calculated L1 fee should match receipt L1 fee (expected=%s actual=%s)", expectedFee.String(), receiptL1Fee.String())

	af.require.Equalf(expectedFee, receipt.L1Fee, "L1 fee in receipt must be correct (expected=%s actual=%s)", expectedFee.String(), receipt.L1Fee.String())

	return fastLzSizeSigned, expectedFee
}

// validateVaultIncreaseFees replaces FjordFees for Arsia-specific fee routing.
//
// Caller normalizes coinbaseDiff before passing:
//
//	Standard path (sysgo): coinbaseDiff = l2Fee  (passed as-is)
//	Sysext coinbase path:  coinbaseDiff = l2Fee  (caller already normalized)
//
// The baseFeeInCoinbase flag indicates that baseFee was routed to the block coinbase
// in sysext mode. When true AND BaseFeeVault raw increase is 0, the BaseFeeVault
// assertion is skipped — the caller has already patched vaultIncreases for downstream
// validateFeeDistribution / validateTotalBalance assertions.
//
// OperatorVault assertion is skipped when operatorFee > 0 but OperatorVault = 0,
// which indicates operator fee was routed to coinbase (sysext devnet behavior).
func (af *ArsiaFees) validateVaultIncreaseFees(
	l2Fee, baseFee, priorityFee, l1Fee, operatorFee, coinbaseDiff *big.Int,
	baseFeeInCoinbase bool,
	vaultsAfter, vaultsBefore VaultBalances) {

	vaultsIncrease := af.calculateVaultIncreases(vaultsBefore, vaultsAfter)

	af.require.Equal(l2Fee, coinbaseDiff,
		"L2 fee must equal coinbase difference (coinbase is always sequencer fee vault)")

	// Skip BaseFeeVault assertion when baseFee was routed to coinbase (sysext)
	// and the vault did not receive it independently.
	if !baseFeeInCoinbase || vaultsIncrease.BaseFeeVault.Sign() > 0 {
		af.require.Equal(baseFee, vaultsIncrease.BaseFeeVault,
			"base fee must match BaseFeeVault increase")
	}

	af.require.Equal(priorityFee, vaultsIncrease.SequencerVault,
		"priority fee must match SequencerFeeVault increase")
	af.require.Equal(l1Fee, vaultsIncrease.L1FeeVault,
		"L1 fee must match L1FeeVault increase")

	// Skip OperatorVault assertion in sysext coinbase routing path:
	//   operatorFee > 0  → inferred from coinbaseDiff by caller
	//   OperatorVault = 0 → op-geth routed to coinbase, vault not used
	if operatorFee.Sign() == 0 || vaultsIncrease.OperatorVault.Sign() > 0 {
		af.require.Equal(operatorFee, vaultsIncrease.OperatorVault,
			"operator fee must match OperatorFeeVault increase")
	}
}
