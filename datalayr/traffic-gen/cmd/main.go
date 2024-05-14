package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"

	tg "github.com/Layr-Labs/datalayr/traffic-gen"
	"github.com/Layr-Labs/datalayr/traffic-gen/flags"
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
	app.Name = "traffic-generator"
	app.Usage = "DataLayr Disperser Traffic Generator"
	app.Description = "Data Sources for generating pseudo traffic to dispersers nodes"

	app.Action = tg.GetRunner(Version)
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("Application failed", "message", err)
	}

	select{}
}
