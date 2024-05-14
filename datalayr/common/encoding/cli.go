package encoding

import (
	"runtime"

	"github.com/Layr-Labs/datalayr/common"
	kzgRs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
	"github.com/urfave/cli"
)

const (
	G1PathFlagName    = "kzg.g1-path"
	G2PathFlagName    = "kzg.g2-path"
	CachePathFlagName = "kzg.cache-path"
	SRSOrderFlagName  = "kzg.srs-order"
	NumWorkerFlagName = "kzg.num-workers"
	VerboseFlagName   = "kzg.verbose"
)

func CLIFlags(envPrefix string) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:     G1PathFlagName,
			Usage:    "Path to G1 SRS",
			Required: true,
			EnvVar:   common.PrefixEnvVar(envPrefix, "G1_PATH"),
		},
		cli.StringFlag{
			Name:     G2PathFlagName,
			Usage:    "Path to G2 SRS",
			Required: true,
			EnvVar:   common.PrefixEnvVar(envPrefix, "G2_PATH"),
		},
		cli.StringFlag{
			Name:     CachePathFlagName,
			Usage:    "Path to SRS Table directory",
			Required: true,
			EnvVar:   common.PrefixEnvVar(envPrefix, "CACHE_PATH"),
		},
		cli.Uint64Flag{
			Name:     SRSOrderFlagName,
			Usage:    "Order of the SRS",
			Required: true,
			EnvVar:   common.PrefixEnvVar(envPrefix, "SRS_ORDER"),
		},
		cli.Uint64Flag{
			Name:     NumWorkerFlagName,
			Usage:    "Number of workers for multithreading",
			Required: false,
			EnvVar:   common.PrefixEnvVar(envPrefix, "NUM_WORKERS"),
			Value:    uint64(runtime.GOMAXPROCS(0)),
		},
		cli.BoolFlag{
			Name:     VerboseFlagName,
			Usage:    "Enable to see verbose output for encoding/decoding",
			Required: false,
			EnvVar:   common.PrefixEnvVar(envPrefix, "VERBOSE"),
		},
	}
}

func ReadCLIConfig(ctx *cli.Context) kzgRs.KzgConfig {
	cfg := kzgRs.KzgConfig{}
	cfg.G1Path = ctx.GlobalString(G1PathFlagName)
	cfg.G2Path = ctx.GlobalString(G2PathFlagName)
	cfg.CacheDir = ctx.GlobalString(CachePathFlagName)
	cfg.SRSOrder = ctx.GlobalUint64(SRSOrderFlagName)
	cfg.NumWorker = ctx.GlobalUint64(NumWorkerFlagName)
	cfg.Verbose = ctx.GlobalBool(VerboseFlagName)
	return cfg
}
