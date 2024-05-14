package disperser

import (
	"time"

	"github.com/Layr-Labs/datalayr/common/chain"
	"github.com/Layr-Labs/datalayr/common/encoding"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/Layr-Labs/datalayr/dl-disperser/flags"
	kzgRs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
	"github.com/urfave/cli"
)

type Config struct {
	Hostname                 string
	GrpcPort                 string
	EnableMetrics            bool
	MetricsPort              string
	Address                  string
	Timeout                  time.Duration
	DlsmAddress              string
	GraphProvider            string
	PollingRetry             uint64
	DbPath                   string
	ChainClientConfig        chain.ClientConfig
	KzgConfig                kzgRs.KzgConfig
	LoggingConfig            logging.Config
	UseCache                 bool
	CodedCacheSize           uint64
	CodedCacheExpireDuration int64
	CodedCacheCleanPeriod    time.Duration
}

// NewConfig parses the Config from the provided flags or environment variables.
func NewConfig(ctx *cli.Context) (*Config, error) {
	timeout, err := time.ParseDuration(ctx.GlobalString(flags.TimeoutFlag.Name))
	if err != nil {
		return &Config{}, err
	}

	return &Config{
		Hostname:                 ctx.GlobalString(flags.HostnameFlag.Name),
		GrpcPort:                 ctx.GlobalString(flags.GrpcPortFlag.Name),
		EnableMetrics:            ctx.GlobalBool(flags.EnableMetricsFlag.Name),
		MetricsPort:              ctx.GlobalString(flags.MetricsPortFlag.Name),
		Timeout:                  timeout,
		DbPath:                   ctx.GlobalString(flags.DbPathFlag.Name),
		PollingRetry:             ctx.GlobalUint64(flags.PollingRetryFlag.Name),
		GraphProvider:            ctx.GlobalString(flags.GraphProviderFlag.Name),
		DlsmAddress:              ctx.GlobalString(flags.DlsmAddressFlag.Name),
		ChainClientConfig:        chain.ReadCLIConfig(ctx),
		KzgConfig:                encoding.ReadCLIConfig(ctx),
		LoggingConfig:            logging.ReadCLIConfig(ctx),
		UseCache:                 ctx.GlobalBool(flags.UseCacheFlag.Name),
		CodedCacheSize:           ctx.GlobalUint64(flags.CodedCacheSizeFlag.Name),
		CodedCacheExpireDuration: ctx.GlobalInt64(flags.CodedCacheExpireDurationFlag.Name),
		CodedCacheCleanPeriod:    ctx.Duration(flags.CodedCacheCleanPeriodFlag.Name),
	}, nil
}
