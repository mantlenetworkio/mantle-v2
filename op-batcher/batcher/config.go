package batcher

import (
	"errors"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-batcher/compressor"
	"github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-batcher/rpc"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/sources"
	"github.com/ethereum-optimism/optimism/op-service/eigenda"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	oppprof "github.com/ethereum-optimism/optimism/op-service/pprof"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum-optimism/optimism/op-service/upgrade"
)

var (
	ErrDisperserSocketEmpty     = errors.New("disperser socket is empty for MantleDA")
	ErrDisperserTimeoutZero     = errors.New("disperser timeout is zero for MantleDA")
	ErrDataStoreDurationZero    = errors.New("datastore duration is zero for MantleDA")
	ErrGraphPollingDurationZero = errors.New("graph node polling duration is zero for MantleDA")
	ErrGraphProviderEmpty       = errors.New("graph node provider is empty for MantleDA")
)

type Config struct {
	log        log.Logger
	metr       metrics.Metricer
	L1Client   *ethclient.Client
	L2Client   *ethclient.Client
	RollupNode *sources.RollupClient
	TxManager  txmgr.TxManager

	NetworkTimeout         time.Duration
	PollInterval           time.Duration
	MaxPendingTransactions uint64

	// Rollup MantleDA
	DisperserSocket                string
	DisperserTimeout               time.Duration
	DataStoreDuration              uint64
	GraphPollingDuration           time.Duration
	RollupMaxSize                  uint64
	MantleDaNodes                  int
	DataLayrServiceManagerAddr     common.Address
	DataLayrServiceManagerContract *bindings.ContractDataLayrServiceManager
	DataLayrServiceManagerABI      *abi.ABI

	// RollupConfig is queried at startup
	Rollup *rollup.Config

	// Channel builder parameters
	Channel ChannelConfig
	EigenDA eigenda.Config
	// Upgrade Da from MantleDA to EigenDA
	DaUpgradeChainConfig *upgrade.UpgradeChainConfig
	//skip eigenda and submit to ethereum blob transaction
	SkipEigenDaRpc bool
}

// Check ensures that the [Config] is valid.
func (c *Config) Check() error {
	if err := c.Rollup.Check(); err != nil {
		return err
	}
	if err := c.Channel.Check(); err != nil {
		return err
	}
	if c.Rollup.MantleDaSwitch {
		if len(c.DisperserSocket) == 0 {
			return ErrDisperserSocketEmpty
		}
		if c.DisperserTimeout == 0 {
			return ErrDisperserTimeoutZero
		}
		if c.DataStoreDuration == 0 {
			return ErrDataStoreDurationZero
		}
		if c.GraphPollingDuration == 0 {
			return ErrGraphPollingDurationZero
		}
	}
	return nil
}

type CLIConfig struct {
	// L1EthRpc is the HTTP provider URL for L1.
	L1EthRpc string

	// L2EthRpc is the HTTP provider URL for the L2 execution engine.
	L2EthRpc string

	// RollupRpc is the HTTP provider URL for the L2 rollup node.
	RollupRpc string

	// DisperserSocket is the websocket for the MantleDA disperser.
	DisperserSocket string

	// DisperserTimeout timeout for context
	DisperserTimeout time.Duration

	// DataStoreDuration data store time on MantleDA
	DataStoreDuration uint64

	//GraphPollingDuration listen to graph node polling time
	GraphPollingDuration time.Duration

	//GraphProvider is graph node url of MantleDA
	GraphProvider string

	//RollupMaxSize is the maximum size of tx data that can be rollup to MantleDA at one time
	RollupMaxSize uint64

	//The number of MantleDA nodes
	MantleDaNodes int

	// MaxChannelDuration is the maximum duration (in #L1-blocks) to keep a
	// channel open. This allows to more eagerly send batcher transactions
	// during times of low L2 transaction volume. Note that the effective
	// L1-block distance between batcher transactions is then MaxChannelDuration
	// + NumConfirmations because the batcher waits for NumConfirmations blocks
	// after sending a batcher tx and only then starts a new channel.
	//
	// If 0, duration checks are disabled.
	MaxChannelDuration uint64

	// The batcher tx submission safety margin (in #L1-blocks) to subtract from
	// a channel's timeout and sequencing window, to guarantee safe inclusion of
	// a channel on L1.
	SubSafetyMargin uint64

	// PollInterval is the delay between querying L2 for more transaction
	// and creating a new batch.
	PollInterval time.Duration

	// MaxPendingTransactions is the maximum number of concurrent pending
	// transactions sent to the transaction manager.
	MaxPendingTransactions uint64

	// MaxL1TxSize is the maximum size of a batch tx submitted to L1.
	MaxL1TxSize uint64

	Stopped bool

	TxMgrConfig      txmgr.CLIConfig
	RPCConfig        rpc.CLIConfig
	LogConfig        oplog.CLIConfig
	MetricsConfig    opmetrics.CLIConfig
	PprofConfig      oppprof.CLIConfig
	CompressorConfig compressor.CLIConfig

	EigenDAConfig  eigenda.CLIConfig
	SkipEigenDaRpc bool
}

func (c CLIConfig) Check() error {
	if err := c.RPCConfig.Check(); err != nil {
		return err
	}
	if err := c.MetricsConfig.Check(); err != nil {
		return err
	}
	if err := c.PprofConfig.Check(); err != nil {
		return err
	}
	if err := c.TxMgrConfig.Check(); err != nil {
		return err
	}
	if err := c.EigenDAConfig.Check(); err != nil {
		return err
	}

	//Used to ensure that when using an Ethereum blob, a single frame is not larger than MaxBlobDataSize * MaxblobNum
	//Considering that the frame needs to go through rlp. EncodeToBytes before submitting da, the maximum size after encoding cannot be accurately calculated.
	//So MaxL1TxSize > MaxBlobDataSize is used here to make a rough judgment
	if c.MaxL1TxSize > eth.MaxBlobDataSize {
		return errors.New("MaxL1TxSize must less than MaxBlobDataSize")
	}
	return nil
}

// NewConfig parses the Config from the provided flags or environment variables.
func NewConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		/* Required Flags */
		L1EthRpc:        ctx.String(flags.L1EthRpcFlag.Name),
		L2EthRpc:        ctx.String(flags.L2EthRpcFlag.Name),
		RollupRpc:       ctx.String(flags.RollupRpcFlag.Name),
		SubSafetyMargin: ctx.Uint64(flags.SubSafetyMarginFlag.Name),
		PollInterval:    ctx.Duration(flags.PollIntervalFlag.Name),

		/* Optional Flags */
		MaxPendingTransactions: ctx.Uint64(flags.MaxPendingTransactionsFlag.Name),
		MaxChannelDuration:     ctx.Uint64(flags.MaxChannelDurationFlag.Name),
		MaxL1TxSize:            ctx.Uint64(flags.MaxL1TxSizeBytesFlag.Name),
		DisperserSocket:        ctx.String(flags.DisperserSocketFlag.Name),
		DisperserTimeout:       ctx.Duration(flags.DisperserTimeoutFlag.Name),
		DataStoreDuration:      ctx.Uint64(flags.DataStoreDurationFlag.Name),
		GraphPollingDuration:   ctx.Duration(flags.GraphPollingDurationFlag.Name),
		GraphProvider:          ctx.String(flags.GraphProviderFlag.Name),
		RollupMaxSize:          ctx.Uint64(flags.RollUpMaxSizeFlag.Name),
		Stopped:                ctx.Bool(flags.StoppedFlag.Name),
		SkipEigenDaRpc:         ctx.Bool(flags.SkipEigenDaRpcFlag.Name),
		TxMgrConfig:            txmgr.ReadCLIConfig(ctx),
		RPCConfig:              rpc.ReadCLIConfig(ctx),
		LogConfig:              oplog.ReadCLIConfig(ctx),
		MetricsConfig:          opmetrics.ReadCLIConfig(ctx),
		PprofConfig:            oppprof.ReadCLIConfig(ctx),
		CompressorConfig:       compressor.ReadCLIConfig(ctx),
		EigenDAConfig:          eigenda.ReadCLIConfig(ctx),
	}
}
