package txmgr

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/urfave/cli/v2"

	opservice "github.com/ethereum-optimism/optimism/op-service"
	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-signer/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

const (
	// Duplicated L1 RPC flag
	L1RPCFlagName = "l1-eth-rpc"
	// Key Management Flags (also have op-signer client flags)
	MnemonicFlagName   = "mnemonic"
	HDPathFlagName     = "hd-path"
	PrivateKeyFlagName = "private-key"
	// TxMgr Flags (new + legacy + some shared flags)
	NumConfirmationsFlagName          = "num-confirmations"
	SafeAbortNonceTooLowCountFlagName = "safe-abort-nonce-too-low-count"
	FeeLimitMultiplierFlagName        = "fee-limit-multiplier"
	FeeLimitThresholdFlagName         = "txmgr.fee-limit-threshold"
	ResubmissionTimeoutFlagName       = "resubmission-timeout"
	NetworkTimeoutFlagName            = "network-timeout"
	TxSendTimeoutFlagName             = "txmgr.send-timeout"
	TxNotInMempoolTimeoutFlagName     = "txmgr.not-in-mempool-timeout"
	ReceiptQueryIntervalFlagName      = "txmgr.receipt-query-interval"
	EnableHsmFlagName                 = "enable-hsm"
	HsmCredenFlagName                 = "hsm-creden"
	HsmAddressFlagName                = "hsm-address"
	HsmAPINameFlagName                = "hsm-api-name"
)

var (
	SequencerHDPathFlag = &cli.StringFlag{
		Name: "sequencer-hd-path",
		Usage: "DEPRECATED: The HD path used to derive the sequencer wallet from the " +
			"mnemonic. The mnemonic flag must also be set.",
		EnvVars: []string{"OP_BATCHER_SEQUENCER_HD_PATH"},
	}
	L2OutputHDPathFlag = &cli.StringFlag{
		Name: "l2-output-hd-path",
		Usage: "DEPRECATED:The HD path used to derive the l2output wallet from the " +
			"mnemonic. The mnemonic flag must also be set.",
		EnvVars: []string{"OP_PROPOSER_L2_OUTPUT_HD_PATH"},
	}
)

func CLIFlags(envPrefix string) []cli.Flag {
	return append([]cli.Flag{
		&cli.StringFlag{
			Name:    MnemonicFlagName,
			Usage:   "The mnemonic used to derive the wallets for either the service",
			EnvVars: opservice.PrefixEnvVar(envPrefix, "MNEMONIC"),
		},
		&cli.StringFlag{
			Name:    HDPathFlagName,
			Usage:   "The HD path used to derive the sequencer wallet from the mnemonic. The mnemonic flag must also be set.",
			EnvVars: opservice.PrefixEnvVar(envPrefix, "HD_PATH"),
		},
		&cli.StringFlag{
			Name:    PrivateKeyFlagName,
			Usage:   "The private key to use with the service. Must not be used with mnemonic.",
			EnvVars: opservice.PrefixEnvVar(envPrefix, "PRIVATE_KEY"),
		},
		&cli.Uint64Flag{
			Name:    NumConfirmationsFlagName,
			Usage:   "Number of confirmations which we will wait after sending a transaction",
			Value:   10,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "NUM_CONFIRMATIONS"),
		},
		&cli.Uint64Flag{
			Name:    SafeAbortNonceTooLowCountFlagName,
			Usage:   "Number of ErrNonceTooLow observations required to give up on a tx at a particular nonce without receiving confirmation",
			Value:   3,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "SAFE_ABORT_NONCE_TOO_LOW_COUNT"),
		},
		&cli.Uint64Flag{
			Name:    FeeLimitMultiplierFlagName,
			Usage:   "The multiplier applied to fee suggestions to put a hard limit on fee increases",
			Value:   5,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "TXMGR_FEE_LIMIT_MULTIPLIER"),
		},
		&cli.Float64Flag{
			Name:    FeeLimitThresholdFlagName,
			Usage:   "The minimum threshold (in GWei) at which fee bumping starts to be capped. Allows arbitrary fee bumps below this threshold.",
			Value:   100.0,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "TXMGR_FEE_LIMIT_THRESHOLD"),
		},
		&cli.DurationFlag{
			Name:    ResubmissionTimeoutFlagName,
			Usage:   "Duration we will wait before resubmitting a transaction to L1",
			Value:   48 * time.Second,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "RESUBMISSION_TIMEOUT"),
		},
		&cli.DurationFlag{
			Name:    NetworkTimeoutFlagName,
			Usage:   "Timeout for all network operations",
			Value:   2 * time.Second,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "NETWORK_TIMEOUT"),
		},
		&cli.DurationFlag{
			Name:    TxSendTimeoutFlagName,
			Usage:   "Timeout for sending transactions. If 0 it is disabled.",
			Value:   2 * time.Minute,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "TXMGR_TX_SEND_TIMEOUT"),
		},
		&cli.DurationFlag{
			Name:    TxNotInMempoolTimeoutFlagName,
			Usage:   "Timeout for aborting a tx send if the tx does not make it to the mempool.",
			Value:   2 * time.Minute,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "TXMGR_TX_NOT_IN_MEMPOOL_TIMEOUT"),
		},
		&cli.DurationFlag{
			Name:    ReceiptQueryIntervalFlagName,
			Usage:   "Frequency to poll for receipts",
			Value:   12 * time.Second,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "TXMGR_RECEIPT_QUERY_INTERVAL"),
		},
		&cli.BoolFlag{
			Name:    EnableHsmFlagName,
			Usage:   "Whether or not to use cloud hsm",
			EnvVars: opservice.PrefixEnvVar(envPrefix, "ENABLE_HSM"),
		},
		&cli.StringFlag{
			Name:    HsmAddressFlagName,
			Usage:   "The address of private-key in hsm",
			EnvVars: opservice.PrefixEnvVar(envPrefix, "HSM_ADDRESS"),
			Value:   "",
		},
		&cli.StringFlag{
			Name:    HsmAPINameFlagName,
			Usage:   "The api-name of private-key in hsm",
			EnvVars: opservice.PrefixEnvVar(envPrefix, "HSM_API_NAME"),
			Value:   "",
		},
		&cli.StringFlag{
			Name:    HsmCredenFlagName,
			Usage:   "The creden of private-key in hsm",
			EnvVars: opservice.PrefixEnvVar(envPrefix, "HSM_CREDEN"),
			Value:   "",
		},
	}, client.CLIFlags(envPrefix)...)
}

type CLIConfig struct {
	L1RPCURL                  string
	Mnemonic                  string
	HDPath                    string
	SequencerHDPath           string
	L2OutputHDPath            string
	PrivateKey                string
	SignerCLIConfig           client.CLIConfig
	NumConfirmations          uint64
	SafeAbortNonceTooLowCount uint64
	FeeLimitMultiplier        uint64
	FeeLimitThresholdGwei     float64
	ResubmissionTimeout       time.Duration
	ReceiptQueryInterval      time.Duration
	NetworkTimeout            time.Duration
	TxSendTimeout             time.Duration
	TxNotInMempoolTimeout     time.Duration
	EnableHsm                 bool
	HsmCreden                 string
	HsmAddress                string
	HsmAPIName                string
}

func (m CLIConfig) Check() error {
	if m.L1RPCURL == "" {
		return errors.New("must provide a L1 RPC url")
	}
	if m.NumConfirmations == 0 {
		return errors.New("NumConfirmations must not be 0")
	}
	if m.NetworkTimeout == 0 {
		return errors.New("must provide NetworkTimeout")
	}
	if m.ResubmissionTimeout == 0 {
		return errors.New("must provide ResubmissionTimeout")
	}
	if m.ReceiptQueryInterval == 0 {
		return errors.New("must provide ReceiptQueryInterval")
	}
	if m.TxNotInMempoolTimeout == 0 {
		return errors.New("must provide TxNotInMempoolTimeout")
	}
	if m.SafeAbortNonceTooLowCount == 0 {
		return errors.New("SafeAbortNonceTooLowCount must not be 0")
	}
	if err := m.SignerCLIConfig.Check(); err != nil {
		return err
	}
	return nil
}

func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		L1RPCURL:                  ctx.String(L1RPCFlagName),
		Mnemonic:                  ctx.String(MnemonicFlagName),
		HDPath:                    ctx.String(HDPathFlagName),
		SequencerHDPath:           ctx.String(SequencerHDPathFlag.Name),
		L2OutputHDPath:            ctx.String(L2OutputHDPathFlag.Name),
		PrivateKey:                ctx.String(PrivateKeyFlagName),
		SignerCLIConfig:           client.ReadCLIConfig(ctx),
		NumConfirmations:          ctx.Uint64(NumConfirmationsFlagName),
		SafeAbortNonceTooLowCount: ctx.Uint64(SafeAbortNonceTooLowCountFlagName),
		FeeLimitMultiplier:        ctx.Uint64(FeeLimitMultiplierFlagName),
		FeeLimitThresholdGwei:     ctx.Float64(FeeLimitThresholdFlagName),
		ResubmissionTimeout:       ctx.Duration(ResubmissionTimeoutFlagName),
		ReceiptQueryInterval:      ctx.Duration(ReceiptQueryIntervalFlagName),
		NetworkTimeout:            ctx.Duration(NetworkTimeoutFlagName),
		TxSendTimeout:             ctx.Duration(TxSendTimeoutFlagName),
		TxNotInMempoolTimeout:     ctx.Duration(TxNotInMempoolTimeoutFlagName),
		EnableHsm:                 ctx.Bool(EnableHsmFlagName),
		HsmAddress:                ctx.String(HsmAddressFlagName),
		HsmAPIName:                ctx.String(HsmAPINameFlagName),
		HsmCreden:                 ctx.String(HsmCredenFlagName),
	}
}

func NewConfig(cfg CLIConfig, l log.Logger) (Config, error) {
	if err := cfg.Check(); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.NetworkTimeout)
	defer cancel()
	l1, err := ethclient.DialContext(ctx, cfg.L1RPCURL)
	if err != nil {
		return Config{}, fmt.Errorf("could not dial eth client: %w", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), cfg.NetworkTimeout)
	defer cancel()
	chainID, err := l1.ChainID(ctx)
	if err != nil {
		return Config{}, fmt.Errorf("could not dial fetch L1 chain ID: %w", err)
	}

	// Allow backwards compatible ways of specifying the HD path
	hdPath := cfg.HDPath
	if hdPath == "" && cfg.SequencerHDPath != "" {
		hdPath = cfg.SequencerHDPath
	} else if hdPath == "" && cfg.L2OutputHDPath != "" {
		hdPath = cfg.L2OutputHDPath
	}

	ctx = context.Background()
	signerFactory, from, err := opcrypto.SignerFactoryFromConfig(l, cfg.PrivateKey, cfg.Mnemonic, hdPath, cfg.SignerCLIConfig, ctx, cfg.EnableHsm, cfg.HsmAPIName, cfg.HsmAddress, cfg.HsmCreden)
	if err != nil {
		return Config{}, fmt.Errorf("could not init signer: %w", err)
	}

	feeLimitThreshold, err := eth.GweiToWei(cfg.FeeLimitThresholdGwei)
	if err != nil {
		return Config{}, fmt.Errorf("invalid fee limit threshold: %w", err)
	}

	return Config{
		Backend:                   l1,
		ResubmissionTimeout:       cfg.ResubmissionTimeout,
		FeeLimitMultiplier:        cfg.FeeLimitMultiplier,
		FeeLimitThreshold:         feeLimitThreshold,
		ChainID:                   chainID,
		TxSendTimeout:             cfg.TxSendTimeout,
		TxNotInMempoolTimeout:     cfg.TxNotInMempoolTimeout,
		NetworkTimeout:            cfg.NetworkTimeout,
		ReceiptQueryInterval:      cfg.ReceiptQueryInterval,
		NumConfirmations:          cfg.NumConfirmations,
		SafeAbortNonceTooLowCount: cfg.SafeAbortNonceTooLowCount,
		Signer:                    signerFactory(chainID),
		From:                      from,
		EnableHsm:                 cfg.EnableHsm,
		HsmAddress:                cfg.HsmAddress,
		HsmCreden:                 cfg.HsmCreden,
		HsmAPIName:                cfg.HsmAPIName,
	}, nil
}

// Config houses parameters for altering the behavior of a SimpleTxManager.
type Config struct {
	Backend ETHBackend
	// ResubmissionTimeout is the interval at which, if no previously
	// published transaction has been mined, the new tx with a bumped gas
	// price will be published. Only one publication at MaxGasPrice will be
	// attempted.
	ResubmissionTimeout time.Duration

	// The multiplier applied to fee suggestions to put a hard limit on fee increases.
	FeeLimitMultiplier uint64

	// Minimum threshold (in Wei) at which the FeeLimitMultiplier takes effect.
	// On low-fee networks, like test networks, this allows for arbitrary fee bumps
	// below this threshold.
	FeeLimitThreshold *big.Int

	// ChainID is the chain ID of the L1 chain.
	ChainID *big.Int

	// TxSendTimeout is how long to wait for sending a transaction.
	// By default it is unbounded. If set, this is recommended to be at least 20 minutes.
	TxSendTimeout time.Duration

	// TxNotInMempoolTimeout is how long to wait before aborting a transaction send if the transaction does not
	// make it to the mempool. If the tx is in the mempool, TxSendTimeout is used instead.
	TxNotInMempoolTimeout time.Duration

	// NetworkTimeout is the allowed duration for a single network request.
	// This is intended to be used for network requests that can be replayed.
	NetworkTimeout time.Duration

	// RequireQueryInterval is the interval at which the tx manager will
	// query the backend to check for confirmations after a tx at a
	// specific gas price has been published.
	ReceiptQueryInterval time.Duration

	// NumConfirmations specifies how many blocks are need to consider a
	// transaction confirmed.
	NumConfirmations uint64

	// SafeAbortNonceTooLowCount specifies how many ErrNonceTooLow observations
	// are required to give up on a tx at a particular nonce without receiving
	// confirmation.
	SafeAbortNonceTooLowCount uint64

	// Signer is used to sign transactions when the gas price is increased.
	Signer opcrypto.SignerFn
	From   common.Address

	// use cloud-hsm to sign
	EnableHsm  bool
	HsmCreden  string
	HsmAddress string
	HsmAPIName string
}
