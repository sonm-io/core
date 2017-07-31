package hub

import (
	"github.com/jinzhu/configor"
)

type HubConfig struct {
	Logger struct {
		Level int `required:"true" default:"1"`
	} `yaml:"logger"`
	Hub struct {
		GRPCEndpoint  string   `required:"true" yaml:"grpc_endpoint"`
		MinerEndpoint string   `required:"true" yaml:"miner_endpoint"`
		Bootnodes     []string `required:"false" yaml:"bootnodes"`
	} `yaml:"hub"`
}

func NewConfig(path string) (*HubConfig, error) {
	conf := &HubConfig{}
	err := configor.Load(conf, path)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
