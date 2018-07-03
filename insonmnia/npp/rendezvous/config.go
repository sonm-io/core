package rendezvous

import (
	"crypto/ecdsa"
	"net"
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/util/debug"
	"github.com/sonm-io/core/util/netutil"
)

// ServerConfig represents a Rendezvous server configuration.
type ServerConfig struct {
	// Listening address.
	Addr       net.Addr
	PrivateKey *ecdsa.PrivateKey
	Logging    logging.Config
	Debug      *debug.Config
}

type serverConfig struct {
	Addr    netutil.TCPAddr    `yaml:"endpoint" required:"true"`
	Eth     accounts.EthConfig `yaml:"ethereum"`
	Logging logging.Config     `yaml:"logging"`
	Debug   *debug.Config      `yaml:"debug"`
}

// NewServerConfig loads a new Rendezvous server config from a file.
func NewServerConfig(path string) (*ServerConfig, error) {
	cfg := &serverConfig{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	privateKey, err := cfg.Eth.LoadKey()
	if err != nil {
		return nil, err
	}

	return &ServerConfig{
		Addr:       &cfg.Addr,
		PrivateKey: privateKey,
		Logging:    cfg.Logging,
		Debug:      cfg.Debug,
	}, nil
}

type Config struct {
	Endpoints             []auth.Addr   `yaml:"endpoints"`
	MaxConnectionAttempts int           `yaml:"max_connection_attempts" default:"5"`
	Timeout               time.Duration `yaml:"timeout" default:"3s"`
}
