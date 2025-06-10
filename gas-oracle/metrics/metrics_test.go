package metrics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitAndRegisterStats(t *testing.T) {
	InitAndRegisterStats(DefaultRegistry)

	GasOracleStats.L1BaseFeeGauge.Update(100)
	l1BaseFee := GasOracleStats.L1BaseFeeGauge.Snapshot().Value()
	require.Equal(t, int64(100), l1BaseFee)

	GasOracleStats.TokenRatioGauge.Update(4000)
	tokenRatio := GasOracleStats.TokenRatioGauge.Snapshot().Value()
	require.Equal(t, float64(4000), tokenRatio)

	GasOracleStats.FeeScalarGauge.Update(1500000)
	feeScalar := GasOracleStats.FeeScalarGauge.Snapshot().Value()
	require.Equal(t, int64(1500000), feeScalar)

	GasOracleStats.OperatorFeeConstantGauge.Update(1000000000)
	operatorFeeConstant := GasOracleStats.OperatorFeeConstantGauge.Snapshot().Value()
	require.Equal(t, int64(1000000000), operatorFeeConstant)
}
