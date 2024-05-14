package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/Layr-Labs/datalayr/common/logging"
	node "github.com/Layr-Labs/datalayr/dl-node"

	"github.com/Layr-Labs/datalayr/dl-node/flags"
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
	app.Name = "dl-node"
	app.Usage = "DataLayr Node"
	app.Description = "Service for receiving and storing encoded blobs from disperser"

	app.Action = NodeMain
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("Application failed.", "Message:", err)
	}

	select {}
}

func NodeMain(ctx *cli.Context) error {
	log.Println("Initializing Node")
	config, err := node.NewConfig(ctx)
	if err != nil {
		return err
	}

	// Create the logger
	logger, err := logging.GetLogger(config.LoggingConfig)
	if err != nil {
		return err
	}

	dln, err := node.NewDln(config, logger)
	if err != nil {
		return err
	}

	err = dln.Start(context.Background())
	if err != nil {
		return err
	}

	// Creates the GRPC server
	serverLogger := logger.Sublogger("Server")
	server := node.NewServer(config, dln, serverLogger)
	server.Start()

	return nil
}
