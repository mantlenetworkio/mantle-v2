package flags

import (
	"github.com/Layr-Labs/datalayr/common/chain"
	"github.com/Layr-Labs/datalayr/common/encoding"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/urfave/cli"
)

const envVarPrefix = "DL_RETRIEVER"

func prefixEnvVar(suffix string) string {
	return envVarPrefix + "_" + suffix
}

// TODO: Some of these flags should be integers?

var (
	/* Required Flags */

	HostnameFlag = cli.StringFlag{
		Name:     "hostname",
		Usage:    "Hostname at which retriever service is available",
		Required: true,
		EnvVar:   prefixEnvVar("HOSTNAME"),
	}
	GrpcPortFlag = cli.StringFlag{
		Name:     "grpc-port",
		Usage:    "Port at which a retriever listens for grpc calls",
		Required: true,
		EnvVar:   prefixEnvVar("GRPC_PORT"),
	}
	TimeoutFlag = cli.StringFlag{
		Name:     "timeout",
		Usage:    "Amount of time to wait for GPRC",
		Required: true,
		EnvVar:   prefixEnvVar("TIMEOUT"),
	}
	PollingRetryFlag = cli.Uint64Flag{
		Name:     "polling-retry",
		Usage:    "number times of retry to fetch event from graph client",
		Required: true,
		EnvVar:   prefixEnvVar("POLLING_RETRY"),
	}
	GraphProviderFlag = cli.StringFlag{
		Name:     "graph-provider",
		Usage:    "Graphql endpoint for graph node",
		Required: true,
		EnvVar:   prefixEnvVar("GRAPH_PROVIDER"),
	}
	DlsmAddressFlag = cli.StringFlag{
		Name:     "dlsm-address",
		Usage:    "Address of the datalayr service manager contract",
		Required: true,
		EnvVar:   prefixEnvVar("DLSM_ADDRESS"),
	}
)

var requiredFlags = []cli.Flag{
	HostnameFlag,
	GrpcPortFlag,
	TimeoutFlag,
	GraphProviderFlag,
	DlsmAddressFlag,
}

var optionalFlags = []cli.Flag{}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag

func init() {
	Flags = append(requiredFlags, optionalFlags...)
	Flags = append(Flags, chain.CLIFlags(envVarPrefix)...)
	Flags = append(Flags, logging.CLIFlags(envVarPrefix)...)
	Flags = append(Flags, encoding.CLIFlags(envVarPrefix)...)
}
