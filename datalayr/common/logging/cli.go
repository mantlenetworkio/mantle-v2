package logging

import (
	"errors"

	"github.com/Layr-Labs/datalayr/common"
	"github.com/rs/zerolog"
	"github.com/urfave/cli"
)

const (
	PathFlagName      = "log.path"
	FileLevelFlagName = "log.level-file"
	StdLevelFlagName  = "log.level-std"
)

func CLIFlags(envPrefix string) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   StdLevelFlagName,
			Usage:  `The lowest log level that will be output to stdout. Accepted options are "trace", "debug", "info", "warn", "error"`,
			Value:  "info",
			EnvVar: common.PrefixEnvVar(envPrefix, "STD_LOG_LEVEL"),
		},
		cli.StringFlag{
			Name:   FileLevelFlagName,
			Usage:  `The lowest log level that will be output to file logs. Accepted options are "trace", "debug", "info", "warn", "error"`,
			Value:  "info",
			EnvVar: common.PrefixEnvVar(envPrefix, "FILE_LOG_LEVEL"),
		},
		cli.StringFlag{
			Name:   PathFlagName,
			Usage:  "Path to file where logs will be written",
			Value:  "text",
			EnvVar: common.PrefixEnvVar(envPrefix, "LOG_PATH"),
		},
	}
}

func DefaultCLIConfig() Config {
	return Config{
		Path:      "",
		FileLevel: "debug",
		StdLevel:  "debug",
	}
}

func (cfg Config) Check() error {

	_, err := zerolog.ParseLevel(cfg.FileLevel)
	if err != nil {
		return errors.New("unrecognized file log level")
	}
	_, err = zerolog.ParseLevel(cfg.StdLevel)
	if err != nil {
		return errors.New("unrecognized stdout log level")
	}
	return nil

}

func ReadCLIConfig(ctx *cli.Context) Config {
	cfg := DefaultCLIConfig()
	cfg.StdLevel = ctx.GlobalString(StdLevelFlagName)
	cfg.FileLevel = ctx.GlobalString(FileLevelFlagName)
	cfg.Path = ctx.GlobalString(PathFlagName)
	return cfg
}
