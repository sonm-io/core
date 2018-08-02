package node

import (
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/dwh"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/matcher"
	"github.com/sonm-io/core/insonmnia/npp"
	"github.com/sonm-io/core/util/config"
)

type nodeConfig struct {
	HttpBindPort            uint16 `yaml:"http_bind_port" default:"15031"`
	BindPort                uint16 `yaml:"bind_port" default:"15030"`
	AllowInsecureConnection bool   `yaml:"allow_insecure_connection" default:"false"`
}

type Config struct {
	Node              nodeConfig          `yaml:"node"`
	NPP               npp.Config          `yaml:"npp"`
	Log               logging.Config      `yaml:"log"`
	Blockchain        *blockchain.Config  `yaml:"blockchain"`
	Eth               accounts.EthConfig  `yaml:"ethereum" required:"false"`
	DWH               dwh.YAMLConfig      `yaml:"dwh"`
	MetricsListenAddr string              `yaml:"metrics_listen_addr" default:"127.0.0.1:14003"`
	Benchmarks        benchmarks.Config   `yaml:"benchmarks"`
	Matcher           *matcher.YAMLConfig `yaml:"matcher"`
}

// NewConfig loads localNode config from given .yaml file
func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	if err := config.FromFile(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
