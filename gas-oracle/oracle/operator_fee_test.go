package oracle

import (
	"math/big"
	"testing"
)

func TestCalculateOperatorFeeConstant(t *testing.T) {
	calculator := NewOperatorFeeCalculator(DefaultIntrinsicSp1GasPerTx, DefaultIntrinsicSp1GasPerBlock, DefaultSp1PricePerBGasInDollars)

	baseTxCount := uint64(200000)
	baseEthPrice := 2500.0
	baseConstant, _ := calculator(baseTxCount, baseEthPrice)
	if baseConstant.Cmp(big.NewInt(1447680000000)) != 0 {
		t.Errorf("baseConstant: %s, expected: %s", baseConstant.String(), "1447680000000")
	}

	tests := []struct {
		name         string
		txCount      uint64
		ethPrice     float64
		expectError  bool
		expectedSign int
	}{
		{
			name:         "zero ETH price",
			txCount:      1000000,
			ethPrice:     0.0,
			expectError:  true,
			expectedSign: 0,
		},
		{
			name:         "negative ETH price",
			txCount:      1000000,
			ethPrice:     -100.0,
			expectError:  true,
			expectedSign: 0,
		},
		{
			name:         "zero tx count",
			txCount:      0,
			ethPrice:     baseEthPrice,
			expectError:  true,
			expectedSign: 0,
		},
		{
			name:         "ETH price x2",
			txCount:      baseTxCount,
			ethPrice:     2 * baseEthPrice,
			expectError:  false,
			expectedSign: -1,
		},
		{
			name:         "ETH price x0.5",
			txCount:      baseTxCount,
			ethPrice:     0.5 * baseEthPrice,
			expectError:  false,
			expectedSign: 1,
		},
		{
			name:         "Tx count x2",
			txCount:      2 * baseTxCount,
			ethPrice:     baseEthPrice,
			expectError:  false,
			expectedSign: -1,
		},
		{
			name:         "Tx count x0.5",
			txCount:      baseTxCount / 2,
			ethPrice:     baseEthPrice,
			expectError:  false,
			expectedSign: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculator(tt.txCount, tt.ethPrice)

			if tt.expectError {
				if err == nil {
					t.Errorf("calculateOperatorFeeConstant(%d,%f) expected error but got none",
						tt.txCount, tt.ethPrice)
				}
			} else {
				if err != nil {
					t.Errorf("calculateOperatorFeeConstant(%d,%f) unexpected error: %v",
						tt.txCount, tt.ethPrice, err)
				}
				if tt.expectedSign != result.Cmp(baseConstant) {
					t.Errorf("calculateOperatorFeeConstant(%d,%f) = %s, expected sign %d, baseConstant: %s",
						tt.txCount, tt.ethPrice, result.String(), tt.expectedSign, baseConstant.String())
				}
			}
		})
	}
}
