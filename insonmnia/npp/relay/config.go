package relay

import (
	"crypto/ecdsa"
	"fmt"
	"os"

	"github.com/jinzhu/configor"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/util/debug"
	"github.com/sonm-io/core/util/netutil"
)

// ClusterConfig represents a cluster membership config.
type ClusterConfig struct {
	Name      string
	Endpoint  string
	Announce  string
	SecretKey string `yaml:"secret_key" json:"-"`
	Members   []string
}

type MonitorConfig struct {
	Endpoint   string
	PrivateKey *ecdsa.PrivateKey `json:"-"`
}

type monitorConfig struct {
	Endpoint string
	ETH      accounts.EthConfig `yaml:"ethereum"`
}

type serverConfig struct {
	Addr    netutil.TCPAddr `yaml:"endpoint" required:"true"`
	Cluster ClusterConfig   `yaml:"cluster"`
	Logging logging.Config  `yaml:"logging"`
	Monitor monitorConfig   `yaml:"monitoring"`
	Debug   *debug.Config   `yaml:"debug"`
}

// ServerConfig describes the complete relay server configuration.
type ServerConfig struct {
	Addr    netutil.TCPAddr
	Cluster ClusterConfig
	Logging logging.Config
	Monitor MonitorConfig
	Debug   *debug.Config
}

// NewServerConfig loads a new Relay server config from a file.
func NewServerConfig(path string) (*ServerConfig, error) {
	cfg := &serverConfig{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	if len(cfg.Cluster.Name) == 0 {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		cfg.Cluster.Name = fmt.Sprintf("%s-%s", hostname, uuid.New())
	}

	privateKey, err := cfg.Monitor.ETH.LoadKey()
	if err != nil {
		return nil, err
	}

	return &ServerConfig{
		Addr:    cfg.Addr,
		Cluster: cfg.Cluster,
		Logging: cfg.Logging,
		Monitor: MonitorConfig{
			Endpoint:   cfg.Monitor.Endpoint,
			PrivateKey: privateKey,
		},
		Debug: cfg.Debug,
	}, nil
}

// Config represents a client-side relay configuration.
//
// Used as a basic building block for high-level configurations.
type Config struct {
	Endpoints   []string
	Concurrency uint8 `yaml:"concurrency" default:"2"`
}
