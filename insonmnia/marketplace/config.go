package marketplace

import (
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
)

type MarketplaceConfig struct {
	ListenAddr string             `yaml:"address"`
	Eth        accounts.EthConfig `required:"true" yaml:"ethereum"`
}

func NewConfig(path string) (*MarketplaceConfig, error) {
	cfg := &MarketplaceConfig{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
