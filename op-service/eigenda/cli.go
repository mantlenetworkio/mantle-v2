package eigenda

import (
	"errors"
	"time"

	"github.com/urfave/cli"
)

const (
	RPCFlagName                      = "da-rpc"
	StatusQueryRetryIntervalFlagName = "da-status-query-retry-interval"
	StatusQueryTimeoutFlagName       = "da-status-query-timeout"
	DARPCTimeoutFlagName             = "da-rpc-timeout"
	EnableDAHsmFlagName              = "enable-da-hsm"
	DAHsmCredenFlagName              = "da-hsm-creden"
	DAHsmPubkeyFlagName              = "da-hsm-pubkey"
	DAHsmAPINameFlagName             = "da-hsm-api-name"
	DAPrivateKeyFlagName             = "da-private-key"
)

func PrefixEnvVar(prefix, suffix string) string {
	return prefix + "_" + suffix
}

type CLIConfig struct {
	RPC                      string
	StatusQueryRetryInterval time.Duration
	StatusQueryTimeout       time.Duration
	DARPCTimeout             time.Duration
	EnableHsm                bool
	HsmCreden                string
	HsmPubkey                string
	HsmAPIName               string
	PrivateKey               string
}

// NewConfig parses the Config from the provided flags or environment variables.
func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		/* Required Flags */
		RPC:                      ctx.String(RPCFlagName),
		StatusQueryRetryInterval: ctx.Duration(StatusQueryRetryIntervalFlagName),
		StatusQueryTimeout:       ctx.Duration(StatusQueryTimeoutFlagName),
		DARPCTimeout:             ctx.Duration(DARPCTimeoutFlagName),

		/* Optional Flags */
		EnableHsm:  ctx.Bool(EnableDAHsmFlagName),
		HsmCreden:  ctx.String(DAHsmCredenFlagName),
		HsmPubkey:  ctx.String(DAHsmPubkeyFlagName),
		HsmAPIName: ctx.String(DAHsmAPINameFlagName),
		PrivateKey: ctx.String(DAPrivateKeyFlagName),
	}
}

func (m CLIConfig) Check() error {
	if m.RPC == "" {
		return errors.New("must provide a DA RPC url")
	}
	if m.StatusQueryTimeout == 0 {
		return errors.New("DA status query timeout must be greater than 0")
	}
	if m.StatusQueryRetryInterval == 0 {
		return errors.New("DA status query retry interval must be greater than 0")
	}
	if m.EnableHsm {
		if m.HsmCreden == "" {
			return errors.New("must provide a HSM creden")
		}
		if m.HsmPubkey == "" {
			return errors.New("must provide a HSM pubkey")
		}
		if m.HsmAPIName == "" {
			return errors.New("must provide a HSM API name")
		}
	}
	return nil
}

func CLIFlags(envPrefix string) []cli.Flag {
	prefixEnvVars := func(name string) string {
		return PrefixEnvVar(envPrefix, name)
	}
	return []cli.Flag{
		cli.StringFlag{
			Name:   RPCFlagName,
			Usage:  "RPC endpoint of the EigenDA disperser",
			EnvVar: prefixEnvVars("DA_RPC"),
		},
		cli.DurationFlag{
			Name:   StatusQueryTimeoutFlagName,
			Usage:  "Timeout for aborting an EigenDA blob dispersal if the disperser does not report that the blob has been confirmed dispersed.",
			Value:  20 * time.Minute,
			EnvVar: prefixEnvVars("DA_STATUS_QUERY_TIMEOUT"),
		},
		cli.DurationFlag{
			Name:   StatusQueryRetryIntervalFlagName,
			Usage:  "Wait time between retries of EigenDA blob status queries (made while waiting for a blob to be confirmed by)",
			Value:  5 * time.Second,
			EnvVar: prefixEnvVars("DA_STATUS_QUERY_INTERVAL"),
		},
		cli.DurationFlag{
			Name:   DARPCTimeoutFlagName,
			Usage:  "Timeout for EigenDA rpc calls",
			Value:  5 * time.Second,
			EnvVar: prefixEnvVars("DA_RPC_TIMEOUT"),
		},
		cli.BoolFlag{
			Name:   EnableDAHsmFlagName,
			Usage:  "EigenDA whether or not to use cloud hsm",
			EnvVar: prefixEnvVars("ENABLE_DA_HSM"),
		},
		cli.StringFlag{
			Name:   DAHsmPubkeyFlagName,
			Usage:  "The public-key of EigenDA account in hsm",
			EnvVar: prefixEnvVars("DA_HSM_PUBKEY"),
			Value:  "",
		},
		cli.StringFlag{
			Name:   DAHsmAPINameFlagName,
			Usage:  "The api-name of EigenDA account in hsm",
			EnvVar: prefixEnvVars("DA_HSM_API_NAME"),
			Value:  "",
		},
		cli.StringFlag{
			Name:   DAHsmCredenFlagName,
			Usage:  "The creden of EigenDA account in hsm",
			EnvVar: prefixEnvVars("DA_HSM_CREDEN"),
			Value:  "",
		},
		cli.StringFlag{
			Name:   DAPrivateKeyFlagName,
			Usage:  "The private-key of EigenDA account",
			EnvVar: prefixEnvVars("DA_PRIVATE_KEY"),
			Value:  "",
		},
	}
}
