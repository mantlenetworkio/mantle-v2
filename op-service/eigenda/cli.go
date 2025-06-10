package eigenda

import (
	"errors"
	"time"

	"github.com/urfave/cli/v2"
)

const (
	EigenDADisperserUrlFlagName = "eigenda-disperser-url"
	EigenDAProxyUrlFlagName     = "eigenda-proxy-url"
	DisperseBlobTimeoutFlagName = "eigenda-disperser-timeout"
	RetrieveBlobTimeoutFlagName = "eigenda-retrieve-timeout"
)

func PrefixEnvVar(prefix, suffix string) string {
	return prefix + "_" + suffix
}

type CLIConfig struct {
	EigenDADisperserUrl string
	EigenDAProxyUrl     string
	DisperseBlobTimeout time.Duration
	RetrieveBlobTimeout time.Duration
}

// NewConfig parses the Config from the provided flags or environment variables.
func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		/* Required Flags */
		EigenDADisperserUrl: ctx.String(EigenDADisperserUrlFlagName),
		EigenDAProxyUrl:     ctx.String(EigenDAProxyUrlFlagName),

		/* Optional Flags */
		DisperseBlobTimeout: ctx.Duration(DisperseBlobTimeoutFlagName),
		RetrieveBlobTimeout: ctx.Duration(RetrieveBlobTimeoutFlagName),
	}
}

func (m CLIConfig) Check() error {
	if m.EigenDADisperserUrl == "" {
		return errors.New("must provide a DA disperser rpc url")
	}
	if m.EigenDAProxyUrl == "" {
		return errors.New("must provide a DA disperser url")
	}

	return nil
}

func CLIFlags(envPrefix string) []cli.Flag {
	prefixEnvVars := func(name string) []string {
		return []string{PrefixEnvVar(envPrefix, name)}
	}
	return []cli.Flag{
		&cli.StringFlag{
			Name:    EigenDADisperserUrlFlagName,
			Usage:   "RPC endpoint of the EigenDA disperser",
			EnvVars: prefixEnvVars("EIGENDA_DISPERSER_URL"),
		},
		&cli.StringFlag{
			Name:    EigenDAProxyUrlFlagName,
			Usage:   "HTTP endpoint of the EigenDA proxy",
			EnvVars: prefixEnvVars("EIGENDA_PROXY_URL"),
		},
		&cli.DurationFlag{
			Name:    DisperseBlobTimeoutFlagName,
			Usage:   "Timeout for EigenDA disperse blob",
			EnvVars: prefixEnvVars("EIGENDA_DISPERSER_TIMEOUT"),
			Value:   20 * time.Minute,
		},
		&cli.DurationFlag{
			Name:    RetrieveBlobTimeoutFlagName,
			Usage:   "Timeout for EigenDA retrieve blob",
			EnvVars: prefixEnvVars("EIGENDA_RETRIEVE_TIMEOUT"),
			Value:   30 * time.Second,
		},
	}
}
