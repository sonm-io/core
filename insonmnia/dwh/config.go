package dwh

import (
	"github.com/jinzhu/configor"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/logging"
)

type Config struct {
	Logging           LoggingConfig      `yaml:"logging"`
	GRPCListenAddr    string             `yaml:"grpc_address" default:"127.0.0.1:15021"`
	HTTPListenAddr    string             `yaml:"http_address" default:"127.0.0.1:15022"`
	Eth               accounts.EthConfig `yaml:"ethereum" required:"true"`
	Storage           *storageConfig     `yaml:"storage" required:"true"`
	Blockchain        *blockchain.Config `yaml:"blockchain"`
	MetricsListenAddr string             `yaml:"metrics_listen_addr" default:"127.0.0.1:14004"`
	ColdStart         *ColdStartConfig   `yaml:"cold_start"`
	NumWorkers        int                `yaml:"num_workers" default:"64"`
}

type storageConfig struct {
	Backend  string `required:"true" yaml:"driver"`
	Endpoint string `required:"true" yaml:"endpoint"`
}

type LoggingConfig struct {
	Level *logging.Level `required:"true" default:"warn"`
}

type ColdStartConfig struct {
	UpToBlock uint64 `yaml:"up_to_block"`
}

type YAMLConfig struct {
	Endpoint string `yaml:"endpoint" required:"false"`
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	if cfg.NumWorkers < 1 {
		return nil, errors.New("at least one worker must be specified")
	}

	return cfg, nil
}
