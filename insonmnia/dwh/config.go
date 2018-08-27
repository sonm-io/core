package dwh

import (
	"errors"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/logging"
)

type DWHConfig struct {
	Logging           logging.Config     `yaml:"logging"`
	GRPCListenAddr    string             `yaml:"grpc_address" default:"127.0.0.1:15021"`
	HTTPListenAddr    string             `yaml:"http_address" default:"127.0.0.1:15022"`
	Eth               accounts.EthConfig `yaml:"ethereum" required:"true"`
	Storage           *storageConfig     `yaml:"storage" required:"true"`
	Blockchain        *blockchain.Config `yaml:"blockchain" required:"true"`
	MetricsListenAddr string             `yaml:"metrics_listen_addr" default:"127.0.0.1:14004"`
	ColdStart         bool               `yaml:"cold_start"`
	NumWorkers        int                `yaml:"num_workers" default:"64"`
}

type storageConfig struct {
	Endpoint string `required:"true" yaml:"endpoint"`
}

type YAMLConfig struct {
	Endpoint string `yaml:"endpoint" required:"false"`
}

func NewDWHConfig(path string) (*DWHConfig, error) {
	cfg := &DWHConfig{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	if cfg.NumWorkers < 1 {
		return nil, errors.New("at least one worker must be specified")
	}

	return cfg, nil
}

type L1ProcessorConfig struct {
	Storage    *storageConfig
	Blockchain *blockchain.Config
	NumWorkers int
	ColdStart  bool
}
