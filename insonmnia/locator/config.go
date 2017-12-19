package locator

import (
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
)

type Config struct {
	ListenAddr    string             `yaml:"address"`
	NodeTTL       time.Duration      `yaml:"node_ttl"`
	CleanupPeriod time.Duration      `yaml:"cleanup_period"`
	Eth           accounts.EthConfig `required:"true" yaml:"ethereum"`
}

// NewConfig loads a config options from the specified YAML file.
func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func DefaultConfig(addr string) *Config {
	return &Config{
		ListenAddr:    addr,
		NodeTTL:       time.Hour,
		CleanupPeriod: time.Minute,
	}
}
