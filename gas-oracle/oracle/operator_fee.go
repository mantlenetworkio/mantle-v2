package oracle

import (
	"fmt"
	"math/big"

	ometrics "github.com/ethereum-optimism/optimism/gas-oracle/metrics"
	"github.com/ethereum/go-ethereum/log"
)

const (
	// Constants for the new operator fee formula
	DefaultIntrinsicSp1GasPerTx     = uint64(3e6)
	DefaultIntrinsicSp1GasPerBlock  = uint64(15e6)
	OperatorFeeUpdateThreshold      = 0.05 // 5% threshold
	DefaultSp1PricePerBGasInDollars = 0.58
)

// OperatorFeeCalculator is a function type that calculates the operator fee constant
// based on txCount and ETH price
type OperatorFeeCalculator func(txCount uint64, ethPrice float64) (*big.Int, error)

// NewOperatorFeeCalculator creates a new operator fee calculator function
// Formula:
//
//	operatorFeeConstant =
//		(intrinsicSp1GasPerTx + (intrinsicSp1GasPerBlock * 43200) / txCount) * (sp1PricePerBGasInDollars * 1e-9 / (ethPrice * 1e-18))
func NewOperatorFeeCalculator(intrinsicSp1GasPerTx, intrinsicSp1GasPerBlock uint64, sp1PricePerBGasInDollars float64) OperatorFeeCalculator {
	if intrinsicSp1GasPerTx == 0 {
		intrinsicSp1GasPerTx = DefaultIntrinsicSp1GasPerTx
		log.Info("Using default intrinsic sp1 gas per tx", "value", intrinsicSp1GasPerTx)
	}
	if intrinsicSp1GasPerBlock == 0 {
		intrinsicSp1GasPerBlock = DefaultIntrinsicSp1GasPerBlock
		log.Info("Using default intrinsic sp1 gas per block", "value", intrinsicSp1GasPerBlock)
	}
	if sp1PricePerBGasInDollars <= 0 {
		sp1PricePerBGasInDollars = DefaultSp1PricePerBGasInDollars
		log.Info("Using default sp1 price per bgas in dollars", "value", sp1PricePerBGasInDollars)
	}

	return func(txCount uint64, ethPrice float64) (*big.Int, error) {
		if ethPrice <= 0 {
			return nil, fmt.Errorf("ETH price is 0 or negative")
		}
		if txCount == 0 {
			return nil, fmt.Errorf("transaction count is 0")
		}

		// Step 1: Calculate apportioned intrinsic sp1 gas of block
		txCountFloat := new(big.Int).SetUint64(txCount)
		sp1GasPerDay := new(big.Int).Mul(new(big.Int).SetUint64(intrinsicSp1GasPerBlock), new(big.Int).SetUint64(43200))
		apportionedSp1GasPerTx := new(big.Int).Div(sp1GasPerDay, txCountFloat)

		// Step 2: Calculate total sp1 gas used per tx
		totalSp1GasPerTx := new(big.Int).Add(new(big.Int).SetUint64(intrinsicSp1GasPerTx), apportionedSp1GasPerTx)

		// Step 3: Calculate sp1 price per gas in wei
		sp1PricePerGasInWei := calSp1PricePerGasInWei(sp1PricePerBGasInDollars, ethPrice)

		// Step 4: Calculate final result
		result := new(big.Int).Mul(totalSp1GasPerTx, sp1PricePerGasInWei)

		log.Info("Calculated operator fee constant",
			"transaction_count", txCount,
			"eth_price", ethPrice,
			"operator_fee_constant", result.String())

		return result, nil
	}
}

// we assume that the Sp1PricePerGasInWei is always a positive and significant value
// output big.Int instead of big.Float to have more human readable value
// truncate to 1 if the result is too small
func calSp1PricePerGasInWei(sp1PricePerBGasInDollars float64, ethPrice float64) *big.Int {
	sp1PricePerGasInWei := new(big.Float).Quo(new(big.Float).SetFloat64(sp1PricePerBGasInDollars*1e9), new(big.Float).SetFloat64(ethPrice))
	resultInt := new(big.Int)
	sp1PricePerGasInWei.Int(resultInt)
	if resultInt.Cmp(big.NewInt(0)) == 0 {
		log.Warn("sp1PricePerGasInWei is too small, truncated to 1")
		return big.NewInt(1)
	}
	return resultInt
}

// UpdateOperatorFeeConstant is the main function that orchestrates the operator fee update process
func (g *GasPriceOracle) updateOperatorFeeConstant() error {
	// Step 1: Get current ETH price from token ratio client
	ethPrice := g.tokenRatio.EthPrice()

	// Step 2: Fetch transaction count from the explorer client
	// DailyTxCountFromUser is the transaction count from the explorer client minus the daily block count
	txCount, err := g.explorerClient.DailyTxCountFromUser(g.ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch transaction count: %w", err)
	}

	// Step 3: Calculate new operator fee constant using the calculator function
	newConstant, err := g.operatorFeeCalculator(txCount, ethPrice)
	if err != nil {
		return fmt.Errorf("failed to calculate operator fee constant: %w", err)
	}

	// Step 4: Get cached operator fee constant
	currentConstant := g.lastOperatorFeeConstant

	log.Debug("Getting cached operator fee constant", "cached_value", currentConstant.String())

	// Step 5: Only update if the value has changed by more than the significance factor
	significanceFactor := g.config.OperatorFeeSignificanceFactor
	if significanceFactor <= 0 {
		significanceFactor = OperatorFeeUpdateThreshold
	}
	if isDifferenceSignificant(currentConstant.Uint64(), newConstant.Uint64(), significanceFactor) {
		log.Info("Updating operator fee constant - change exceeds threshold",
			"current", currentConstant.String(),
			"new", newConstant.String())

		// Update the cache with the new value
		g.lastOperatorFeeConstant = newConstant
		return g.setOperatorFeeConstant(newConstant)
	} else {
		log.Debug("Operator fee constant unchanged or change is below threshold, skipping update",
			"current_value", currentConstant.String())
	}

	return nil
}

// updateOperatorFeeConstantOnContract updates the operator fee constant on the smart contract
func (g *GasPriceOracle) setOperatorFeeConstant(newConstant *big.Int) error {
	// Send transaction to update operator fee constant
	tx, err := g.contract.SetOperatorFeeConstant(g.auth.Opts(), newConstant)
	if err != nil {
		return fmt.Errorf("failed to update operator fee constant: %w", err)
	}

	log.Info("Operator fee constant update transaction sent",
		"tx_hash", tx.Hash().Hex())
	ometrics.GasOracleStats.OperatorFeeConstantGauge.Update(newConstant.Int64())

	// Wait for receipt if configured
	if g.config.waitForReceipt {
		// Wait for the receipt
		receipt, err := waitForReceipt(g.l2Backend, tx)
		if err != nil {
			return err
		}
		log.Info("Operator fee constant update transaction confirmed",
			"tx_hash", tx.Hash().Hex(),
			"block_number", receipt.BlockNumber)
	}

	return nil
}
