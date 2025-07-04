package flags

import (
	"github.com/urfave/cli/v2"

	opservice "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
)

const envPrefix = "OP_HEARTBEAT"

const (
	HTTPAddrFlagName = "http.addr"
	HTTPPortFlagName = "http.port"
)

var (
	HTTPAddrFlag = &cli.StringFlag{
		Name:    HTTPAddrFlagName,
		Usage:   "Address the server should listen on",
		Value:   "0.0.0.0",
		EnvVars: opservice.PrefixEnvVar(envPrefix, "HTTP_ADDR"),
	}
	HTTPPortFlag = &cli.IntFlag{
		Name:    HTTPPortFlagName,
		Usage:   "Port the server should listen on",
		Value:   8080,
		EnvVars: opservice.PrefixEnvVar(envPrefix, "HTTP_PORT"),
	}
)

var Flags []cli.Flag

func init() {
	Flags = []cli.Flag{
		HTTPAddrFlag,
		HTTPPortFlag,
	}

	Flags = append(Flags, oplog.CLIFlags(envPrefix)...)
	Flags = append(Flags, opmetrics.CLIFlags(envPrefix)...)
}
