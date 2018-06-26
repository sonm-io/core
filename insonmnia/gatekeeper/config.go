package gatekeeper

import (
	"fmt"
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/logging"
)

type gatekeeperConfig struct {
	Delay                time.Duration `yaml:"delay" default:"15s"`
	Period               time.Duration `yaml:"period" default:"15s"`
	ReloadFreezingPeriod time.Duration `yaml:"reloadFreezingPeriod" default:"60m"`
	Direction            string        `yaml:"direction"`
}

type Config struct {
	Gatekeeper gatekeeperConfig   `yaml:"gatekeeper"`
	Log        logging.Config     `yaml:"log"`
	Blockchain *blockchain.Config `yaml:"blockchain"`
	Eth        accounts.EthConfig `yaml:"ethereum" required:"false"`
}

// NewConfig loads localNode config from given .yaml file
func NewConfig(path string) (*Config, error) {
	cfg := &Config{}

	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	if cfg.Gatekeeper.Direction != "masterchain" && cfg.Gatekeeper.Direction != "sidechain" {
		return nil, fmt.Errorf("direction field must be sidechain or masterchain")
	}
	return cfg, nil
}
