package miner

import (
	"github.com/jinzhu/configor"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
	"github.com/sonm-io/core/insonmnia/npp"
	"github.com/sonm-io/core/insonmnia/state"
)

type SSHConfig struct {
	BindEndpoint   string `required:"true" yaml:"bind"`
	PrivateKeyPath string `required:"true" yaml:"private_key_path"`
}

type ResourcesConfig struct {
	Cgroup    string                `required:"true" yaml:"cgroup"`
	Resources *specs.LinuxResources `required:"false" yaml:"resources"`
}

type WhitelistConfig struct {
	Url                 string   `yaml:"url"`
	Enabled             *bool    `yaml:"enabled" default:"true" required:"true"`
	PrivilegedAddresses []string `yaml:"privileged_addresses"`
	RefreshPeriod       uint     `yaml:"refresh_period" default:"60"`
}

type Config struct {
	Endpoint          string              `yaml:"endpoint" required:"true"`
	Logging           logging.Config      `yaml:"logging"`
	Resources         *ResourcesConfig    `yaml:"resources" required:"false" `
	Eth               accounts.EthConfig  `yaml:"ethereum"`
	NPP               npp.Config          `yaml:"npp"`
	SSH               *SSHConfig          `yaml:"ssh" required:"false" `
	PublicIPs         []string            `yaml:"public_ip_addrs" required:"false" `
	Plugins           plugin.Config       `yaml:"plugins"`
	Storage           state.StorageConfig `yaml:"store"`
	Benchmarks        benchmarks.Config   `yaml:"benchmarks"`
	Whitelist         WhitelistConfig     `yaml:"whitelist"`
	MetricsListenAddr string              `yaml:"metrics_listen_addr" default:"127.0.0.1:14000"`
}

// NewConfig creates a new Miner config from the specified YAML file.
func NewConfig(path string) (*Config, error) {
	cfg := &Config{}

	err := configor.Load(cfg, path)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}
