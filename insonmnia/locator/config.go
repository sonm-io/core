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
	ListenAddr          string             `yaml:"address"`
	NodeTTL             time.Duration      `yaml:"node_ttl"`
	Eth                 accounts.EthConfig `required:"true" yaml:"ethereum"`
	OnlyPublicClientIPs bool               `required:"false" yaml:"only_public_client_ips"`
	Store               storeConfig        `required:"true" yaml:"store"`
	MetricsListenAddr   string             `yaml:"metrics_listen_addr" default:"127.0.0.1:14002"`
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
