package oracle

import (
	"math/big"
	"testing"
)

func TestCalculateOperatorFeeConstant(t *testing.T) {
	calculator := NewOperatorFeeCalculator(DefaultIntrinsicSp1GasPerTx, DefaultIntrinsicSp1GasPerBlock, DefaultSp1PricePerBGasInDollars, DefaultSp1GasScalar, 0)

	baseTxCount := uint64(200000)
	baseEthPrice := 2500.0
	baseConstant, _ := calculator.CalOperatorFeeConstant(baseTxCount, baseEthPrice)
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
			result, err := calculator.CalOperatorFeeConstant(tt.txCount, tt.ethPrice)

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

func TestCalculateOperatorFeeScalar(t *testing.T) {
	calculator := NewOperatorFeeCalculator(DefaultIntrinsicSp1GasPerTx, DefaultIntrinsicSp1GasPerBlock, DefaultSp1PricePerBGasInDollars, DefaultSp1GasScalar, 0)

	baseEthPrice := 2500.0
	baseScalar, _ := calculator.CalOperatorFeeScalar(baseEthPrice)
	if baseScalar.Cmp(big.NewInt(580000000)) != 0 {
		t.Errorf("baseScalar: %s, expected: %s", baseScalar.String(), "580000000")
	}

	tests := []struct {
		name         string
		ethPrice     float64
		expectError  bool
		expectedSign int
	}{
		{
			name:         "zero ETH price",
			ethPrice:     0.0,
			expectError:  true,
			expectedSign: 0,
		},
		{
			name:         "negative ETH price",
			ethPrice:     -100.0,
			expectError:  true,
			expectedSign: 0,
		},
		{
			name:         "ETH price x2",
			ethPrice:     2 * baseEthPrice,
			expectError:  false,
			expectedSign: -1,
		},
		{
			name:         "ETH price x0.5",
			ethPrice:     0.5 * baseEthPrice,
			expectError:  false,
			expectedSign: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculator.CalOperatorFeeScalar(tt.ethPrice)

			if tt.expectError {
				if err == nil {
					t.Errorf("calculateOperatorFeeScalar(%f) expected error but got none", tt.ethPrice)
				}
			} else {
				if err != nil {
					t.Errorf("calculateOperatorFeeScalar(%f) unexpected error: %v", tt.ethPrice, err)
				}
				if tt.expectedSign != result.Cmp(baseScalar) {
					t.Errorf("calculateOperatorFeeScalar(%f) = %s, expected sign %d, baseScalar: %s",
						tt.ethPrice, result.String(), tt.expectedSign, baseScalar.String())
				}
			}
		})
	}
}

func TestCalculateOperatorFeeConstantWithMarkup(t *testing.T) {
	baseTxCount := uint64(200000)
	baseEthPrice := 2500.0
	baseConstant, _ := NewOperatorFeeCalculator(DefaultIntrinsicSp1GasPerTx, DefaultIntrinsicSp1GasPerBlock, DefaultSp1PricePerBGasInDollars, DefaultSp1GasScalar, 0).CalOperatorFeeConstant(baseTxCount, baseEthPrice)

	tests := []struct {
		name           string
		markup         int64
		expectError    bool
		expectedResult *big.Int
	}{
		{
			name:           "markup 100",
			markup:         100,
			expectError:    false,
			expectedResult: new(big.Int).Mul(baseConstant, big.NewInt(2)),
		},
		{
			name:           "markup -50	",
			markup:         -50,
			expectError:    false,
			expectedResult: new(big.Int).Div(baseConstant, big.NewInt(2)),
		},
		{
			name:           "markup -100",
			markup:         -100,
			expectError:    false,
			expectedResult: new(big.Int),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculator := NewOperatorFeeCalculator(DefaultIntrinsicSp1GasPerTx, DefaultIntrinsicSp1GasPerBlock, DefaultSp1PricePerBGasInDollars, DefaultSp1GasScalar, tt.markup)
			result, err := calculator.CalOperatorFeeConstant(baseTxCount, baseEthPrice)
			if tt.expectError {
				if err == nil {
					t.Errorf("calculateOperatorFeeConstantWithMarkup(%d) expected error but got none",
						tt.markup)
				}
			} else {
				if err != nil {
					t.Errorf("calculateOperatorFeeConstantWithMarkup(%d) unexpected error: %v",
						tt.markup, err)
				}
				if result.Cmp(tt.expectedResult) != 0 {
					t.Errorf("calculateOperatorFeeConstantWithMarkup(%d) = %s, expected %s",
						tt.markup, result.String(), tt.expectedResult.String())
				}
			}
		})
	}
}
