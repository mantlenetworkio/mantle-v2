package oracle

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum-optimism/optimism/gas-oracle/flags"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"github.com/urfave/cli/v2"
)

// Config represents the configuration options for the gas oracle
type Config struct {
	l1ChainID                       *big.Int
	l2ChainID                       *big.Int
	ethereumHttpUrl                 string
	layerTwoHttpUrl                 string
	gasPriceOracleAddress           common.Address
	privateKey                      *ecdsa.PrivateKey
	gasPrice                        *big.Int
	waitForReceipt                  bool
	tokenRatioEpochLengthSeconds    uint64
	tokenRatioSignificanceFactor    float64
	tokenRatioScalar                float64
	tokenRatioCexURL                string
	tokenRatioDexURL                string
	tokenRatioUpdateFrequencySecond uint64
	// hsm config
	EnableHsm  bool
	HsmAPIName string
	HsmCreden  string
	HsmAddress string
	// Metrics config
	MetricsEnabled bool
	MetricsHTTP    string
	MetricsPort    int
	// operator fee config
	OperatorFeeUpdateEnabled      bool
	OperatorFeeMarkupPercentage   int64
	OperatorFeeUpdateInterval     uint64
	OperatorFeeSignificanceFactor float64
	IntrinsicSp1GasPerTx          uint64
	IntrinsicSp1GasPerBlock       uint64
	Sp1PricePerBGasInDollars      float64
	Sp1GasScalar                  uint64
	// mantle explorer config
	BlockscoutExplorerURL string
	EtherscanExplorerURL  string
	EtherscanAPIKey       string
}

// NewConfig creates a new Config
func NewConfig(ctx *cli.Context) *Config {
	cfg := Config{}
	cfg.ethereumHttpUrl = ctx.String(flags.EthereumHttpUrlFlag.Name)
	cfg.layerTwoHttpUrl = ctx.String(flags.LayerTwoHttpUrlFlag.Name)
	addr := ctx.String(flags.GasPriceOracleAddressFlag.Name)
	cfg.gasPriceOracleAddress = common.HexToAddress(addr)
	cfg.tokenRatioCexURL = ctx.String(flags.TokenRatioCexURL.Name)
	cfg.tokenRatioDexURL = ctx.String(flags.TokenRatioDexURL.Name)
	cfg.tokenRatioUpdateFrequencySecond = ctx.Uint64(flags.TokenRatioUpdateFrequencySecond.Name)
	cfg.tokenRatioEpochLengthSeconds = ctx.Uint64(flags.TokenRatioEpochLengthSecondsFlag.Name)
	cfg.tokenRatioSignificanceFactor = ctx.Float64(flags.TokenRatioSignificanceFactorFlag.Name)
	cfg.tokenRatioScalar = ctx.Float64(flags.TokenRatioScalarFlag.Name)
	cfg.EnableHsm = ctx.Bool(flags.EnableHsmFlag.Name)
	cfg.HsmAddress = ctx.String(flags.HsmAddressFlag.Name)
	cfg.HsmAPIName = ctx.String(flags.HsmAPINameFlag.Name)
	cfg.HsmCreden = ctx.String(flags.HsmCredenFlag.Name)

	if cfg.EnableHsm {
		log.Info("gasoracle", "enableHsm", cfg.EnableHsm,
			"hsmAddress", cfg.HsmAddress)
	} else {
		if ctx.IsSet(flags.PrivateKeyFlag.Name) {
			hex := ctx.String(flags.PrivateKeyFlag.Name)
			hex = strings.TrimPrefix(hex, "0x")
			key, err := crypto.HexToECDSA(hex)
			if err != nil {
				log.Error(fmt.Sprintf("Option %q: %v", flags.PrivateKeyFlag.Name, err))
			}
			cfg.privateKey = key
		} else {
			log.Crit("No private key configured")
		}
	}

	if ctx.IsSet(flags.L1ChainIDFlag.Name) {
		chainID := ctx.Uint64(flags.L1ChainIDFlag.Name)
		cfg.l1ChainID = new(big.Int).SetUint64(chainID)
	}
	if ctx.IsSet(flags.L2ChainIDFlag.Name) {
		chainID := ctx.Uint64(flags.L2ChainIDFlag.Name)
		cfg.l2ChainID = new(big.Int).SetUint64(chainID)
	}

	if ctx.IsSet(flags.TransactionGasPriceFlag.Name) {
		gasPrice := ctx.Uint64(flags.TransactionGasPriceFlag.Name)
		cfg.gasPrice = new(big.Int).SetUint64(gasPrice)
	}

	if ctx.IsSet(flags.WaitForReceiptFlag.Name) {
		cfg.waitForReceipt = true
	}

	cfg.MetricsEnabled = ctx.Bool(flags.MetricsEnabledFlag.Name)
	cfg.MetricsHTTP = ctx.String(flags.MetricsHTTPFlag.Name)
	cfg.MetricsPort = ctx.Int(flags.MetricsPortFlag.Name)

	if ctx.IsSet(flags.OperatorFeeUpdateEnabledFlag.Name) {
		cfg.OperatorFeeUpdateEnabled = ctx.Bool(flags.OperatorFeeUpdateEnabledFlag.Name)
		cfg.OperatorFeeMarkupPercentage = ctx.Int64(flags.OperatorFeeMarkupFlag.Name)
		cfg.OperatorFeeUpdateInterval = ctx.Uint64(flags.OperatorFeeUpdateIntervalFlag.Name)
		cfg.OperatorFeeSignificanceFactor = ctx.Float64(flags.OperatorFeeSignificanceFactorFlag.Name)
		cfg.IntrinsicSp1GasPerTx = ctx.Uint64(flags.IntrinsicSp1GasPerTxFlag.Name)
		cfg.IntrinsicSp1GasPerBlock = ctx.Uint64(flags.IntrinsicSp1GasPerBlockFlag.Name)
		cfg.Sp1PricePerBGasInDollars = ctx.Float64(flags.Sp1PricePerBGasInDollarsFlag.Name)
		cfg.Sp1GasScalar = ctx.Uint64(flags.Sp1GasScalarFlag.Name)
	}

	cfg.BlockscoutExplorerURL = ctx.String(flags.BlockscoutExplorerURLFlag.Name)
	cfg.EtherscanExplorerURL = ctx.String(flags.EtherscanExplorerURLFlag.Name)
	cfg.EtherscanAPIKey = ctx.String(flags.EtherscanAPIKeyFlag.Name)

	return &cfg
}
