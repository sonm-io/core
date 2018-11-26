package tensor

import (
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/logging"
)

type Config struct {
	Logging           logging.Config     `yaml:"logging"`
	GRPCListenAddr    string             `yaml:"grpc_address" default:"127.0.0.1:15021"`
	HTTPListenAddr    string             `yaml:"http_address" default:"127.0.0.1:15022"`
	Eth               accounts.EthConfig `yaml:"ethereum" required:"true"`
	Blockchain        *blockchain.Config `yaml:"blockchain" required:"true"`
	MetricsListenAddr string             `yaml:"metrics_listen_addr" default:"127.0.0.1:14004"`
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

