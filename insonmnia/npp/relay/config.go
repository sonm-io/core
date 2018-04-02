package relay

import (
	"crypto/ecdsa"
	"fmt"
	"os"

	"github.com/jinzhu/configor"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/util/netutil"
	"go.uber.org/zap/zapcore"
)

// ClusterConfig represents a cluster membership config.
type ClusterConfig struct {
	Name      string
	Endpoint  string
	Announce  string
	SecretKey string `yaml:"secret_key" json:"-"`
	Members   []string
}

// LoggingConfig represents a logging config.
type LoggingConfig struct {
	Level string `required:"true" default:"debug"`
	level zapcore.Level
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
	Logging LoggingConfig   `yaml:"logging"`
	Monitor monitorConfig   `yaml:"monitoring"`
}

// ServerConfig describes the complete relay server configuration.
type ServerConfig struct {
	Addr    netutil.TCPAddr
	Cluster ClusterConfig
	Logging LoggingConfig
	Monitor MonitorConfig
}

// NewServerConfig loads a new Relay server config from a file.
func NewServerConfig(path string) (*ServerConfig, error) {
	cfg := &serverConfig{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	lvl, err := logging.ParseLogLevel(cfg.Logging.Level)
	if err != nil {
		return nil, err
	}
	cfg.Logging.level = lvl

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
	}, nil
}

// LogLevel returns the minimum logging level configured.
func (c *ServerConfig) LogLevel() zapcore.Level {
	return c.Logging.level
}

// Config represents a client-side relay configuration.
//
// Used as a basic building block for high-level configurations.
type Config struct {
	Endpoints []netutil.TCPAddr
}
