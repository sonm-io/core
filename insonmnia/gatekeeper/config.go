package gatekeeper

import (
	"fmt"
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/logging"
)

const (
	masterchainDirection = "masterchain"
	sidechainDirection   = "sidechain"
)

type gatekeeperConfig struct {
	Delay                time.Duration `yaml:"delay" default:"15s"`
	Period               time.Duration `yaml:"period" default:"15s"`
	ReloadFreezingPeriod time.Duration `yaml:"reload_freezing_period" default:"60m"`
	Direction            string        `yaml:"direction"`
}

type Config struct {
	Gatekeeper gatekeeperConfig   `yaml:"gatekeeper"`
	Log        logging.Config     `yaml:"log"`
	Blockchain *blockchain.Config `yaml:"blockchain"`
	Eth        accounts.EthConfig `yaml:"ethereum" required:"false"`
}

func (c *Config) validate() error {
	if c.Gatekeeper.Direction != masterchainDirection && c.Gatekeeper.Direction != sidechainDirection {
		return fmt.Errorf("direction field must be sidechain or masterchain")
	}
	return nil
}

// NewConfig loads localNode config from given .yaml file
func NewConfig(path string) (*Config, error) {
	cfg := &Config{}

	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}
