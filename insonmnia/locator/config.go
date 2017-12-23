package locator

import (
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
)

type storeConfig struct {
	Type     string `required:"true" yaml:"type"`
	Endpoint string `required:"true" yaml:"endpoint"`
	Bucket   string `required:"true" yaml:"bucket"`
}

type Config struct {
	ListenAddr string             `yaml:"address"`
	NodeTTL    time.Duration      `yaml:"node_ttl"`
	Eth        accounts.EthConfig `required:"true" yaml:"ethereum"`
	Store      storeConfig        `required:"true" yaml:"store"`
}

// NewConfig loads a hub config from the specified YAML file.
func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func testConfig(addr string) *Config {
	return &Config{
		ListenAddr: addr,
		NodeTTL:    time.Second,
		Store: storeConfig{
			Type:     "boltdb",
			Endpoint: "/tmp/sonm/bolt-locator-test",
			Bucket:   "sonm",
		},
	}
}
