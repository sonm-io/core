package miner

import (
	"github.com/jinzhu/configor"
)

type MinerConfig struct {
	Logger struct {
		Level int `required:"true" default:"1"`
	} `yaml:"logger"`
	Miner struct {
		HubAddress string `required:"false" yaml:"hub_address"`
	} `yaml:"miner"`
}

func NewConfig(path string) (*MinerConfig, error) {
	conf := &MinerConfig{}
	err := configor.Load(conf, path)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
