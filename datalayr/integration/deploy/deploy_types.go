package deploy

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// participants.json
type participants struct {
	Dispersers []Participant `json:"dis"`
	Nodes      []Participant `json:"dln"`
	Stakers    []Staker      `json:"staker"`

	NumDis    int `json:"numDis"`
	NumDln    int `json:"numDln"`
	NumStaker int `json:"numStaker"`
}

type Participant struct {
	Address string `json:"address"`
	Private string `json:"private"`
}

type Staker struct {
	Participant
	Stake string `json:"stake"`
}

// Docker compose
type testbed struct {
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	Image   string   `yaml:"image"`
	Volumes []string `yaml:"volumes"`
	Ports   []string `yaml:"ports"`
	EnvFile []string `yaml:"env_file"`
	Command []string `yaml:"command"`
}

type EnvList map[string]string

type ContractDeployer struct {
	Name            string `yaml:"name"`
	Rpc             string `yaml:"rpc"`
	VerifierUrl     string `yaml:"verifierUrl"`
	VerifyContracts bool   `yaml:"verifyContracts"`
	Slow            bool   `yaml:slow`
	// PrivateKey string `yaml:"private_key"`
}

type EigenLayerContract struct {
	Deployer          string `yaml:"deployer"`
	Delegation        string `json:"delegation"`
	InvestmentManager string `json:"investmentManager"`
	Mantle            string `json:"mantle"`
	MantleSencodStrat string `json:"mantleSencodStrat"`
	MantleFirstStrat  string `json:"mantleFirstStrat"`
	RgPermission      string `json:"rgPermission"`
}

type EigenDAContract struct {
	Deployer         string `yaml:"deployer"`
	DlRegistry       string `json:"dlReg"`
	DlServiceManager string `json:"dlsm"`
	PubkeyCompendium string `json:"pubkeyCompendium"`
}

type RollupExampleContract struct {
	Deployer       string `yaml:"deployer"`
	DataLayrRollup string `yaml:"datalayrRollup"`
}

type ServicesSpec struct {
	Counts struct {
		NumDis       int `yaml:"dispersers"`
		NumDln       int `yaml:"operators"`
		NumRet       int `yaml:"retrievers"`
		NumRollupSeq int `yaml:"rollupSequencers"`
		NumRollupCha int `yaml:"rollupChallengers"`
	} `yaml:"counts"`
	Stakes struct {
		Total        float32   `yaml:"total"`
		Distribution []float32 `yaml:"distribution"`
	} `yaml:"stakes"`
	BasePort  int       `yaml:"basePort"`
	Variables Variables `yaml:"variables"`
}

type Variables map[string]map[string]string

type PkConfig struct {
	File string            `yaml:"file"`
	List []string          `yaml:"list"`
	Map  map[string]string `yaml:"map"`
}

type Environment struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

func (e Environment) IsLocal() bool {
	return e.Type == "local"
}

type Config struct {
	Path string

	Environment Environment `yaml:"environment"`

	Deployers []ContractDeployer `yaml:"deployers"`

	EigenLayer    EigenLayerContract    `yaml:"eigenlayer"`
	EigenDA       EigenDAContract       `yaml:"eigenda"`
	RollupExample RollupExampleContract `yaml:"rollup"`

	Pks *PkConfig `yaml:"privateKeys"`

	Services ServicesSpec `yaml:"services"`

	Dispersers  []DisperserVars
	Operators   []OperatorVars
	Retrievers  []RetrieverVars
	Stakers     []Participant
	Sequencers  []RollupSequencerVars
	Challengers []RollupChallengerVars
}

func (c Config) IsEigenLayerDeployed() bool {
	return c.EigenLayer.InvestmentManager != ""
}

func (c Config) IsEigenDADeployed() bool {
	return c.EigenDA.DlServiceManager != ""
}

func (c Config) IsRollupDeployed() bool {
	return c.RollupExample.DataLayrRollup != ""
}

func NewTestConfig(path string) (testEnv *Config) {

	configPath := path + "/config.lock.yaml"
	if _, err := os.Stat(configPath); err != nil {
		configPath = path + "/config.yaml"

	}
	data := readFile(configPath)

	err := yaml.Unmarshal(data, &testEnv)
	if err != nil {
		log.Panicf("Error %s:", err.Error())
	}
	testEnv.Path = path

	return
}

func (env *Config) SaveTestConfig() {
	obj, _ := yaml.Marshal(env)
	writeFile(env.Path+"/config.lock.yaml", obj)
}
