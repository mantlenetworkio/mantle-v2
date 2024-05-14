package retriever

import (
	"time"

	"github.com/Layr-Labs/datalayr/common/chain"
	"github.com/Layr-Labs/datalayr/common/encoding"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/Layr-Labs/datalayr/dl-retriever/flags"
	kzgRs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
	"github.com/urfave/cli"
)

type Config struct {
	Hostname      string
	GrpcPort      string
	DlsmAddress   string
	GraphProvider string
	Timeout       time.Duration

	ChainClientConfig chain.ClientConfig
	KzgConfig         kzgRs.KzgConfig
	LoggingConfig     logging.Config
}

// NewConfig parses the Config from the provided flags or environment variables.
func NewConfig(ctx *cli.Context) (*Config, error) {
	t, err := time.ParseDuration(ctx.String(flags.TimeoutFlag.Name))
	if err != nil {
		return nil, err
	}

	return &Config{
		Hostname:          ctx.String(flags.HostnameFlag.Name),
		GrpcPort:          ctx.String(flags.GrpcPortFlag.Name),
		Timeout:           t,
		GraphProvider:     ctx.String(flags.GraphProviderFlag.Name),
		DlsmAddress:       ctx.String(flags.DlsmAddressFlag.Name),
		ChainClientConfig: chain.ReadCLIConfig(ctx),
		KzgConfig:         encoding.ReadCLIConfig(ctx),
		LoggingConfig:     logging.ReadCLIConfig(ctx),
	}, nil
}
