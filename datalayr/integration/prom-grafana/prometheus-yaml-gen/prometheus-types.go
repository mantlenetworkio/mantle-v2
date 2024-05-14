package promYamlGen

// Subgraph yaml
type config struct {
	Global              Global         `yaml:"global"`
	ScrapeConfigs       []ScrapeConfig `yaml:"scrape_configs"`
}

type Global struct {
	ScrapeInterval string  `yaml:"scrape_interval"`
	EvaluationInterval string `yaml:"evaluation_interval"`
}

type ScrapeConfig struct {
	JobName       string `yaml:"job_name"`
	StaticConfigs	[]StaticConfig `yaml:"static_configs"`
}

type StaticConfig struct {
	Targets      []string `yaml:"targets"`
}
