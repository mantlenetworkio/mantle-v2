package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/Layr-Labs/datalayr/common/logging"
	disperser "github.com/Layr-Labs/datalayr/dl-disperser"
	"github.com/Layr-Labs/datalayr/dl-disperser/flags"
)

var (
	Version   = ""
	GitCommit = ""
	GitDate   = ""
)

func main() {

	app := cli.NewApp()
	app.Flags = flags.Flags
	app.Version = fmt.Sprintf("%s-%s-%s", Version, GitCommit, GitDate)
	app.Name = "dl-disperser"
	app.Usage = "DataLayr Disperser"
	app.Description = "Service for encoding blobs and distributing coded chunks to nodes"

	app.Action = DisperserMain
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("Application failed.", "Message:", err)
	}

	select {}
}

func DisperserMain(ctx *cli.Context) error {
	log.Println("Initializing Disperser")

	config, err := disperser.NewConfig(ctx)
	if err != nil {
		return err
	}

	// Instantiate logger
	logger, err := logging.GetLogger(config.LoggingConfig)
	if err != nil {
		return err
	}

	dis, err := disperser.NewDisperser(config, logger)
	if err != nil {
		return err
	}

	err = dis.Start(context.Background())
	if err != nil {
		return err
	}

	// Creates the GRPC server
	serverLogger := logger.Sublogger("Server")
	server := disperser.NewServer(config, dis, serverLogger)
	server.Start()

	return nil
}
