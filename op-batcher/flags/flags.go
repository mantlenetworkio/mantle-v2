package flags

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/optimism/op-batcher/compressor"
	"github.com/ethereum-optimism/optimism/op-batcher/rpc"
	opservice "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/eigenda"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	oppprof "github.com/ethereum-optimism/optimism/op-service/pprof"
	oprpc "github.com/ethereum-optimism/optimism/op-service/rpc"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

const EnvVarPrefix = "OP_BATCHER"

var (
	// Required flags
	L1EthRpcFlag = &cli.StringFlag{
		Name:    "l1-eth-rpc",
		Usage:   "HTTP provider URL for L1",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "L1_ETH_RPC"),
	}
	L2EthRpcFlag = &cli.StringFlag{
		Name:    "l2-eth-rpc",
		Usage:   "HTTP provider URL for L2 execution engine",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "L2_ETH_RPC"),
	}
	RollupRpcFlag = &cli.StringFlag{
		Name:    "rollup-rpc",
		Usage:   "HTTP provider URL for Rollup node",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "ROLLUP_RPC"),
	}
	// Optional flags
	DisperserSocketFlag = &cli.StringFlag{
		Name:    "disperser-socket",
		Usage:   "Websocket for MantleDA disperser",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "DISPERSER_SOCKET"),
	}
	DataStoreDurationFlag = &cli.Uint64Flag{
		Name:    "datastore-duration",
		Usage:   "Duration to store blob",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "DATA_STORE_DURATION"),
	}
	DisperserTimeoutFlag = &cli.DurationFlag{
		Name:    "disperser-timeout",
		Usage:   "disperser timeout",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "DISPERSER_TIMEOUT"),
	}
	GraphPollingDurationFlag = &cli.DurationFlag{
		Name:    "graph-polling-duration",
		Usage:   "polling duration for fetch data from da graph node",
		Value:   1200 * time.Millisecond,
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "GRAPH_POLLING_DURATION"),
	}
	GraphProviderFlag = &cli.StringFlag{
		Name:    "graph-node-provider",
		Usage:   "graph node url of MantleDA graph node",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "GRAPH_PROVIDER"),
	}
	RollUpMaxSizeFlag = &cli.Uint64Flag{
		Name:    "rollup-max-size",
		Usage:   "Each rollup data to MantleDa maximum limit, rollup data can not be greater than the value, otherwise the rollup failure",
		Value:   31600, // ktz for order is 3000
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "ROLLUP_MAX_SIZE"),
	}

	SubSafetyMarginFlag = &cli.Uint64Flag{
		Name: "sub-safety-margin",
		Usage: "The batcher tx submission safety margin (in #L1-blocks) to subtract " +
			"from a channel's timeout and sequencing window, to guarantee safe inclusion " +
			"of a channel on L1.",
		Value:   10,
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "SUB_SAFETY_MARGIN"),
	}
	PollIntervalFlag = &cli.DurationFlag{
		Name:    "poll-interval",
		Usage:   "How frequently to poll L2 for new blocks",
		Value:   6 * time.Second,
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "POLL_INTERVAL"),
	}
	MaxPendingTransactionsFlag = &cli.Uint64Flag{
		Name:    "max-pending-tx",
		Usage:   "The maximum number of pending transactions. 0 for no limit.",
		Value:   1,
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "MAX_PENDING_TX"),
	}
	MaxChannelDurationFlag = &cli.Uint64Flag{
		Name:    "max-channel-duration",
		Usage:   "The maximum duration of L1-blocks to keep a channel open. 0 to disable.",
		Value:   0,
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "MAX_CHANNEL_DURATION"),
	}
	MaxL1TxSizeBytesFlag = &cli.Uint64Flag{
		Name:    "max-l1-tx-size-bytes",
		Usage:   "The maximum size of a batch tx submitted to L1.",
		Value:   120_000,
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "MAX_L1_TX_SIZE_BYTES"),
	}
	StoppedFlag = &cli.BoolFlag{
		Name:    "stopped",
		Usage:   "Initialize the batcher in a stopped state. The batcher can be started using the admin_startBatcher RPC",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "STOPPED"),
	}
	SkipEigenDaRpcFlag = &cli.BoolFlag{
		Name:    "skip-eigenda-da-rpc",
		Usage:   "skip eigenDA rpc and submit da data to ethereum blob transaction when mantle_da_switch is open",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "SKIP_EIGENDA_DA_RPC"),
	}
	// Legacy Flags
	SequencerHDPathFlag = txmgr.SequencerHDPathFlag
)

var requiredFlags = []cli.Flag{
	L1EthRpcFlag,
	L2EthRpcFlag,
	RollupRpcFlag,
}

var optionalFlags = []cli.Flag{
	SubSafetyMarginFlag,
	PollIntervalFlag,
	MaxPendingTransactionsFlag,
	MaxChannelDurationFlag,
	MaxL1TxSizeBytesFlag,
	StoppedFlag,
	SkipEigenDaRpcFlag,
	SequencerHDPathFlag,
	DisperserTimeoutFlag,
	DisperserSocketFlag,
	DataStoreDurationFlag,
	GraphPollingDurationFlag,
	GraphProviderFlag,
	RollUpMaxSizeFlag,
}

func init() {
	optionalFlags = append(optionalFlags, oprpc.CLIFlags(EnvVarPrefix)...)
	optionalFlags = append(optionalFlags, oplog.CLIFlags(EnvVarPrefix)...)
	optionalFlags = append(optionalFlags, opmetrics.CLIFlags(EnvVarPrefix)...)
	optionalFlags = append(optionalFlags, oppprof.CLIFlags(EnvVarPrefix)...)
	optionalFlags = append(optionalFlags, rpc.CLIFlags(EnvVarPrefix)...)
	optionalFlags = append(optionalFlags, txmgr.CLIFlags(EnvVarPrefix)...)
	optionalFlags = append(optionalFlags, compressor.CLIFlags(EnvVarPrefix)...)
	optionalFlags = append(optionalFlags, eigenda.CLIFlags(EnvVarPrefix)...)

	Flags = append(requiredFlags, optionalFlags...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag

func CheckRequired(ctx *cli.Context) error {
	for _, f := range requiredFlags {
		if !ctx.IsSet(f.Names()[0]) {
			return fmt.Errorf("flag %s is required", f.Names()[0])
		}
	}
	return nil
}
