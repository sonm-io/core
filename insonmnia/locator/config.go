package locator

import (
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
)

type LocatorConfig struct {
	ListenAddr    string             `yaml:"address"`
	NodeTTL       time.Duration      `yaml:"node_ttl"`
	CleanupPeriod time.Duration      `yaml:"cleanup_period"`
	Eth           accounts.EthConfig `required:"true" yaml:"ethereum"`
}

// NewConfig loads a hub config from the specified YAML file.
func NewConfig(path string) (*LocatorConfig, error) {
	cfg := &LocatorConfig{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func DefaultConfig(addr string) *LocatorConfig {
	return &LocatorConfig{
		ListenAddr:    addr,
		NodeTTL:       time.Hour,
		CleanupPeriod: time.Minute,
	}
}
