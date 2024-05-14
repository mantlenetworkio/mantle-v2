package trafficGen

import (
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/Layr-Labs/datalayr/traffic-gen/flags"
	"github.com/urfave/cli"
	"time"
)

type Config struct {
	Hostname              string
	GrpcPort              string
	Timeout               time.Duration
	StoreDuration         uint64
	DataSize              uint64
	LivenessThreshold     float64
	AdversarialThreshold  float64
	IdlePeriod            uint64
	IdlePeriodStd         uint64
	Number                uint64
	LoggingConfig         logging.Config
}

// NewConfig parses the Config from the provided flags or environment variables.
func NewConfig(ctx *cli.Context) Config {
	return Config{
		Hostname:              ctx.GlobalString(flags.HostnameFlag.Name),
		GrpcPort:              ctx.GlobalString(flags.GrpcPortFlag.Name),
		Timeout:               ctx.GlobalDuration(flags.TimeoutFlag.Name),
		StoreDuration:         ctx.GlobalUint64(flags.StoreDurationFlag.Name),
		DataSize:              ctx.GlobalUint64(flags.DataSizeFlag.Name),
		LivenessThreshold:     ctx.GlobalFloat64(flags.LivenessThresholdFlag.Name),
		AdversarialThreshold:  ctx.GlobalFloat64(flags.AdversarialThresholdFlag.Name),
		IdlePeriod:            ctx.GlobalUint64(flags.IdlePeriodFlag.Name),
		IdlePeriodStd:         ctx.GlobalUint64(flags.IdlePeriodStdFlag.Name),
		Number:                ctx.GlobalUint64(flags.NumberFlag.Name),
		LoggingConfig:         logging.ReadCLIConfig(ctx),
	}
}
