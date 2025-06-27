package oracle

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/log"
)

const (
	// Constants for the new operator fee formula
	DefaultIntrinsicSp1GasPerTx     = uint64(3e6)
	DefaultIntrinsicSp1GasPerBlock  = uint64(15e6)
	DefaultSp1PricePerBGasInDollars = 0.58
	DefaultSp1GasScalar             = uint64(0.0025 * 1e6)
)

// OperatorFeeCalculator is designed to calculate the operator fee constant and scalar
// based on txCount and ETH price
type OperatorFeeCalculator struct {
	intrinsicSp1GasPerTx        uint64
	intrinsicSp1GasPerBlock     uint64
	sp1PricePerBGasInDollars    float64
	sp1GasScalar                uint64
	operatorFeeMarkupPercentage int64
}

func NewOperatorFeeCalculator(intrinsicSp1GasPerTx, intrinsicSp1GasPerBlock uint64, sp1PricePerBGasInDollars float64, sp1GasScalar uint64, operatorFeeMarkupPercentage int64) *OperatorFeeCalculator {
	if intrinsicSp1GasPerTx == 0 {
		intrinsicSp1GasPerTx = DefaultIntrinsicSp1GasPerTx
		log.Info("Using default intrinsic sp1 gas per tx", "value", intrinsicSp1GasPerTx)
	}
	log.Debug("Given intrinsic sp1 gas per tx", "value", intrinsicSp1GasPerTx)
	if intrinsicSp1GasPerBlock == 0 {
		intrinsicSp1GasPerBlock = DefaultIntrinsicSp1GasPerBlock
		log.Info("Using default intrinsic sp1 gas per block", "value", intrinsicSp1GasPerBlock)
	}
	log.Debug("Given intrinsic sp1 gas per block", "value", intrinsicSp1GasPerBlock)
	if sp1PricePerBGasInDollars <= 0 {
		sp1PricePerBGasInDollars = DefaultSp1PricePerBGasInDollars
		log.Info("Using default sp1 price per bgas in dollars", "value", sp1PricePerBGasInDollars)
	}
	log.Debug("Given sp1 price per bgas in dollars", "value", sp1PricePerBGasInDollars)
	if sp1GasScalar == 0 {
		sp1GasScalar = DefaultSp1GasScalar
		log.Info("Using default sp1 gas scalar", "value", sp1GasScalar)
	}
	log.Debug("Given sp1 gas scalar", "value", sp1GasScalar)
	if operatorFeeMarkupPercentage < -100 {
		operatorFeeMarkupPercentage = -100
		log.Info("Markup percentage is too small, truncated to -100", "value", -100)
	}
	log.Debug("Given operator fee markup percentage", "value", operatorFeeMarkupPercentage)

	return &OperatorFeeCalculator{
		intrinsicSp1GasPerTx:        intrinsicSp1GasPerTx,
		intrinsicSp1GasPerBlock:     intrinsicSp1GasPerBlock,
		sp1PricePerBGasInDollars:    sp1PricePerBGasInDollars,
		sp1GasScalar:                sp1GasScalar,
		operatorFeeMarkupPercentage: operatorFeeMarkupPercentage,
	}
}

// Formula:
//
//	operatorFeeConstant =
//		(intrinsicSp1GasPerTx + (intrinsicSp1GasPerBlock * 43200) / txCount) * (sp1PricePerBGasInDollars * 1e-9 / (ethPrice * 1e-18))
func (o *OperatorFeeCalculator) CalOperatorFeeConstant(txCount uint64, ethPrice float64) (*big.Int, error) {
	if ethPrice <= 0 {
		return nil, fmt.Errorf("ETH price is 0 or negative")
	}
	if txCount == 0 {
		return nil, fmt.Errorf("transaction count is 0")
	}

	// Step 1: Calculate apportioned intrinsic sp1 gas of block
	txCountFloat := new(big.Int).SetUint64(txCount)
	sp1GasPerDay := new(big.Int).Mul(new(big.Int).SetUint64(o.intrinsicSp1GasPerBlock), new(big.Int).SetUint64(43200))
	apportionedSp1GasPerTx := new(big.Int).Div(sp1GasPerDay, txCountFloat)

	// Step 2: Calculate total sp1 gas used per tx
	totalSp1GasPerTx := new(big.Int).Add(new(big.Int).SetUint64(o.intrinsicSp1GasPerTx), apportionedSp1GasPerTx)

	// Step 3: Calculate sp1 price per gas in wei
	sp1PricePerGasInWei := calSp1PricePerGasInWei(o.sp1PricePerBGasInDollars, ethPrice)

	// Step 4: Calculate final result
	result := new(big.Int).Mul(totalSp1GasPerTx, sp1PricePerGasInWei)
	result.Mul(result, new(big.Int).SetInt64(100+o.operatorFeeMarkupPercentage))
	result.Div(result, new(big.Int).SetInt64(100))

	log.Info("Calculated operator fee constant",
		"transaction_count", txCount,
		"eth_price", ethPrice,
		"operator_fee_constant", result.String())

	return result, nil
}

// Formula:
//
//	operatorFeeScalar = sp1GasScalar * (sp1PricePerBGasInDollars * 1e-9 / (ethPrice * 1e-18))
func (o *OperatorFeeCalculator) CalOperatorFeeScalar(ethPrice float64) (*big.Int, error) {
	if ethPrice <= 0 {
		return nil, fmt.Errorf("ETH price is 0 or negative")
	}

	// Step 1: Calculate sp1 price per gas in wei
	sp1PricePerGasInWei := calSp1PricePerGasInWei(o.sp1PricePerBGasInDollars, ethPrice)

	// Step 2: Calculate final result
	result := new(big.Int).Mul(new(big.Int).SetUint64(o.sp1GasScalar), sp1PricePerGasInWei)
	result.Mul(result, new(big.Int).SetInt64(100+o.operatorFeeMarkupPercentage))
	result.Div(result, new(big.Int).SetInt64(100))

	log.Info("Calculated operator fee scalar",
		"eth_price", ethPrice,
		"operator_fee_scalar", result.String())

	return result, nil
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
