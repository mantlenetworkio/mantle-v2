package main

import (
	"fmt"
	"os"

	"flag"

	"github.com/Layr-Labs/datalayr/integration/deploy"
)

const (
	pathFlagName = "path"
	pathEnvName  = "EIGENDA_EXPERIMENT_PATH"
)

var pathFlag string

func init() {
	flag.StringVar(&pathFlag, pathFlagName, "", "path at which to read config. Alternatively, set the "+pathEnvName+" environment variable")
}

func main() {

	flag.Parse()

	if pathFlag == "" {
		pathFlag = os.Getenv(pathEnvName)
	}

	if pathFlag == "" {
		fmt.Println("No config path received. Exiting.")
		return
	}

	testEnv := deploy.NewTestConfig(pathFlag)
	testEnv.DeployExperiment()

}
