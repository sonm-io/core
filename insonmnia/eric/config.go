package eric

import (
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/logging"
)

type ericConfig struct{}

type Config struct {
	Eric       ericConfig         `yaml:"eric" required:"false"`
	Log        logging.Config     `yaml:"log"`
	Blockchain *blockchain.Config `yaml:"blockchain"`
	Eth        accounts.EthConfig `yaml:"ethereum" required:"false"`
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{}

	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
