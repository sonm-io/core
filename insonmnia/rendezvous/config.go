package rendezvous

import (
	"crypto/ecdsa"
	"net"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"go.uber.org/zap/zapcore"
)

// LoggingConfig represents a logging config.
type LoggingConfig struct {
	Level string `required:"true" default:"debug"`
	level zapcore.Level
}

// Config represents a Rendezvous server configuration.
type Config struct {
	// Listening address.
	Addr       net.Addr
	PrivateKey *ecdsa.PrivateKey
	Logging    LoggingConfig
}

// LogLevel returns the minimum logging level configured.
func (c *Config) LogLevel() zapcore.Level {
	return c.Logging.level
}

type config struct {
	Addr    string             `yaml:"endpoint" required:"true"`
	Eth     accounts.EthConfig `yaml:"ethereum"`
	Logging LoggingConfig      `yaml:"logging"`
}

// NewConfig loads a new Rendezvous server config from a file.
func NewConfig(path string) (*Config, error) {
	cfg := &config{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveTCPAddr("tcp", cfg.Addr)
	if err != nil {
		return nil, err
	}

	privateKey, err := cfg.Eth.LoadKey()
	if err != nil {
		return nil, err
	}

	lvl, err := logging.ParseLogLevel(cfg.Logging.Level)
	if err != nil {
		return nil, err
	}
	cfg.Logging.level = lvl

	return &Config{
		Addr:       addr,
		PrivateKey: privateKey,
		Logging:    cfg.Logging,
	}, nil
}
