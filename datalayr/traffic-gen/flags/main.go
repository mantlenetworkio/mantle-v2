package flags

import (
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/urfave/cli"
)

const envVarPrefix = "DL_TRAFFIC_GENERATOR"

func prefixEnvVar(suffix string) string {
	return envVarPrefix + "_" + suffix
}

// TODO: Some of these flags should be integers?

var (
	/* Required Flags */

	HostnameFlag = cli.StringFlag{
		Name:     "hostname",
		Usage:    "Hostname at which a disperser is available",
		Required: true,
		EnvVar:   prefixEnvVar("HOSTNAME"),
	}
	GrpcPortFlag = cli.StringFlag{
		Name:     "grpc-port",
		Usage:    "port at which generator send grpc calls to",
		Required: true,
		EnvVar:   prefixEnvVar("GRPC_PORT"),
	}
	TimeoutFlag = cli.StringFlag{
		Name:     "timeout",
		Usage:    "Amount of time to wait for GPRC",
		Required: true,
		EnvVar:   prefixEnvVar("TIMEOUT"),
	}
	StoreDurationFlag = cli.Uint64Flag{
		Name:     "store-duration",
		Usage:    "amount of time the data persists",
		Required: true,
		EnvVar:   prefixEnvVar("STORE_DURATION"),
	}
	DataSizeFlag = cli.StringFlag{
		Name:     "data-size",
		Usage:    "size of data",
		Required: true,
		EnvVar:   prefixEnvVar("DATA_SIZE"),
	}
	LivenessThresholdFlag = cli.StringFlag{
		Name:     "live-threshold",
		Usage:    "liveness threshold, ratio of nodes have to sign",
		Required: true,
		EnvVar:   prefixEnvVar("LIVENESS_THRESHOLD"),
	}
	AdversarialThresholdFlag = cli.StringFlag{
		Name:     "adv-threshold",
		Usage:    "adversarial threshold",
		Required: true,
		EnvVar:   prefixEnvVar("ADV_THRESHOLD"),
	}
	IdlePeriodFlag = cli.StringFlag{
		Name:     "idle-period",
		Usage:    "how long to wait for after previous dispersal. In millisecond.",
		Required: true,
		EnvVar:   prefixEnvVar("IDLE_PERIOD"),
	}
	IdlePeriodStdFlag = cli.StringFlag{
		Name:     "idle-period-std",
		Usage:    "standard deviation (Gaussain) to add to the idle period. In millisecond",
		Required: true,
		EnvVar:   prefixEnvVar("IDLE_PERIOD_VAR"),
	}
	NumberFlag = cli.StringFlag{
		Name:     "number",
		Usage:    "number of dispersing instances to launch",
		Required: true,
		EnvVar:   prefixEnvVar("NUMBER"),
	}
)

var requiredFlags = []cli.Flag{
	HostnameFlag,
	GrpcPortFlag,
	TimeoutFlag,
	StoreDurationFlag,
	DataSizeFlag,
	LivenessThresholdFlag,
	AdversarialThresholdFlag,
	IdlePeriodFlag,
	IdlePeriodStdFlag,
	NumberFlag,
}

var optionalFlags = []cli.Flag{}

func init() {
	Flags = append(requiredFlags, optionalFlags...)
	Flags = append(Flags, logging.CLIFlags(envVarPrefix)...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag
