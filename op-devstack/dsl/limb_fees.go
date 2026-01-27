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

type LimbFees struct {
	commonImpl
	l2Network *L2Network
}

type LimbFeesValidationResult struct {
	TransactionReceipt *types.Receipt
	L1Fee              *big.Int
	L2Fee              *big.Int
	BaseFee            *big.Int
	PriorityFee        *big.Int
	TotalFee           *big.Int
	VaultBalances      LimbVaultBalances
	WalletBalanceDiff  *big.Int
	TransferAmount     *big.Int
}

type LimbVaultBalances struct {
	BaseFeeVault   *big.Int
	L1FeeVault     *big.Int
	SequencerVault *big.Int
}

func NewLimbFees(t devtest.T, l2Network *L2Network) *LimbFees {
	return &LimbFees{
		commonImpl: commonFromT(t),
		l2Network:  l2Network,
	}
}

func (lf *LimbFees) ValidateTransaction(from *EOA, to *EOA, amount *big.Int) LimbFeesValidationResult {
	client := lf.l2Network.inner.L2ELNode(match.FirstL2EL).EthClient()

	startBalance := from.GetBalance()
	vaultsBefore := lf.getVaultBalances(client)

	tx := from.Transfer(to.Address(), eth.WeiBig(amount))
	receipt, err := tx.Included.Eval(lf.ctx)
	lf.require.NoError(err)
	lf.require.Equal(types.ReceiptStatusSuccessful, receipt.Status)

	// Get block info for base fee information
	blockInfo, err := client.InfoByHash(lf.ctx, receipt.BlockHash)
	lf.require.NoError(err)

	endBalance := from.GetBalance()
	vaultsAfter := lf.getVaultBalances(client)
	vaultIncreases := lf.calculateVaultIncreases(vaultsBefore, vaultsAfter)

	// Calculate receipt-based fees for validation
	// In Limb, L1 fee is converted to L2 gas
	baseFee := new(big.Int).Mul(blockInfo.BaseFee(), big.NewInt(int64(receipt.GasUsed)))
	totalFee := new(big.Int).Mul(receipt.EffectiveGasPrice, big.NewInt(int64(receipt.GasUsed)))
	priorityFee := new(big.Int).Sub(totalFee, baseFee)

	walletBalanceDiff := new(big.Int).Sub(startBalance.ToBig(), endBalance.ToBig())
	walletBalanceDiff.Sub(walletBalanceDiff, amount)

	// print out the fees
	lf.log.Info("receipt details", "gasUsed", receipt.GasUsed, "effectiveGasPrice", receipt.EffectiveGasPrice, "l1Fee", receipt.L1Fee)
	lf.log.Info("Limb fees", "baseFee", baseFee, "priorityFee", priorityFee, "totalFee", totalFee)
	lf.log.Info("Limb vault balances", "baseFeeVault", vaultIncreases.BaseFeeVault, "l1FeeVault", vaultIncreases.L1FeeVault, "sequencerVault", vaultIncreases.SequencerVault)
	lf.log.Info("Limb wallet balance diff", "walletBalanceDiff", walletBalanceDiff)
	lf.log.Info("Limb transfer amount", "transferAmount", amount)

	// Validate total balance first to ensure all fees are accounted for
	lf.validateTotalBalance(walletBalanceDiff, totalFee, vaultIncreases)

	// Then validate individual fee components
	lf.validateFeeDistribution(baseFee, priorityFee, vaultIncreases)
	lf.validateLimbFeatures(receipt)

	return LimbFeesValidationResult{
		TransactionReceipt: receipt,
		L1Fee:              receipt.L1Fee,
		L2Fee:              new(big.Int).Sub(totalFee, receipt.L1Fee),
		BaseFee:            baseFee,
		PriorityFee:        priorityFee,
		TotalFee:           totalFee,
		VaultBalances:      vaultIncreases,
		WalletBalanceDiff:  walletBalanceDiff,
		TransferAmount:     amount,
	}
}

func (lf *LimbFees) getVaultBalances(client apis.EthClient) LimbVaultBalances {
	baseFee := lf.getBalance(client, predeploys.BaseFeeVaultAddr)
	l1Fee := lf.getBalance(client, predeploys.L1FeeVaultAddr)
	sequencer := lf.getBalance(client, predeploys.SequencerFeeVaultAddr)

	return LimbVaultBalances{
		BaseFeeVault:   baseFee,
		L1FeeVault:     l1Fee,
		SequencerVault: sequencer,
	}
}

func (lf *LimbFees) getBalance(client apis.EthClient, addr common.Address) *big.Int {
	balance, err := client.BalanceAt(lf.ctx, addr, nil)
	lf.require.NoError(err)
	return balance
}

func (lf *LimbFees) calculateVaultIncreases(before, after LimbVaultBalances) LimbVaultBalances {
	return LimbVaultBalances{
		BaseFeeVault:   new(big.Int).Sub(after.BaseFeeVault, before.BaseFeeVault),
		L1FeeVault:     new(big.Int).Sub(after.L1FeeVault, before.L1FeeVault),
		SequencerVault: new(big.Int).Sub(after.SequencerVault, before.SequencerVault),
	}
}

func (lf *LimbFees) validateFeeDistribution(baseFee, priorityFee *big.Int, vaults LimbVaultBalances) {
	lf.require.True(baseFee.Sign() > 0, "Base fee must be positive")
	lf.require.True(priorityFee.Sign() >= 0, "Priority fee must be non-negative")

	// In Limb, L1 fee goes to BaseFeeVault, so L1FeeVault should have zero increase
	lf.require.Equal(0, vaults.L1FeeVault.Cmp(big.NewInt(0)),
		"L1FeeVault increase must be zero in Limb (L1 fee goes to BaseFeeVault)")

	lf.require.Equal(baseFee, vaults.BaseFeeVault,
		"BaseFeeVault must contain base fee in Limb")

	lf.require.Equal(priorityFee, vaults.SequencerVault,
		"Priority fee must match SequencerFeeVault increase")
}

func (lf *LimbFees) validateTotalBalance(walletDiff *big.Int, totalFee *big.Int, vaults LimbVaultBalances) {
	totalVaultIncrease := new(big.Int).Add(vaults.BaseFeeVault, vaults.SequencerVault)
	totalVaultIncrease.Add(totalVaultIncrease, vaults.L1FeeVault)

	lf.require.Equal(walletDiff, totalFee, "Wallet balance difference must equal total fees")
	lf.require.Equal(totalVaultIncrease, totalFee, "Total vault increases must equal total fees")
}

func (lf *LimbFees) validateLimbFeatures(receipt *types.Receipt) {
	lf.require.NotNil(receipt.L1Fee, "L1 fee should be present in Limb")
	lf.require.Greater(receipt.GasUsed, uint64(20000), "Gas used should be reasonable for transfer")
	lf.require.Greater(receipt.EffectiveGasPrice.Uint64(), uint64(0), "Effective gas price should be > 0")
}

func (lf *LimbFees) LogResults(result LimbFeesValidationResult) {
	lf.log.Info("Comprehensive Limb fees validation completed",
		"gasUsed", result.TransactionReceipt.GasUsed,
		"effectiveGasPrice", result.TransactionReceipt.EffectiveGasPrice,
		"l1Fee", result.L1Fee,
		"l2Fee", result.L2Fee,
		"baseFee", result.BaseFee,
		"priorityFee", result.PriorityFee,
		"totalFee", result.TotalFee,
		"baseFeeVault", result.VaultBalances.BaseFeeVault,
		"l1FeeVault", result.VaultBalances.L1FeeVault,
		"sequencerVault", result.VaultBalances.SequencerVault,
		"walletBalanceDiff", result.WalletBalanceDiff,
		"transferAmount", result.TransferAmount)
}
