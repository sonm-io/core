package locator

import (
	"time"

	"github.com/jinzhu/configor"
)

type LocatorConfig struct {
	ListenAddr    string        `yaml:"address"`
	NodeTTL       time.Duration `yaml:"node_ttl"`
	CleanupPeriod time.Duration `yaml:"cleanup_period"`
	Eth           EthConfig     `required:"true" yaml:"ethereum"`
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
		Eth: EthConfig{
			PrivateKey: "d07fff36ef2c3d15144974c25d3f5c061ae830a81eefd44292588b3cea2c701c",
		},
	}
}

type EthConfig struct {
	PrivateKey string `required:"true" yaml:"private_key"`
}
