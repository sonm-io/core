package miner

import (
	"github.com/jinzhu/configor"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
	"github.com/sonm-io/core/insonmnia/state"
	"go.uber.org/zap/zapcore"
)

type SSHConfig struct {
	BindEndpoint   string `required:"true" yaml:"bind"`
	PrivateKeyPath string `required:"true" yaml:"private_key_path"`
}

type LoggingConfig struct {
	Level       string `required:"true" default:"debug"`
	parsedLevel zapcore.Level
}

type ResourcesConfig struct {
	Cgroup    string                `required:"true" yaml:"cgroup"`
	Resources *specs.LinuxResources `required:"false" yaml:"resources"`
}

type config struct {
	Resources       *ResourcesConfig    `required:"false" yaml:"resources"`
	SSHConfig       *SSHConfig          `required:"false" yaml:"ssh"`
	LoggingConfig   LoggingConfig       `yaml:"logging"`
	PublicIPsConfig []string            `required:"false" yaml:"public_ip_addrs"`
	PluginsConfig   plugin.Config       `yaml:"plugins"`
	StoreConfig     state.StorageConfig `yaml:"store"`
	BenchConfig     benchmarks.Config   `yaml:"benchmarks"`
}

func (c *config) LogLevel() zapcore.Level {
	return c.LoggingConfig.parsedLevel
}

func (c *config) HubResources() *ResourcesConfig {
	return c.Resources
}

func (c *config) PublicIPs() []string {
	return c.PublicIPsConfig
}

func (c *config) SSH() *SSHConfig {
	return c.SSHConfig
}

func (c *config) Plugins() plugin.Config {
	return c.PluginsConfig
}

func (c *config) Storage() *state.StorageConfig {
	return &c.StoreConfig
}

func (c *config) Benchmarks() benchmarks.Config {
	return c.BenchConfig
}

// NewConfig creates a new Miner config from the specified YAML file.
func NewConfig(path string) (Config, error) {
	cfg := &config{}
	err := configor.Load(cfg, path)

	if err != nil {
		return nil, err
	}

	lvl, err := logging.ParseLogLevel(cfg.LoggingConfig.Level)
	if err != nil {
		return nil, err
	}
	cfg.LoggingConfig.parsedLevel = lvl

	return cfg, nil
}

// Config represents a Miner configuration interface.
type Config interface {
	logging.Leveler

	// HubResources returns resources allocated for a Hub.
	HubResources() *ResourcesConfig
	// PublicIPs returns all IPs that can be used to communicate with the miner.
	PublicIPs() []string
	// SSH returns settings for built-in ssh server
	SSH() *SSHConfig
	// Storage returns storage config
	Storage() *state.StorageConfig
	// Plugins returns plugins settings.
	Plugins() plugin.Config
	// Benchmarks returns benchmarking settings.
	Benchmarks() benchmarks.Config
}
