package miner

import (
	"github.com/jinzhu/configor"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/miner/gpu"
)

// HubConfig describes Hub configuration.
type HubConfig struct {
	Endpoint string           `required:"true" yaml:"endpoint"`
	CGroups  *ResourcesConfig `required:"false" yaml:"resources"`
}

// FirewallConfig describes firewall detection settings.
type FirewallConfig struct {
	// STUN server endpoint (with port).
	Server string `yaml:"server"`
}

// GPUConfig contains options related to NVIDIA GPU support
type GPUConfig struct {
	NvidiaDockerDriver string `yaml:"nvidiadockerdriver"`
}

type SSHConfig struct {
	BindEndpoint   string `required:"true" yaml:"bind"`
	PrivateKeyPath string `required:"true" yaml:"private_key_path"`
}

type LoggingConfig struct {
	Level int `required:"true" default:"1"`
}

type ResourcesConfig struct {
	Cgroup    string                `required:"true" yaml:"cgroup"`
	Resources *specs.LinuxResources `required:"false" yaml:"resources"`
}

type LocatorConfig struct {
	Endpoint string `required:"true" yaml:"endpoint"`
}

type config struct {
	HubConfig       HubConfig           `required:"true" yaml:"hub"`
	FirewallConfig  *FirewallConfig     `required:"false" yaml:"firewall"`
	Eth             *accounts.EthConfig `yaml:"ethereum"`
	GPUConfig       *gpu.Config         `required:"false" yaml:"GPUConfig"`
	SSHConfig       *SSHConfig          `required:"false" yaml:"ssh"`
	LoggingConfig   LoggingConfig       `yaml:"logging"`
	LocatorConfig   *LocatorConfig      `required:"true" yaml:"locator"`
	UUIDPathConfig  string              `required:"false" yaml:"uuid_path"`
	PublicIPsConfig []string            `required:"false" yaml:"public_ip_addrs"`
}

func (c *config) HubEndpoint() string {
	return c.HubConfig.Endpoint
}

func (c *config) HubResources() *ResourcesConfig {
	return c.HubConfig.CGroups
}

func (c *config) Firewall() *FirewallConfig {
	return c.FirewallConfig
}

func (c *config) PublicIPs() []string {
	return c.PublicIPsConfig
}

func (c *config) GPU() *gpu.Config {
	return c.GPUConfig
}

func (c *config) SSH() *SSHConfig {
	return c.SSHConfig
}

func (c *config) Logging() LoggingConfig {
	return c.LoggingConfig
}

func (c *config) UUIDPath() string {
	return c.UUIDPathConfig
}

func (c *config) ETH() *accounts.EthConfig {
	return c.Eth
}

func (c *config) LocatorEndpoint() string {
	return c.LocatorConfig.Endpoint
}

// NewConfig creates a new Miner config from the specified YAML file.
func NewConfig(path string) (Config, error) {
	cfg := &config{}
	err := configor.Load(cfg, path)
	if cfg.UUIDPath() == "" {
		cfg.UUIDPathConfig = path + ".uuid"
	}
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Config represents a Miner configuration interface.
type Config interface {
	// HubEndpoint returns a string representation of a Hub endpoint to communicate with.
	HubEndpoint() string
	// HubResources returns resources allocated for a Hub.
	HubResources() *ResourcesConfig
	// Firewall returns firewall detection settings.
	Firewall() *FirewallConfig
	// PublicIPs returns all IPs that can be used to communicate with the miner.
	PublicIPs() []string
	// GPU returns options about NVIDIA GPU support via nvidia-docker-plugin
	GPU() *gpu.Config
	// SSH returns settings for built-in ssh server
	SSH() *SSHConfig
	// Logging returns logging settings.
	Logging() LoggingConfig
	// Path to store Miner uuid
	UUIDPath() string
	// ETH returns ethereum configuration
	ETH() *accounts.EthConfig
	// LocatorEndpoint returns locator endpoint.
	LocatorEndpoint() string
}
