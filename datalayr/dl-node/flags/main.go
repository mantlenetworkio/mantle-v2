package flags

import (
	"github.com/Layr-Labs/datalayr/common"
	"github.com/Layr-Labs/datalayr/common/chain"
	"github.com/Layr-Labs/datalayr/common/encoding"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/urfave/cli"
	// node "github.com/Layr-Labs/datalayr/dl-node"
)

const envVarPrefix = "DL_NODE"

// TODO: Some of these flags should be integers?

var (
	/* Required Flags */

	HostnameFlag = cli.StringFlag{
		Name:     "hostname",
		Usage:    "Hostname at which node is available",
		Required: true,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "HOSTNAME"),
	}
	GrpcPortFlag = cli.StringFlag{
		Name:     "grpc-port",
		Usage:    "Port at which node listens for grpc calls",
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
		Usage:    "Port at which node listens for metrics calls",
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
	PrivateBlsFlag = cli.StringFlag{
		Name:     "private-bls",
		Usage:    "BLS private key for node operator",
		Required: true,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "PRIVATE_BLS"),
	}
	DlsmAddressFlag = cli.StringFlag{
		Name:     "dlsm-address",
		Usage:    "Address of the datalayr service manager contract",
		Required: true,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "DLSM_ADDRESS"),
	}
	ChallengeOrderFlag = cli.StringFlag{
		Name:     "challenge-order",
		Usage:    "Order of the challenge",
		Required: true,
		EnvVar:   common.PrefixEnvVar(envVarPrefix, "CHALLENGE_ORDER"),
	}
)

var requiredFlags = []cli.Flag{
	HostnameFlag,
	GrpcPortFlag,
	EnableMetricsFlag,
	MetricsPortFlag,
	TimeoutFlag,
	DbPathFlag,
	GraphProviderFlag,
	PrivateBlsFlag,
	DlsmAddressFlag,
	ChallengeOrderFlag,
}

var optionalFlags = []cli.Flag{}

func init() {
	Flags = append(requiredFlags, optionalFlags...)
	Flags = append(Flags, chain.CLIFlags(envVarPrefix)...)
	Flags = append(Flags, logging.CLIFlags(envVarPrefix)...)
	Flags = append(Flags, encoding.CLIFlags(envVarPrefix)...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag
