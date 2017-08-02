package miner

import (
	"github.com/jinzhu/configor"
)

// Config describes configuration of Miner
type Config struct {
	// TODO: move Logger section into common package
	Logger struct {
		Level int `required:"true" default:"1"`
	} `yaml:"logger"`
	Miner struct {
		HubAddress string     `required:"false" yaml:"hub_address"`
		Resources  *Resources `required:"false" yaml:"resources"`
	} `yaml:"miner"`
}

// NewConfig parses a configuration file pointed by path
func NewConfig(path string) (*Config, error) {
	conf := &Config{}
	err := configor.Load(conf, path)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
