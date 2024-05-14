package integration

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"testing"

	"github.com/Layr-Labs/datalayr/integration/deploy"
	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	newExperiment   bool
	disperserSocket string
	retrieverSocket string
	configPath      string

	testConfig *deploy.Config
)

func init() {
	flag.BoolVar(&newExperiment, "new-experiment", true, "Whether or not to redeploy Eigenlayer")
	flag.StringVar(&disperserSocket, "dis-socket", "", "Socket of the disperser")
	flag.StringVar(&retrieverSocket, "ret-socket", "", "Socket of the retriever")
	flag.StringVar(&configPath, "config", "", "Path to the config")
}

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	By("bootstrapping test environment")

	if configPath == "" {
		env, err := godotenv.Read(".env")
		if err != nil {
			log.Fatalf("Failed to read env file. Error: %s", err)
		}
		if _, ok := env["EXPERIMENT"]; ok {
			configPath = "./data/" + env["EXPERIMENT"]
		} else {
			log.Fatal("EXPERIMENT variables not defined")
		}
	}

	testConfig = deploy.NewTestConfig(configPath)
	fmt.Printf("New experiment: %v\n", newExperiment)
	if newExperiment && testConfig.Environment.IsLocal() {
		testConfig.DeployExperiment()
	}

	if testConfig.Environment.IsLocal() {
		startBinaries()
	}
})

var _ = AfterSuite(func() {
	if testConfig.Environment.IsLocal() {
		stopBinaries()
	}
})

// Start the binaries.
func startBinaries() {
	cmd := exec.Command(
		"./bin.sh",
		"start-detached")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Print(fmt.Sprint(err) + ": " + stderr.String())
		GinkgoT().Fatalf("Failed to start binaries. Err: %s", err.Error())
	} else {
		fmt.Print(out.String())
	}
}

// Stop the binaries.
func stopBinaries() {
	cmd := exec.Command(
		"./bin.sh",
		"stop")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Print(fmt.Sprint(err) + ": " + stderr.String())
		GinkgoT().Fatalf("Failed to stop binaries. Err: %s", err.Error())
	} else {
		fmt.Print(out.String())
	}
}

// Start the docker containers
func startNodes() {
	cmd := exec.Command(
		"./docker.sh",
		"start")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Print(fmt.Sprint(err) + ": " + stderr.String())
		GinkgoT().Fatalf("Failed to start docker containers. Err: %s", err.Error())
	} else {
		fmt.Print(out.String())
	}
}

// Stop the docker containers.
func stopNodes() {
	cmd := exec.Command(
		"./docker.sh",
		"stop")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Print(fmt.Sprint(err) + ": " + stderr.String())
		GinkgoT().Fatalf("Failed to stop docker containers. Err: %s", err.Error())
	} else {
		fmt.Print(out.String())
	}
}
