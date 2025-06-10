package metrics

import (
	"github.com/ethereum/go-ethereum/metrics"
)

var (
	GasOracleStats struct {
		// metrics for L1 base fee, L1 bas price
		// TokenRatioGauge token_ratio = eth_price / mnt_price
		TokenRatioGauge *metrics.GaugeFloat64
		// TokenRatioWithScalarGauge token_ratio = token_ratio * token_ratio_scalar
		TokenRatioWithScalarGauge *metrics.GaugeFloat64
		// TokenRatioOnchainGauge token_ratio on chain
		TokenRatioOnchainGauge *metrics.GaugeFloat64
		// L1BaseFeeGauge
		L1BaseFeeGauge *metrics.Gauge
		// FeeScalarGauge value to scale the fee up by
		FeeScalarGauge *metrics.Gauge
		// L1GasPriceGauge l1_base_fee + l1_priority_fee
		L1GasPriceGauge *metrics.Gauge
		// OperatorFeeConstantGauge
		OperatorFeeConstantGauge *metrics.Gauge
	}
)

func InitAndRegisterStats(r metrics.Registry) {
	metrics.Enable()

	GasOracleStats.TokenRatioGauge = metrics.NewRegisteredGaugeFloat64("token_ratio", r)
	GasOracleStats.TokenRatioWithScalarGauge = metrics.NewRegisteredGaugeFloat64("token_ratio_with_scalar", r)
	GasOracleStats.TokenRatioOnchainGauge = metrics.NewRegisteredGaugeFloat64("token_ratio_onchain", r)
	GasOracleStats.L1BaseFeeGauge = metrics.NewRegisteredGauge("l1_base_fee", r)
	GasOracleStats.FeeScalarGauge = metrics.NewRegisteredGauge("fee_scalar", r)
	GasOracleStats.L1GasPriceGauge = metrics.NewRegisteredGauge("l1_gas_price", r)
	GasOracleStats.OperatorFeeConstantGauge = metrics.NewRegisteredGauge("operator_fee_constant", r)
}
