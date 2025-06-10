package main

import (
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	opservice "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/metrics/doc"
)

var (
	Version   = "v0.0.0"
	GitCommit = ""
	GitDate   = ""
)

func main() {
	oplog.SetupDefaults()

	app := cli.NewApp()
	app.Flags = flags.Flags
	app.Version = opservice.FormatVersion(Version, GitCommit, GitDate, "")
	app.Name = "op-batcher"
	app.Usage = "Batch Submitter Service"
	app.Description = "Service for generating and submitting L2 tx batches to L1"
	app.Action = curryMain(Version)
	app.Commands = []*cli.Command{
		{
			Name:        "doc",
			Subcommands: doc.NewSubcommands(metrics.NewMetrics("default")),
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Crit("Application failed", "message", err)
	}
}

// curryMain transforms the batcher.Main function into an app.Action
// This is done to capture the Version of the batcher.
func curryMain(version string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		return batcher.Main(version, ctx)
	}
}
