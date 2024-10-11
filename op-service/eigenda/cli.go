package eigenda

import (
	"errors"
	"time"

	"github.com/urfave/cli"
)

const (
	EigenDADisperserRpcFlagName = "eigenda-disperser-rpc"
	EigenDAProxyUrlFlagName     = "eigenda-proxy-url"
	DisperseBlobTimeoutFlagName = "eigenda-disperser-timeout"
	RetrieveBlobTimeoutFlagName = "eigenda-retrieve-timeout"
)

func PrefixEnvVar(prefix, suffix string) string {
	return prefix + "_" + suffix
}

type CLIConfig struct {
	EigenDADisperserRpc string
	EigenDAProxyUrl     string
	DisperseBlobTimeout time.Duration
	RetrieveBlobTimeout time.Duration
}

// NewConfig parses the Config from the provided flags or environment variables.
func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		/* Required Flags */
		EigenDADisperserRpc: ctx.String(EigenDADisperserRpcFlagName),
		EigenDAProxyUrl:     ctx.String(EigenDAProxyUrlFlagName),

		/* Optional Flags */
		DisperseBlobTimeout: ctx.Duration(DisperseBlobTimeoutFlagName),
		RetrieveBlobTimeout: ctx.Duration(RetrieveBlobTimeoutFlagName),
	}
}

func (m CLIConfig) Check() error {
	if m.EigenDADisperserRpc == "" {
		return errors.New("must provide a DA disperser rpc url")
	}
	if m.EigenDAProxyUrl == "" {
		return errors.New("must provide a DA disperser url")
	}

	return nil
}

func CLIFlags(envPrefix string) []cli.Flag {
	prefixEnvVars := func(name string) string {
		return PrefixEnvVar(envPrefix, name)
	}
	return []cli.Flag{
		cli.StringFlag{
			Name:     EigenDADisperserRpcFlagName,
			Usage:    "RPC endpoint of the EigenDA disperser",
			EnvVar:   prefixEnvVars("EIGEN_DA_DISPERSER_URL"),
			Required: true,
		},
		cli.StringFlag{
			Name:     EigenDAProxyUrlFlagName,
			Usage:    "HTTP endpoint of the EigenDA proxy",
			EnvVar:   prefixEnvVars("EIGEN_DA_PROXY_URL"),
			Required: true,
		},
		cli.DurationFlag{
			Name:   DisperseBlobTimeoutFlagName,
			Usage:  "Timeout for EigenDA disperse blob",
			EnvVar: prefixEnvVars("EIGEN_DA_DISPERSER_TIMEOUT"),
			Value:  20 * time.Minute,
		},
		cli.DurationFlag{
			Name:   RetrieveBlobTimeoutFlagName,
			Usage:  "Timeout for EigenDA retrieve blob",
			EnvVar: prefixEnvVars("EIGEN_DA_RETRIEVE_TIMEOUT"),
			Value:  30 * time.Second,
		},
	}
}
