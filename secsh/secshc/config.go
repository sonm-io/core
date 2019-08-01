package secshc

import (
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/npp"
)

type RPTYConfig struct {
	Eth accounts.EthConfig `yaml:"ethereum"`
	NPP npp.Config         `yaml:"npp"`
}

func NewRPTYConfig(path string) (*RPTYConfig, error) {
	cfg := &RPTYConfig{}
	if err := configor.Load(cfg, path); err != nil {
		return nil, err
	}

	return cfg, nil
}
