package oracle

import (
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/logging"
)

type oracleConfig struct {
	Mode                 bool          `yaml:"isMaster" default:"false"`
	PriceUpdatePeriod    time.Duration `yaml:"priceUpdatePeriod" default:"15s"`
	ContractUpdatePeriod time.Duration `yaml:"contractUpdatePeriod" default:"15m"`
	Percent              float64       `yaml:"deviationPercent" required:"1.0"`
}

type Config struct {
	Oracle     oracleConfig       `yaml:"oracle"`
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
	return cfg, nil
}
