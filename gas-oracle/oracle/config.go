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

	"github.com/urfave/cli"
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
	OperatorFeeUpdateEnabled          bool
	OperatorFeeConstantUpdateInterval uint64
	OperatorFeeScalarUpdateInterval   uint64
	OperatorFeeSignificanceFactor     float64
	IntrinsicSp1GasPerTx              uint64
	IntrinsicSp1GasPerBlock           uint64
	Sp1PricePerBGasInDollars          float64
	Sp1GasScalar                      uint64
	// mantle explorer config
	BlockscoutExplorerURL string
	EtherscanExplorerURL  string
	EtherscanAPIKey       string
}

// NewConfig creates a new Config
func NewConfig(ctx *cli.Context) *Config {
	cfg := Config{}
	cfg.ethereumHttpUrl = ctx.GlobalString(flags.EthereumHttpUrlFlag.Name)
	cfg.layerTwoHttpUrl = ctx.GlobalString(flags.LayerTwoHttpUrlFlag.Name)
	addr := ctx.GlobalString(flags.GasPriceOracleAddressFlag.Name)
	cfg.gasPriceOracleAddress = common.HexToAddress(addr)
	cfg.tokenRatioCexURL = ctx.GlobalString(flags.TokenRatioCexURL.Name)
	cfg.tokenRatioDexURL = ctx.GlobalString(flags.TokenRatioDexURL.Name)
	cfg.tokenRatioUpdateFrequencySecond = ctx.GlobalUint64(flags.TokenRatioUpdateFrequencySecond.Name)
	cfg.tokenRatioEpochLengthSeconds = ctx.GlobalUint64(flags.TokenRatioEpochLengthSecondsFlag.Name)
	cfg.tokenRatioSignificanceFactor = ctx.GlobalFloat64(flags.TokenRatioSignificanceFactorFlag.Name)
	cfg.tokenRatioScalar = ctx.GlobalFloat64(flags.TokenRatioScalarFlag.Name)
	cfg.EnableHsm = ctx.GlobalBool(flags.EnableHsmFlag.Name)
	cfg.HsmAddress = ctx.GlobalString(flags.HsmAddressFlag.Name)
	cfg.HsmAPIName = ctx.GlobalString(flags.HsmAPINameFlag.Name)
	cfg.HsmCreden = ctx.GlobalString(flags.HsmCredenFlag.Name)

	if cfg.EnableHsm {
		log.Info("gasoracle", "enableHsm", cfg.EnableHsm,
			"hsmAddress", cfg.HsmAddress)
	} else {
		if ctx.GlobalIsSet(flags.PrivateKeyFlag.Name) {
			hex := ctx.GlobalString(flags.PrivateKeyFlag.Name)
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

	if ctx.GlobalIsSet(flags.L1ChainIDFlag.Name) {
		chainID := ctx.GlobalUint64(flags.L1ChainIDFlag.Name)
		cfg.l1ChainID = new(big.Int).SetUint64(chainID)
	}
	if ctx.GlobalIsSet(flags.L2ChainIDFlag.Name) {
		chainID := ctx.GlobalUint64(flags.L2ChainIDFlag.Name)
		cfg.l2ChainID = new(big.Int).SetUint64(chainID)
	}

	if ctx.GlobalIsSet(flags.TransactionGasPriceFlag.Name) {
		gasPrice := ctx.GlobalUint64(flags.TransactionGasPriceFlag.Name)
		cfg.gasPrice = new(big.Int).SetUint64(gasPrice)
	}

	if ctx.GlobalIsSet(flags.WaitForReceiptFlag.Name) {
		cfg.waitForReceipt = true
	}

	cfg.MetricsEnabled = ctx.GlobalBool(flags.MetricsEnabledFlag.Name)
	cfg.MetricsHTTP = ctx.GlobalString(flags.MetricsHTTPFlag.Name)
	cfg.MetricsPort = ctx.GlobalInt(flags.MetricsPortFlag.Name)

	if ctx.GlobalIsSet(flags.OperatorFeeUpdateEnabledFlag.Name) {
		cfg.OperatorFeeUpdateEnabled = ctx.GlobalBool(flags.OperatorFeeUpdateEnabledFlag.Name)
		cfg.OperatorFeeConstantUpdateInterval = ctx.GlobalUint64(flags.OperatorFeeConstantUpdateIntervalFlag.Name)
		cfg.OperatorFeeScalarUpdateInterval = ctx.GlobalUint64(flags.OperatorFeeScalarUpdateIntervalFlag.Name)
		cfg.OperatorFeeSignificanceFactor = ctx.GlobalFloat64(flags.OperatorFeeSignificanceFactorFlag.Name)
		cfg.IntrinsicSp1GasPerTx = ctx.GlobalUint64(flags.IntrinsicSp1GasPerTxFlag.Name)
		cfg.IntrinsicSp1GasPerBlock = ctx.GlobalUint64(flags.IntrinsicSp1GasPerBlockFlag.Name)
		cfg.Sp1PricePerBGasInDollars = ctx.GlobalFloat64(flags.Sp1PricePerBGasInDollarsFlag.Name)
		cfg.Sp1GasScalar = ctx.GlobalUint64(flags.Sp1GasScalarFlag.Name)
	}

	cfg.BlockscoutExplorerURL = ctx.GlobalString(flags.BlockscoutExplorerURLFlag.Name)
	cfg.EtherscanExplorerURL = ctx.GlobalString(flags.EtherscanExplorerURLFlag.Name)
	cfg.EtherscanAPIKey = ctx.GlobalString(flags.EtherscanAPIKeyFlag.Name)

	return &cfg
}
