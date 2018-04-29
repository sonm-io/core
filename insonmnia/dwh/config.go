package dwh

import (
	"github.com/jinzhu/configor"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
)

type Config struct {
	Logging           LoggingConfig      `yaml:"logging"`
	GRPCListenAddr    string             `yaml:"grpc_address" default:"127.0.0.1:15021"`
	HTTPListenAddr    string             `yaml:"http_address" default:"127.0.0.1:15022"`
	Eth               accounts.EthConfig `yaml:"ethereum" required:"true"`
	Storage           *storageConfig     `yaml:"storage" required:"true"`
	Blockchain        *blockchainConfig  `yaml:"blockchain"`
	MetricsListenAddr string             `yaml:"metrics_listen_addr" default:"127.0.0.1:14004"`
}

type storageConfig struct {
	Backend  string `required:"true" yaml:"driver"`
	Endpoint string `required:"true" yaml:"endpoint"`
}

type blockchainConfig struct {
	EthEndpoint string `required:"true" yaml:"eth_endpoint"`
}

type LoggingConfig struct {
	Level *logging.Level `required:"true" default:"info"`
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	if _, ok := setupDBCallbacks[cfg.Storage.Backend]; !ok {
		return nil, errors.Errorf("backend `%s` is not supported", cfg.Storage.Backend)
	}

	return cfg, nil
}
