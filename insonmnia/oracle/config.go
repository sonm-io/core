package oracle

import (
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/logging"
)

type oracleConfig struct {
	IsMaster             bool          `yaml:"is_master" default:"false"`
	PriceUpdatePeriod    time.Duration `yaml:"price_update_period" default:"15s"`
	ContractUpdatePeriod time.Duration `yaml:"contract_update_period" default:"15m"`
	Percent              float64       `yaml:"deviation_percent" default:"1.0"`
	FromNow              bool          `yaml:"from_now" default:"true"`
}

type Config struct {
	Oracle     oracleConfig       `yaml:"oracle"`
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
