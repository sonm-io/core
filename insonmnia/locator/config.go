package locator

import (
	"time"

	"fmt"
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
)

type StoreConfig struct {
	Type     string `required:"true" default:"boltdb" yaml:"type"`
	Endpoint string `required:"true" default:"/tmp/sonm/boltdb" yaml:"endpoint"`
	Bucket   string `required:"true" default:"sonm" yaml:"bucket"`
	//	KeyFile  string `yaml:"key_file"`
	//	CertFile string `yaml:"cert_file"`
}

type Config struct {
	ListenAddr    string             `yaml:"address"`
	NodeTTL       time.Duration      `yaml:"node_ttl"`
	CleanupPeriod time.Duration      `yaml:"cleanup_period"`
	Eth           accounts.EthConfig `required:"true" yaml:"ethereum"`

	Store StoreConfig `yaml:"store"`
}

// NewConfig loads a config options from the specified YAML file.
func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("cannot load config from file:%v", err)
	}
	return cfg, nil
}

func TestConfig(addr string) *Config {
	return &Config{
		ListenAddr:    addr,
		NodeTTL:       time.Hour,
		CleanupPeriod: time.Minute,

		Store: StoreConfig{
			Type:     "boltdb",
			Bucket:   "sonm",
			Endpoint: "/tmp/sonm/boltdb",
		},
	}
}
