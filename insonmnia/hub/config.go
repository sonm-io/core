package hub

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type HubConfig struct {
	GRPCEndpoint  string   `yaml:"grpc_endpoint"`
	MinerEndpoint string   `yaml:"grpc_endpoint"`
	Bootnodes     []string `yaml:"bootnodes"`
}

func (conf *HubConfig) validate() error {
	var errs []string
	if len(conf.GRPCEndpoint) == 0 {
		errs = append(errs, "GRPC Endpoint is required")
	}
	if len(conf.MinerEndpoint) == 0 {
		errs = append(errs, "Miner Endpoint is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("Config validation error: %s", strings.Join(errs, ";"))
	}

	return nil
}

func NewConfig(path string) (*HubConfig, error) {
	conf, err := loadConfigFromFile(path)
	// able to add some default values here
	return conf, err
}

func loadConfigFromFile(path string) (*HubConfig, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := &HubConfig{}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		return nil, err
	}

	err = conf.validate()
	if err != nil {
		return nil, err
	}

	return conf, nil
}
