package miner

import (
	"github.com/jinzhu/configor"
)

// HubConfig describes Hub configuration.
type HubConfig struct {
	Endpoint  string     `required:"false" yaml:"endpoint"`
	Resources *Resources `required:"false" yaml:"resources"`
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

type config struct {
	HubConfig      *HubConfig      `required:"false" yaml:"hub"`
	FirewallConfig *FirewallConfig `required:"false" yaml:"firewall"`
	GPUConfig      *GPUConfig      `required:"false"`
	SSHConfig      *SSHConfig      `required:"false" yaml:"ssh"`
	LoggingConfig  LoggingConfig   `yaml:"logging"`
}

func (c *config) HubEndpoint() string {
	return c.HubConfig.Endpoint
}

func (c *config) HubResources() *Resources {
	return c.HubConfig.Resources
}

func (c *config) Firewall() *FirewallConfig {
	return c.FirewallConfig
}

func (c *config) GPU() *GPUConfig {
	return c.GPUConfig
}

func (c *config) SSH() *SSHConfig {
	return c.SSHConfig
}

func (c *config) Logging() LoggingConfig {
	return c.LoggingConfig
}

// NewConfig creates a new Miner config from the specified YAML file.
func NewConfig(path string) (Config, error) {
	cfg := &config{}
	err := configor.Load(cfg, path)
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
	HubResources() *Resources
	// Firewall returns firewall detection settings.
	Firewall() *FirewallConfig
	// GPU returns options about NVIDIA GPU support via nvidia-docker-plugin
	GPU() *GPUConfig
	// SSH returns settings for built-in ssh server
	SSH() *SSHConfig
	// Logging returns logging settings.
	Logging() LoggingConfig
}
