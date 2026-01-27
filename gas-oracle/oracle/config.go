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
	L1ChainID                       *big.Int
	L2ChainID                       *big.Int
	EthereumHttpUrl                 string
	LayerTwoHttpUrl                 string
	GasPriceOracleAddress           common.Address
	PrivateKey                      *ecdsa.PrivateKey
	GasPrice                        *big.Int
	WaitForReceipt                  bool
	TokenRatioEpochLengthSeconds    uint64
	TokenRatioSignificanceFactor    float64
	TokenRatioScalar                float64
	TokenRatioCexURL                string
	TokenRatioDexURL                string
	TokenRatioUpdateFrequencySecond uint64
	// hsm config
	EnableHsm  bool
	HsmAPIName string
	HsmCreden  string
	HsmAddress string
	// Metrics config
	MetricsEnabled bool
	MetricsHTTP    string
	MetricsPort    int
}

// NewConfig creates a new Config
func NewConfig(ctx *cli.Context) *Config {
	cfg := Config{}
	cfg.EthereumHttpUrl = ctx.String(flags.EthereumHttpUrlFlag.Name)
	cfg.LayerTwoHttpUrl = ctx.String(flags.LayerTwoHttpUrlFlag.Name)
	addr := ctx.String(flags.GasPriceOracleAddressFlag.Name)
	cfg.GasPriceOracleAddress = common.HexToAddress(addr)
	cfg.TokenRatioCexURL = ctx.String(flags.TokenRatioCexURL.Name)
	cfg.TokenRatioDexURL = ctx.String(flags.TokenRatioDexURL.Name)
	cfg.TokenRatioUpdateFrequencySecond = ctx.Uint64(flags.TokenRatioUpdateFrequencySecond.Name)
	cfg.TokenRatioEpochLengthSeconds = ctx.Uint64(flags.TokenRatioEpochLengthSecondsFlag.Name)
	cfg.TokenRatioSignificanceFactor = ctx.Float64(flags.TokenRatioSignificanceFactorFlag.Name)
	cfg.TokenRatioScalar = ctx.Float64(flags.TokenRatioScalarFlag.Name)
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
			cfg.PrivateKey = key
		} else {
			log.Crit("No private key configured")
		}
	}

	if ctx.IsSet(flags.L1ChainIDFlag.Name) {
		chainID := ctx.Uint64(flags.L1ChainIDFlag.Name)
		cfg.L1ChainID = new(big.Int).SetUint64(chainID)
	}
	if ctx.IsSet(flags.L2ChainIDFlag.Name) {
		chainID := ctx.Uint64(flags.L2ChainIDFlag.Name)
		cfg.L2ChainID = new(big.Int).SetUint64(chainID)
	}

	if ctx.IsSet(flags.TransactionGasPriceFlag.Name) {
		gasPrice := ctx.Uint64(flags.TransactionGasPriceFlag.Name)
		cfg.GasPrice = new(big.Int).SetUint64(gasPrice)
	}

	if ctx.IsSet(flags.WaitForReceiptFlag.Name) {
		cfg.WaitForReceipt = true
	}

	cfg.MetricsEnabled = ctx.Bool(flags.MetricsEnabledFlag.Name)
	cfg.MetricsHTTP = ctx.String(flags.MetricsHTTPFlag.Name)
	cfg.MetricsPort = ctx.Int(flags.MetricsPortFlag.Name)

	return &cfg
}
