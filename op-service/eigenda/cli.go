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
	DAEthRPCFlagName                 = "eigenda-eth-rpc"
	DASvcManagerAddrFlagName         = "eigenda-svc-manager-addr"
	DAEthConfirmationDepthFlagName   = "eigenda-eth-confirmation-depth"
	// Kzg flags
	DAG1PathFlagName        = "eigenda-g1-path"
	DAG2TauFlagName         = "eigenda-g2-tau-path"
	DAMaxBlobLengthFlagName = "eigenda-max-blob-length"
	DACachePathFlagName     = "eigenda-cache-path"
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

	// ETH vars
	EthRPC               string
	SvcManagerAddr       string
	EthConfirmationDepth uint64

	// KZG vars
	CacheDir         string
	G1Path           string
	G2PowerOfTauPath string

	MaxBlobLength uint64
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

		G1Path:               ctx.String(DAG1PathFlagName),
		G2PowerOfTauPath:     ctx.String(DAG2TauFlagName),
		CacheDir:             ctx.String(DACachePathFlagName),
		MaxBlobLength:        ctx.Uint64(DAMaxBlobLengthFlagName),
		SvcManagerAddr:       ctx.String(DASvcManagerAddrFlagName),
		EthRPC:               ctx.String(DAEthRPCFlagName),
		EthConfirmationDepth: ctx.Uint64(DAEthConfirmationDepthFlagName),
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
		cli.StringFlag{
			Name:   DAEthRPCFlagName,
			Usage:  "JSON RPC node endpoint for the Ethereum network used for DA blobs verify. See available list here: https://docs.eigenlayer.xyz/eigenda/networks/",
			EnvVar: prefixEnvVars("DA_ETH_RPC"),
		},
		cli.StringFlag{
			Name:   DASvcManagerAddrFlagName,
			Usage:  "The deployed EigenDA service manager address. The list can be found here: https://github.com/Layr-Labs/eigenlayer-middleware/?tab=readme-ov-file#current-mainnet-deployment",
			EnvVar: prefixEnvVars("DA_SERVICE_MANAGER_ADDR"),
		},
		cli.Uint64Flag{
			Name:   DAEthConfirmationDepthFlagName,
			Usage:  "The number of Ethereum blocks of confirmation before verify.",
			EnvVar: prefixEnvVars("DA_ETH_CONFIRMATION_DEPTH"),
			Value:  6,
		},
		cli.Uint64Flag{
			Name:   DAMaxBlobLengthFlagName,
			Usage:  "Maximum blob length to be written or read from EigenDA. Determines the number of SRS points loaded into memory for KZG commitments. Example units: '30MiB', '4Kb', '30MB'. Maximum size slightly exceeds 1GB.",
			EnvVar: prefixEnvVars("DA_MAX_BLOB_LENGTH"),
			Value:  2_000_000,
		},
		cli.StringFlag{
			Name:   DAG1PathFlagName,
			Usage:  "Directory path to g1.point file.",
			EnvVar: prefixEnvVars("DA_TARGET_KZG_G1_PATH"),
			Value:  "resources/g1.point",
		},
		cli.StringFlag{
			Name:   DAG2TauFlagName,
			Usage:  "Directory path to g2.point.powerOf2 file.",
			EnvVar: prefixEnvVars("DA_TARGET_G2_TAU_PATH"),
			Value:  "resources/g2.point.powerOf2",
		},
		cli.StringFlag{
			Name:   DACachePathFlagName,
			Usage:  "Directory path to SRS tables for caching.",
			EnvVar: prefixEnvVars("DA_TARGET_CACHE_PATH"),
			Value:  "resources/SRSTables/",
		},
	}
}
