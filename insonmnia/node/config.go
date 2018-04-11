package node

import (
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/npp"
)

type nodeConfig struct {
	HttpBindPort uint16 `yaml:"http_bind_port" default:"15031"`
	BindPort     uint16 `yaml:"bind_port" default:"15030"`
}

type marketConfig struct {
	Endpoint string `required:"true" yaml:"endpoint"`
}

type hubConfig struct {
	Endpoint string `required:"false" yaml:"endpoint"`
}

type Config struct {
	Node              nodeConfig         `yaml:"node"`
	NPPCfg            npp.Config         `yaml:"npp"`
	Market            marketConfig       `required:"true" yaml:"market"`
	Log               logging.Config     `yaml:"log"`
	Eth               accounts.EthConfig `required:"false" yaml:"ethereum"`
	Hub               hubConfig          `required:"false" yaml:"hub"`
	MetricsListenAddr string             `yaml:"metrics_listen_addr" default:"127.0.0.1:14003"`
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
