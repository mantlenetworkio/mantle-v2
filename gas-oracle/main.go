package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/optimism/gas-oracle/flags"
	ometrics "github.com/ethereum-optimism/optimism/gas-oracle/metrics"
	"github.com/ethereum-optimism/optimism/gas-oracle/oracle"
	opservice "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
)

var (
	GitVersion = ""
	GitCommit  = ""
	GitDate    = ""
)

func main() {
	app := cli.NewApp()
	app.Flags = flags.Flags

	app.Version = opservice.FormatVersion(GitVersion, GitCommit, GitDate, "")
	app.Name = "gas-oracle"
	app.Usage = "Remotely Control the Optimism Gas Price"
	app.Description = "Configure with a private key and an Optimism HTTP endpoint " +
		"to send transactions that update the L2 gas price."

	// Configure the logging
	app.Before = func(ctx *cli.Context) error {
		loglevel := ctx.Generic(flags.LogLevelFlag.Name).(*oplog.LevelFlagValue).Level()
		oplog.SetGlobalLogHandler(log.NewTerminalHandlerWithLevel(os.Stdout, loglevel, isatty.IsTerminal(os.Stderr.Fd())))
		return nil
	}

	// Define the functionality of the application
	app.Action = func(ctx *cli.Context) error {
		if args := ctx.Args(); args.Len() > 0 {
			return fmt.Errorf("invalid command: %q", args.Get(0))
		}

		config := oracle.NewConfig(ctx)
		gpo, err := oracle.NewGasPriceOracle(config)
		if err != nil {
			return err
		}

		if config.MetricsEnabled {
			address := fmt.Sprintf("%s:%d", config.MetricsHTTP, config.MetricsPort)
			log.Info("Enabling stand-alone metrics HTTP endpoint", "address", address)
			ometrics.Setup(address)
			ometrics.InitAndRegisterStats(ometrics.DefaultRegistry)
		}

		if err := gpo.Start(); err != nil {
			return err
		}

		gpo.Wait()

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Crit("application failed", "message", err)
	}
}
