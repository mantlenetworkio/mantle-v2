package flags

import (
	"github.com/Layr-Labs/datalayr/common"
	"github.com/Layr-Labs/datalayr/common/chain"
	"github.com/Layr-Labs/datalayr/common/encoding"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/urfave/cli"
	"time"
)

const envVarPrefix = "DL_DISPERSER"

// TODO: Some of these flags should be integers?

var (
	/* Required Flags */

	HostnameFlag = cli.StringFlag{
		Name:     "hostname",
		Usage:    "Hostname at which disperser is available",
		Required: true,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "HOSTNAME"),
	}
	GrpcPortFlag = cli.StringFlag{
		Name:     "grpc-port",
		Usage:    "Port at which disperser listens for grpc calls",
		Required: true,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "GRPC_PORT"),
	}
	EnableMetricsFlag = cli.BoolFlag{
		Name:     "enable-metrics",
		Usage:    "enable prometheus to serve metrics collection",
		Required: true,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "ENABLE_METRICS"),
	}
	MetricsPortFlag = cli.StringFlag{
		Name:     "metrics-port",
		Usage:    "Port at which disperser listens for http calls",
		Required: false,
		Value:    "9091",
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "METRICS_PORT"),
	}
	TimeoutFlag = cli.StringFlag{
		Name:     "timeout",
		Usage:    "Amount of time to wait for GPRC",
		Required: true,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "TIMEOUT"),
	}
	PollingRetryFlag = cli.Uint64Flag{
		Name:     "polling-retry",
		Usage:    "number times of retry to fetch event from graph client",
		Required: true,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "POLLING_RETRY"),
	}
	DbPathFlag = cli.StringFlag{
		Name:     "db-path",
		Usage:    "Path for level db",
		Required: true,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "DB_PATH"),
	}
	GraphProviderFlag = cli.StringFlag{
		Name:     "graph-provider",
		Usage:    "Graphql endpoint for graph node",
		Required: true,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "GRAPH_PROVIDER"),
	}
	DlsmAddressFlag = cli.StringFlag{
		Name:     "dlsm-address",
		Usage:    "Address of the datalayr service manager contract",
		Required: true,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "DLSM_ADDRESS"),
	}
	UseCacheFlag = cli.BoolFlag{
		Name:     "use-cache",
		Usage:    "use cache which remember only the first seen data",
		Required: false,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "USE_CACHE"),
	}
	CodedCacheSizeFlag = cli.Uint64Flag{
		Name:     "coded-cache-size",
		Usage:    "the size of cache in bytes which holds dipersal data",
		Required: false,
		Value: uint64(2000000000),
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "MAX_CODED_CACHE"),
	}
	CodedCacheExpireDurationFlag = cli.Int64Flag{
		Name:     "coded-cache-expire-duration",
		Usage:    "the time to live for cached coded data if not dispersed. In Second",
		Required: false,
		Value: int64(120),
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "CODED_CACHE_EXPIRE_DURATION"),
	}
	CodedCacheCleanPeriodFlag = cli.DurationFlag{
		Name:     "coded-cache-clean-period",
		Usage:    "the frequency local process to clean expired cache",
		Required: false,
		Value: 10 * time.Second,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "CODED_CACHE_CLEAN_PERIOD"),
	}
)

var requiredFlags = []cli.Flag{
	HostnameFlag,
	GrpcPortFlag,
	EnableMetricsFlag,
	MetricsPortFlag,
	TimeoutFlag,
	PollingRetryFlag,
	DbPathFlag,
	GraphProviderFlag,
	DlsmAddressFlag,
}

var optionalFlags = []cli.Flag{
	UseCacheFlag,
	CodedCacheSizeFlag,
	CodedCacheExpireDurationFlag,
	CodedCacheCleanPeriodFlag,
}

func init() {
	Flags = append(requiredFlags, optionalFlags...)
	Flags = append(Flags, chain.CLIFlags(envVarPrefix)...)
	Flags = append(Flags, logging.CLIFlags(envVarPrefix)...)
	Flags = append(Flags, encoding.CLIFlags(envVarPrefix)...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag
