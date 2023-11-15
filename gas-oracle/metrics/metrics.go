package metrics

import (
	"github.com/ethereum/go-ethereum/metrics"
)

var (
	GasOracleStats struct {
		// metrics for L1 base fee, L1 bas price, da fee
		// TokenRatioGauge token_ratio = eth_price / mnt_price
		TokenRatioGauge metrics.GaugeFloat64
		// L1BaseFeeGauge (l1_base_fee + l1_priority_fee) * token_ratio
		L1BaseFeeGauge metrics.Gauge
		// FeeScalarGauge value to scale the fee up by
		FeeScalarGauge metrics.Gauge
		// L1GasPriceGauge l1_base_fee + l1_priority_fee
		L1GasPriceGauge metrics.Gauge
	}
)

func InitAndRegisterStats(r metrics.Registry) {
	metrics.Enabled = true

	// stats for L1 base fee, L1 bas price, fee scalar
	GasOracleStats.TokenRatioGauge = metrics.NewRegisteredGaugeFloat64("token_ratio", r)
	GasOracleStats.L1BaseFeeGauge = metrics.NewRegisteredGauge("l1_base_fee", r)
	GasOracleStats.FeeScalarGauge = metrics.NewRegisteredGauge("fee_scalar", r)
	GasOracleStats.L1GasPriceGauge = metrics.NewRegisteredGauge("l1_gas_price", r)
}
