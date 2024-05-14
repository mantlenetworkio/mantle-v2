package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/Layr-Labs/datalayr/common/logging"
	retriever "github.com/Layr-Labs/datalayr/dl-retriever"
	"github.com/Layr-Labs/datalayr/dl-retriever/flags"
)

var (
	Version   = ""
	GitCommit = ""
	GitDate   = ""
)

func main() {

	app := cli.NewApp()
	app.Version = fmt.Sprintf("%s-%s-%s", Version, GitCommit, GitDate)
	app.Name = "dl-retriever"
	app.Usage = "DataLayr Retriever"
	app.Description = "Service for collecting coded chunks and decode the original data"
	app.Flags = flags.Flags

	app.Action = RetrieverMain
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("Application failed", "message", err)
	}

	select {}
}

func RetrieverMain(ctx *cli.Context) error {
	config, err := retriever.NewConfig(ctx)
	if err != nil {
		return err
	}

	logger, err := logging.GetLogger(config.LoggingConfig)
	if err != nil {
		return err
	}

	ret, err := retriever.NewRetriever(config, logger)
	if err != nil {
		return err
	}

	serverLogger := logger.Sublogger("Server")
	server := retriever.NewServer(config, ret, serverLogger)
	server.Start()

	return nil
}
