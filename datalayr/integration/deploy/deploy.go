package deploy

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	disImage = "ghcr.io/layr-labs/datalayr/dl-disperser:latest"
	dlnImage = "ghcr.io/layr-labs/datalayr/dl-node:latest"
	retImage = "ghcr.io/layr-labs/datalayr/dl-retriever:latest"
	seqImage = "ghcr.io/layr-labs/datalayr/rollup-sequencer:latest"
	chaImage = "ghcr.io/layr-labs/datalayr/rollup-challenger:latest"
)

var (
	contractsLocation = "../contracts/eignlayr-contracts"
	// gethPrivateKeys   = "./geth-node/secret/private-keys.txt"
)

func (env *Config) getDeployer(name string) (ContractDeployer, bool) {

	for _, deployer := range env.Deployers {
		if deployer.Name == name {
			return deployer, true
		}
	}
	return ContractDeployer{}, false
}

// Constructs a mapping between service names/deployer names (e.g., 'dis0', 'dln1') and private keys. Order of priority: Map, List, File
func (env *Config) loadPrivateKeys() error {

	// construct full list of names
	// nTotal := env.Services.Counts.NumDis + env.Services.Counts.NumDln + env.Services.Counts.NumRet + env.Services.Counts.NumSeq + env.Services.Counts.NumCha
	// names := make([]string, len(env.Deployers)+nTotal)
	names := make([]string, 0)
	for _, d := range env.Deployers {
		names = append(names, d.Name)
	}
	addNames := func(prefix string, num int) {
		for i := 0; i < num; i++ {
			names = append(names, fmt.Sprintf("%v%v", prefix, i))
		}
	}
	addNames("dis", env.Services.Counts.NumDis)
	addNames("dln", env.Services.Counts.NumDln)
	addNames("staker", env.Services.Counts.NumDln)
	addNames("ret", env.Services.Counts.NumRet)
	addNames("seq", env.Services.Counts.NumRollupSeq)
	addNames("cha", env.Services.Counts.NumRollupCha)

	log.Println(names)

	// Collect private keys from file and list
	pks := env.Pks.List

	if env.Pks.File != "" {
		fileData := readFile(env.Pks.File)
		filePks := strings.Split(string(fileData), "\n")
		pks = append(pks, filePks...)
	}

	// Add missing items to map
	if env.Pks.Map == nil {
		env.Pks.Map = make(map[string]string)
	}

	ind := 0
	for _, name := range names {
		_, exists := env.Pks.Map[name]
		if !exists {

			if ind >= len(pks) {
				return errors.New("not enough pks")
			}

			env.Pks.Map[name] = pks[ind]
			ind++
		}
	}

	return nil
}

// Deploys the eigenlayer contracts to the chain.
// Returns a list of addresses corresponding to the contracts
func (env *Config) deployEigenlayer() {
	log.Print("Deploy the eigenlayer contracts")

	//deploy eigenlayer
	// get deployer
	deployer, ok := env.getDeployer(env.EigenLayer.Deployer)
	if !ok {
		log.Panicf("Deployer improperly configured")
	}

	// Change to contracts directory, defer back to integration
	changeDirectory("../contracts/eignlayr-contracts")

	createDirectory("data")
	fmt.Println("execForgeScript EigenLayrDeployer")
	execForgeScript("script/Deployer.s.sol:EigenLayrDeployer", env.Pks.Map[deployer.Name], deployer)
	fmt.Println("execForgeScript EigenLayrDeployer finished")
	addresses := make(map[string]string)
	contracts := []string{"investmentManager", "mantle", "mantleFirstStrat", "mantleSencodStrat", "delegation", "rgPermission"}
	for _, name := range contracts {
		data := readFile("data/" + name + ".addr")
		addresses[name] = string(data)
	}

	addressesJson, err := json.Marshal(addresses)
	if err != nil {
		log.Panicf("Error: %s", err.Error())
	} else {
		log.Print(string(addressesJson))
	}

	writeFile("data/addresses.json", addressesJson)

	//change back
	changeDirectory("../eignlayr-contracts")

	//add relevant addresses to path
	err = json.Unmarshal(addressesJson, &env.EigenLayer)
	if err != nil {
		log.Panicf("Error: %s", err.Error())
	}

	changeDirectory("../../integration")
	writeFile(env.Path+"/addresses.json", addressesJson)
}

func (env *Config) deployDatalayr() {
	log.Print("Deploy the datalayr contracts")

	//deploy eigenlayer
	// get deployer
	deployer, ok := env.getDeployer(env.EigenDA.Deployer)
	if !ok {
		log.Panicf("Deployer improperly configured")
	}

	//deploy datalayr
	changeDirectory("../contracts/datalayr-contracts")
	defer changeDirectory("../../integration")

	//copy over lib
	copyDirectory("../eignlayr-contracts", "./lib")

	addressesJson, err := json.Marshal(env.EigenLayer)
	if err != nil {
		log.Panicf("Error: %s", err.Error())
	} else {
		log.Print(string(addressesJson))
	}

	createDirectory("data")
	writeFile("data/datalayr_deploy_config.json", addressesJson)

	execForgeScript("script/Deployer.s.sol:DataLayrDeployer", env.Pks.Map[deployer.Name], deployer)

	//get new addresses
	addresses := make(map[string]string)
	contracts := []string{"dlReg", "dlsm", "pubkeyCompendium"}
	for _, name := range contracts {
		data := readFile("data/" + name + ".addr")
		addresses[name] = string(data)
	}

	addressesJson, err = json.Marshal(addresses)
	if err != nil {
		log.Panicf("Error: %s", err.Error())
	} else {
		log.Print(string(addressesJson))
	}

	//add relevant addresses to path
	err = json.Unmarshal(addressesJson, &env.EigenDA)
	if err != nil {
		log.Panicf("Error: %s", err.Error())
	}

	removeDirectory("./lib/eignlayr-contracts")
}

// Deploys the eigenlayer contracts to the chain.
// Returns a list of addresses corresponding to the contracts
func (env *Config) deployRollup() {
	log.Print("Deploy the rollup contracts")

	// get deployer
	deployer, ok := env.getDeployer(env.RollupExample.Deployer)
	if !ok {
		log.Panicf("Deployer improperly configured")
	}

	// Change to contracts directory, defer back to integration
	changeDirectory("../contracts/rollup-example-contracts")
	defer changeDirectory("../../integration")

	copyDirectory("../eignlayr-contracts", "lib/eignlayr-contracts")
	defer removeDirectory("lib/eignlayr-contracts")

	copyDirectory("../datalayr-contracts", "lib/datalayr-contracts")
	defer removeDirectory("lib/datalayr-contracts")
	copyDirectory("../eignlayr-contracts", "lib/datalayr-contracts/lib/eignlayr-contracts")

	createDirectory("data")

	addressesJson, err := json.Marshal(env.EigenDA)
	if err != nil {
		log.Panicf("Error: %s", err.Error())
	}
	writeFile("data/addresses.json", addressesJson)

	execForgeScript("script/Deploy.s.sol:Deploy", env.Pks.Map[deployer.Name], deployer)

	data := readFile("data/rollup.addr")
	env.RollupExample.DataLayrRollup = string(data)
}

func copyAbis(contractDir string, contracts []string) {
	for _, name := range contracts {
		dlJsonPath := "../contracts/" + contractDir + "/out/" + name + ".sol/" + name + ".json"
		dlJson := readFile(dlJsonPath)
		// Convert to map
		var dlMap map[string]interface{}
		err := json.Unmarshal([]byte(dlJson), &dlMap)
		if err != nil {
			log.Panicf("Error: %s", err.Error())
		}
		// Write to abi json
		dlAbi, err := json.Marshal(dlMap["abi"])
		if err != nil {
			log.Panicf("Error: %s", err.Error())
		}

		abiPath := "../subgraph/abis/" + name + ".json"

		removeFile(abiPath)
		writeFile(abiPath, dlAbi)
	}
}

// Deploys the subgraph
func (env *Config) updateSubgraph(startBlock int) {
	dlReg := env.EigenDA.DlRegistry
	dlsm := env.EigenDA.DlServiceManager
	pubkeyCompendium := env.EigenDA.PubkeyCompendium
	networkTemplateFile := "../subgraph/templates/networks.json"
	networkFile := "../subgraph/networks.json"
	subgraphTemplateFile := "../subgraph/templates/subgraph.yaml"
	subgraphFile := "../subgraph/subgraph.yaml"

	// Get network template file
	networkData := readFile(networkTemplateFile)

	// Convert to map
	var networkTemplate map[string]map[string]map[string]string
	err := json.Unmarshal([]byte(networkData), &networkTemplate)
	if err != nil {
		log.Panicf("Failed to unmarshal networks.json. Error: %s", err)
	}

	networkTemplate["mainnet"]["BLSRegistry"]["address"] = dlReg
	networkTemplate["mainnet"]["DataLayrServiceManager"]["address"] = dlsm
	networkTemplate["mainnet"]["BLSPublicKeyCompendium"]["address"] = pubkeyCompendium

	// Write to networks.json
	networkJson, err := json.Marshal(networkTemplate)
	if err != nil {
		log.Panicf("Error: %s", err.Error())
	}
	writeFile(networkFile, networkJson)
	log.Print("networks.json written")

	// Read subgraph template file
	subgraphTemplateData := readFile(subgraphTemplateFile)

	data := subgraph{}
	err = yaml.Unmarshal(subgraphTemplateData, &data)
	if err != nil {
		log.Panicf("Error %s:", err.Error())
	}

	// Need to remove 0x due to how yaml marshalling works with this library
	data.DataSources[0].Source.Address = dlReg[2:]
	data.DataSources[0].Source.StartBlock = startBlock

	data.DataSources[1].Source.Address = dlsm[2:]
	data.DataSources[1].Source.StartBlock = startBlock

	data.DataSources[2].Source.Address = pubkeyCompendium[2:]
	data.DataSources[2].Source.StartBlock = startBlock

	// Write to subgraph file
	subgraphYaml, err := yaml.Marshal(&data)
	if err != nil {
		log.Panic(err)
	}
	writeFile(subgraphFile, subgraphYaml)
	log.Print("subgraph.yaml written")

	contracts := []string{"DataLayrServiceManager", "BLSPublicKeyCompendium"}
	copyAbis("datalayr-contracts", contracts)
	contracts = []string{"BLSRegistry"}
	copyAbis("eignlayr-contracts", contracts)
}

// Deploys the subgraph
func (env *Config) deploySubgraph() {

	// Yarn commands
	changeDirectory("../subgraph")
	defer changeDirectory("../integration")

	execYarnCmd("codegen")
	execYarnCmd("remove-local")
	execYarnCmd("create-local")

	command := "echo 'v0.0.1' | yarn deploy-local"
	execBashCmd(command)
}

func (env *Config) applyDefaults(c any, prefix, stub string, ind int) {

	pv := reflect.ValueOf(c)
	v := pv.Elem()

	prefix += "_"

	for key, value := range env.Services.Variables["globals"] {
		field := v.FieldByName(prefix + key)
		if field.IsValid() && field.CanSet() {
			field.SetString(value)
		}
	}

	for key, value := range env.Services.Variables[stub] {
		field := v.FieldByName(prefix + key)
		fmt.Println(prefix + key)
		if field.IsValid() && field.CanSet() {
			field.SetString(value)
		}
	}

	for key, value := range env.Services.Variables[fmt.Sprintf("%v%v", stub, ind)] {
		field := v.FieldByName(prefix + key)
		if field.IsValid() && field.CanSet() {
			field.SetString(value)
		}
	}

}

// Generates disperser .env
func (env *Config) generateDisperserVars(ind int, key, address, logPath, dbPath, grpcPort, metricsPort string) DisperserVars {

	v := DisperserVars{
		DL_DISPERSER_HOSTNAME:       "",
		DL_DISPERSER_GRPC_PORT:      grpcPort,
		DL_DISPERSER_ENABLE_METRICS: "true",
		DL_DISPERSER_METRICS_PORT:   metricsPort,
		DL_DISPERSER_TIMEOUT:        "",
		DL_DISPERSER_POLLING_RETRY:  "1",
		DL_DISPERSER_DB_PATH:        dbPath,
		DL_DISPERSER_GRAPH_PROVIDER: "",
		DL_DISPERSER_CHAIN_RPC:      "",
		DL_DISPERSER_PRIVATE_KEY:    key[2:],
		DL_DISPERSER_CHAIN_ID:       "",
		DL_DISPERSER_DLSM_ADDRESS:   env.EigenDA.DlServiceManager,
		DL_DISPERSER_G1_PATH:        "",
		DL_DISPERSER_G2_PATH:        "",
		DL_DISPERSER_CACHE_PATH:     "",
		DL_DISPERSER_SRS_ORDER:      "",
		DL_DISPERSER_NUM_WORKERS:    fmt.Sprint(runtime.GOMAXPROCS(0)),
		DL_DISPERSER_STD_LOG_LEVEL:  "debug",
		DL_DISPERSER_FILE_LOG_LEVEL: "trace",
		DL_DISPERSER_LOG_PATH:       logPath,
		DL_DISPERSER_USE_CACHE:      "false",
		DL_DISPERSER_VERBOSE:        "true",
	}

	env.applyDefaults(&v, "DL_DISPERSER", "dis", ind)

	return v

}

// Generates datalayr node .env
func (env *Config) generateOperatorVars(ind int, key, address, logPath, dbPath, grpcPort, metricsPort string) OperatorVars {

	max, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	// max.Exp(big.NewInt(2), big.NewInt(130), nil).Sub(max, big.NewInt(1))

	//Generate cryptographically strong pseudo-random between 0 - max
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatal("Could not generate key")
	}

	//String representation of n in base 32
	blsKey := n.Text(10)

	v := OperatorVars{
		DL_NODE_HOSTNAME:        "",
		DL_NODE_GRPC_PORT:       grpcPort,
		DL_NODE_ENABLE_METRICS:  "true",
		DL_NODE_METRICS_PORT:    metricsPort,
		DL_NODE_TIMEOUT:         "",
		DL_NODE_DB_PATH:         dbPath,
		DL_NODE_CHAIN_ID:        "",
		DL_NODE_CHAIN_RPC:       "",
		DL_NODE_GRAPH_PROVIDER:  "",
		DL_NODE_PRIVATE_KEY:     key[2:],
		DL_NODE_PRIVATE_BLS:     blsKey,
		DL_NODE_DLSM_ADDRESS:    env.EigenDA.DlServiceManager,
		DL_NODE_G1_PATH:         "",
		DL_NODE_G2_PATH:         "",
		DL_NODE_CACHE_PATH:      "",
		DL_NODE_SRS_ORDER:       "",
		DL_NODE_NUM_WORKERS:     fmt.Sprint(runtime.GOMAXPROCS(0)),
		DL_NODE_CHALLENGE_ORDER: "",
		DL_NODE_STD_LOG_LEVEL:   "debug",
		DL_NODE_FILE_LOG_LEVEL:  "trace",
		DL_NODE_LOG_PATH:        logPath,
		DL_NODE_VERBOSE:         "true",
	}

	env.applyDefaults(&v, "DL_NODE", "dln", ind)

	return v

}

// Generates datalayr node .env
func (env *Config) generateRetrieverVars(ind int, key, address, logPath, grpcPort, apiPort string) RetrieverVars {

	v := RetrieverVars{
		DL_RETRIEVER_HOSTNAME:       "",
		DL_RETRIEVER_GRPC_PORT:      grpcPort,
		DL_RETRIEVER_TIMEOUT:        "",
		DL_RETRIEVER_GRAPH_PROVIDER: "",
		DL_RETRIEVER_DLSM_ADDRESS:   env.EigenDA.DlServiceManager,
		DL_RETRIEVER_CHAIN_RPC:      "",
		DL_RETRIEVER_PRIVATE_KEY:    key[2:],
		DL_RETRIEVER_CHAIN_ID:       "",
		DL_RETRIEVER_STD_LOG_LEVEL:  "debug",
		DL_RETRIEVER_FILE_LOG_LEVEL: "trace",
		DL_RETRIEVER_LOG_PATH:       logPath,
		DL_RETRIEVER_G1_PATH:        "",
		DL_RETRIEVER_G2_PATH:        "",
		DL_RETRIEVER_CACHE_PATH:     "",
		DL_RETRIEVER_SRS_ORDER:      "",
		DL_RETRIEVER_NUM_WORKERS:    fmt.Sprint(runtime.GOMAXPROCS(0)),
		DL_RETRIEVER_VERBOSE:        "true",
	}

	env.applyDefaults(&v, "DL_RETRIEVER", "ret", ind)

	return v

}

func (env *Config) generateSequencerVars(ind int, key, address, logPath string, grpcPort int) RollupSequencerVars {

	v := RollupSequencerVars{
		SEQUENCER_DISPERSER:      fmt.Sprintf("%v:%v", env.Dispersers[0].DL_DISPERSER_HOSTNAME, env.Dispersers[0].DL_DISPERSER_GRPC_PORT),
		SEQUENCER_GRPC_PORT:      fmt.Sprint(grpcPort),
		SEQUENCER_CHAIN_ID:       "",
		SEQUENCER_CHAIN_RPC:      "",
		SEQUENCER_GRAPH_PROVIDER: "",
		SEQUENCER_PRIVATE_KEY:    key[2:],
		SEQUENCER_ROLLUP_ADDRESS: env.RollupExample.DataLayrRollup,
		SEQUENCER_DURATION:       "1",
		SEQUENCER_TIMEOUT:        "30s",
		SEQUENCER_STD_LOG_LEVEL:  "debug",
		SEQUENCER_FILE_LOG_LEVEL: "trace",
		SEQUENCER_LOG_PATH:       logPath,
	}

	env.applyDefaults(&v, "SEQUENCER", "seq", ind)

	return v
}

func (env *Config) generateChallengerVars(ind int, key, address, logPath string, grpcPort int) RollupChallengerVars {

	v := RollupChallengerVars{
		CHALLENGER_RETRIEVER:       fmt.Sprintf("%v:%v", env.Retrievers[0].DL_RETRIEVER_HOSTNAME, env.Retrievers[0].DL_RETRIEVER_GRPC_PORT),
		CHALLENGER_CHAIN_ID:        "",
		CHALLENGER_CHAIN_RPC:       "",
		CHALLENGER_GRAPH_PROVIDER:  "",
		CHALLENGER_PRIVATE_KEY:     key[2:],
		CHALLENGER_ROLLUP_ADDRESS:  env.RollupExample.DataLayrRollup,
		CHALLENGER_G1_PATH:         "",
		CHALLENGER_G2_PATH:         "",
		CHALLENGER_CACHE_PATH:      "",
		CHALLENGER_SRS_ORDER:       "",
		CHALLENGER_KZG_NUM_WORKERS: fmt.Sprint(runtime.GOMAXPROCS(0)),
		CHALLENGER_STD_LOG_LEVEL:   "debug",
		CHALLENGER_FILE_LOG_LEVEL:  "trace",
		CHALLENGER_LOG_PATH:        logPath,
		CHALLENGER_TIMEOUT:         "30s",
	}

	env.applyDefaults(&v, "CHALLENGER", "cha", ind)

	return v

}

// Used to generate a docker compose file corresponding to the test environment
func genService(compose testbed, name, image, envFile, grpcPort string, command []string) {
	newDir, err := os.Getwd()
	if err != nil {
		log.Panicf("Failed to get working directory. Error: %s", err)
	}

	compose.Services[name] = Service{
		Image:   image,
		EnvFile: []string{envFile},
		Ports: []string{
			grpcPort + ":" + grpcPort,
		},
		Volumes: []string{newDir + "/data:/data"},
		Command: command,
	}
}

func (env *Config) getPaths(name string) (logPath, dbPath, envFilename, envFile string) {

	if env.Environment.IsLocal() {
		logPath = ""
		dbPath = env.Path + "/db/" + name
	} else {
		logPath = "/data/logs/" + name
		dbPath = "/data/db/" + name
	}

	envFilename = "envs/" + name + ".env"
	envFile = env.Path + "/" + envFilename
	return
}

func (env *Config) getKey(name string) (key, address string) {
	key = env.Pks.Map[name]
	log.Printf("key: %v", key)
	address = GetAddress(key)
	return
}

// Generates all of the config for the test environment.
// Returns an object that corresponds to the participants of the
// current experiment.
func (env *Config) GenerateAllVariables() {
	// Create envs directory
	createDirectory(env.Path + "/envs")

	// Gather keys
	// keyData := readFile(gethPrivateKeys)
	// keys := strings.Split(string(keyData), "\n")
	// id := 1

	// Create compose file
	composeFile := env.Path + "/docker-compose.yml"
	servicesMap := make(map[string]Service)
	compose := testbed{
		Services: servicesMap,
	}

	// Create participants
	port := env.Services.BasePort

	// Generate disperser nodes
	for i := 0; i < env.Services.Counts.NumDis; i++ {
		metricsPort := 9091 // port
		grpcPort := port + 1
		port += 2

		name := fmt.Sprintf("dis%v", i)
		logPath, dbPath, filename, envFile := env.getPaths(name)
		key, address := env.getKey(name)

		// Convert key to address
		config := env.generateDisperserVars(i, key, address, logPath, dbPath, fmt.Sprint(grpcPort), fmt.Sprint(metricsPort))
		writeEnv(config.getEnvMap(), envFile)
		env.Dispersers = append(env.Dispersers, config)

		genService(
			compose, name, disImage,
			filename, fmt.Sprint(grpcPort), []string{})
	}

	for i := 0; i < env.Services.Counts.NumDln; i++ {
		metricsPort := 9091 // port
		grpcPort := port + 1
		port += 2

		name := fmt.Sprintf("dln%v", i)
		logPath, dbPath, filename, envFile := env.getPaths(name)
		key, address := env.getKey(name)

		// Convert key to address

		config := env.generateOperatorVars(i, key, address, logPath, dbPath, fmt.Sprint(grpcPort), fmt.Sprint(metricsPort))
		writeEnv(config.getEnvMap(), envFile)
		env.Operators = append(env.Operators, config)

		genService(
			compose, name, dlnImage,
			filename, fmt.Sprint(grpcPort), []string{})

	}

	// Stakers
	for i := 0; i < env.Services.Counts.NumDln; i++ {

		name := fmt.Sprintf("staker%v", i)
		key, address := env.getKey(name)

		// Create staker paritipants
		participant := Participant{
			Address: address,
			Private: key[2:],
		}
		env.Stakers = append(env.Stakers, participant)
	}

	for i := 0; i < env.Services.Counts.NumRet; i++ {
		apiPort := port
		grpcPort := port + 1
		port += 2

		name := fmt.Sprintf("ret%v", i)
		logPath, _, filename, envFile := env.getPaths(name)
		key, address := env.getKey(name)

		config := env.generateRetrieverVars(i, key, address, logPath, fmt.Sprint(grpcPort), fmt.Sprint(apiPort))
		env.Retrievers = append(env.Retrievers, config)
		writeEnv(config.getEnvMap(), envFile)

		genService(
			compose, name, retImage,
			filename, fmt.Sprint(grpcPort), []string{})

	}

	for i := 0; i < env.Services.Counts.NumRollupSeq; i++ {
		grpcPort := port + 1
		port += 1

		name := fmt.Sprintf("seq%v", i)
		logPath, _, filename, envFile := env.getPaths(name)
		key, address := env.getKey(name)

		config := env.generateSequencerVars(i, key, address, logPath, grpcPort)
		env.Sequencers = append(env.Sequencers, config)
		writeEnv(config.getEnvMap(), envFile)

		genService(
			compose, name, seqImage, filename, fmt.Sprint(grpcPort), []string{})
	}

	for i := 0; i < env.Services.Counts.NumRollupCha; i++ {
		grpcPort := port + 1
		port += 1

		name := fmt.Sprintf("cha%v", i)
		logPath, _, filename, envFile := env.getPaths(name)
		key, address := env.getKey(name)

		config := env.generateChallengerVars(i, key, address, logPath, grpcPort)
		env.Challengers = append(env.Challengers, config)
		writeEnv(config.getEnvMap(), envFile)

		genService(
			compose, name, chaImage, filename, fmt.Sprint(grpcPort), []string{})
	}

	if env.Environment.IsLocal() {

		// Write to compose file
		composeYaml, err := yaml.Marshal(&compose)
		if err != nil {
			log.Panicf("Error: %s", err.Error())
		}
		writeFile(composeFile, composeYaml)
	}

}

func (env *Config) generateParticipantsJson() participants {

	dispersers := make([]Participant, 0)
	operators := make([]Participant, 0)
	stakers := make([]Staker, 0)

	total := float32(0)
	stakes := make([]float32, len(env.Services.Stakes.Distribution))
	for _, stake := range env.Services.Stakes.Distribution {
		total += stake
	}
	for ind, stake := range env.Services.Stakes.Distribution {
		stakes[ind] = stake / total * env.Services.Stakes.Total
	}
	for name := range env.Pks.Map {
		key, address := env.getKey(name)
		switch name[:3] {
		case "dis", "seq", "cha":
			dispersers = append(dispersers, Participant{
				Address: address,
				Private: key,
			})
		case "dln":
			operators = append(operators, Participant{
				Address: address,
				Private: key,
			})
		case "sta":
			id, err := strconv.ParseUint(name[6:], 10, 64)
			if err != nil {
				log.Fatal("Could not parse id")
			}
			stakers = append(stakers, Staker{
				Participant: Participant{
					Address: address,
					Private: key,
				},
				Stake: strconv.FormatFloat(float64(stakes[id]), 'f', 0, 32),
			})
		}
	}

	participants := participants{
		Dispersers: dispersers,
		Nodes:      operators,
		Stakers:    stakers,
		NumDis:     env.Services.Counts.NumDis + env.Services.Counts.NumRollupSeq + env.Services.Counts.NumRollupCha,
		NumDln:     env.Services.Counts.NumDln,
		NumStaker:  env.Services.Counts.NumDln,
	}

	// Write to participants.json
	participantsJson, err := json.Marshal(participants)
	if err != nil {
		log.Panicf("Error: %s", err.Error())
	}
	writeFile(env.Path+"/participants.json", participantsJson)
	writeFile(contractsLocation+"/data/participants.json", participantsJson)

	return participants

}

// Initializes the participants by providing them with the required stake.
func (env *Config) InitializeParticipants(p participants) {
	changeDirectory(contractsLocation)
	defer changeDirectory("../../integration")

	deployer, ok := env.getDeployer(env.EigenLayer.Deployer)
	if !ok {
		log.Panic("deployer not properly configured")
	}
	time.Sleep(10 * time.Second)
	execForgeScript("script/Allocate.s.sol:Allocate", env.Pks.Map[deployer.Name], deployer)
	time.Sleep(10 * time.Second)
	fmt.Println("Nodes GrantPermission!!!")
	execForgeScript("script/GrantPermission.s.sol:GrantPermission", env.Pks.Map[deployer.Name], deployer)
	time.Sleep(30 * time.Second)
	var wg sync.WaitGroup
	fmt.Println("Nodes BecomeOperator!!!")
	// wg.Add(len(p.Nodes))
	for i := 0; i < len(p.Nodes); i++ {
		// go func(i int) {
		// 	defer wg.Done()
		// 	node_sk := p.Nodes[i].Private
		// 	execForgeScript("script/BecomeOperator.s.sol:BecomeOperator", node_sk, deployer)
		// }(i)
		fmt.Println("p.Nodes: ", i)
		node_sk := p.Nodes[i].Private
		execForgeScript("script/BecomeOperator.s.sol:BecomeOperator", node_sk, deployer)
		time.Sleep(5000)
	}

	// wg.Wait()
	time.Sleep(20 * time.Second) // sleep for data into graph node
	fmt.Println("staker DepositAndDelegate!!!")
	// wg.Add(len(p.Stakers))
	for i := 0; i < len(p.Stakers); i++ {
		// go func(i int) {
		// 	defer wg.Done()
		// 	staker_sk := p.Stakers[i].Private
		// 	execForgeScript("script/DepositAndDelegate.s.sol:DepositAndDelegate", staker_sk, deployer)
		// }(i)

		fmt.Println("p.Stakers: ", i)
		staker_sk := p.Stakers[i].Private
		execForgeScript("script/DepositAndDelegate.s.sol:DepositAndDelegate", staker_sk, deployer)
		time.Sleep(5000)
	}

	wg.Wait()
	log.Print("Finished initializing participants!")
}

// Deploys a datalayr experiment
func (env *Config) DeployExperiment() {

	defer env.SaveTestConfig()

	log.Print("Deploying experiment...")

	// Log to file
	f, err := os.OpenFile(env.Path+"/deploy.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Panicf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Create a new experiment and deploy the contracts

	err = env.loadPrivateKeys()
	if err != nil {
		log.Panicf("could not load private keys: %v", err)
	}

	startBlock := GetLatestBlockNumber(env.Deployers[0].Rpc)

	if env.EigenLayer.Deployer != "" && !env.IsEigenLayerDeployed() {
		fmt.Println("Deploying EigenLayer")
		env.deployEigenlayer()
	}
	time.Sleep(20 * time.Second)
	if env.EigenDA.Deployer != "" && !env.IsEigenDADeployed() {
		fmt.Println("Deploying EigenDA")
		env.deployDatalayr()
		time.Sleep(3 * time.Second)
		fmt.Println("Initial allocations")
		participants := env.generateParticipantsJson()
		env.InitializeParticipants(participants)
		time.Sleep(3 * time.Second)
		fmt.Println("Deploying Subgraph")
		env.updateSubgraph(startBlock)
		time.Sleep(3 * time.Second)
		if env.Environment.IsLocal() {
			env.deploySubgraph()
		}
	}

	if env.RollupExample.Deployer != "" && !env.IsRollupDeployed() {
		fmt.Println("Deploying Rollup")
		env.deployRollup()
	}
	time.Sleep(3 * time.Second)
	fmt.Println("Generating variables")
	env.GenerateAllVariables()

	fmt.Println("Test environment has succesfully deployed!")

}
