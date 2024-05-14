package dln

import (
	"time"

	"github.com/Layr-Labs/datalayr/common/chain"
	"github.com/Layr-Labs/datalayr/common/encoding"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/Layr-Labs/datalayr/dl-node/flags"
	kzgRs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
	"github.com/urfave/cli"
)

// Config contains all of the configuration information for a DLN
type Config struct {
	Hostname      string
	GrpcPort      string
	EnableMetrics bool
	MetricsPort   string
	Timeout       time.Duration
	DbPath        string
	LogPath       string

	GraphProvider string
	PrivateBls    string
	Address       string
	DlsmAddress   string

	ChallengeOrder uint64

	ChainClientConfig chain.ClientConfig
	LoggingConfig     logging.Config
	KzgConfig         kzgRs.KzgConfig
}

// NewConfig parses the Config from the provided flags or environment variables.
func NewConfig(ctx *cli.Context) (*Config, error) {
	timeout, err := time.ParseDuration(ctx.GlobalString(flags.TimeoutFlag.Name))
	if err != nil {
		return &Config{}, err
	}

	return &Config{
		Hostname:          ctx.GlobalString(flags.HostnameFlag.Name),
		GrpcPort:          ctx.GlobalString(flags.GrpcPortFlag.Name),
		EnableMetrics:     ctx.GlobalBool(flags.EnableMetricsFlag.Name),
		MetricsPort:       ctx.GlobalString(flags.MetricsPortFlag.Name),
		Timeout:           timeout,
		DbPath:            ctx.GlobalString(flags.DbPathFlag.Name),
		GraphProvider:     ctx.GlobalString(flags.GraphProviderFlag.Name),
		PrivateBls:        ctx.GlobalString(flags.PrivateBlsFlag.Name),
		DlsmAddress:       ctx.GlobalString(flags.DlsmAddressFlag.Name),
		ChallengeOrder:    ctx.GlobalUint64(flags.ChallengeOrderFlag.Name),
		ChainClientConfig: chain.ReadCLIConfig(ctx),
		KzgConfig:         encoding.ReadCLIConfig(ctx),
		LoggingConfig:     logging.ReadCLIConfig(ctx),
	}, nil
}
