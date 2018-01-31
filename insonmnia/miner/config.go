package miner

import (
	"strings"

	"github.com/jinzhu/configor"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
	pb "github.com/sonm-io/core/proto"
)

// HubConfig describes Hub configuration.
type HubConfig struct {
	EthAddr          string           `required:"true" yaml:"eth_addr"`
	ResolveEndpoints bool             `required:"false" yaml:"resolve_endpoints"`
	Endpoints        []string         `required:"false" yaml:"endpoints"`
	CGroups          *ResourcesConfig `required:"false" yaml:"resources"`
}

// FirewallConfig describes firewall detection settings.
type FirewallConfig struct {
	// STUN server endpoint (with port).
	Server string `yaml:"server"`
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
	HubConfig               HubConfig           `required:"true" yaml:"hub"`
	FirewallConfig          *FirewallConfig     `required:"false" yaml:"firewall"`
	Eth                     *accounts.EthConfig `yaml:"ethereum"`
	GPUConfig               string              `required:"false" default:"" yaml:"GPUConfig"`
	SSHConfig               *SSHConfig          `required:"false" yaml:"ssh"`
	LoggingConfig           LoggingConfig       `yaml:"logging"`
	LocatorConfig           *LocatorConfig      `required:"true" yaml:"locator"`
	UUIDPathConfig          string              `required:"false" yaml:"uuid_path"`
	PublicIPsConfig         []string            `required:"false" yaml:"public_ip_addrs"`
	MetricsListenAddrConfig string              `yaml:"metrics_listen_addr" default:"127.0.0.1:14001"`
	PluginsConfig           plugin.Config       `yaml:"plugins"`
}

func (c *config) HubResolveEndpoints() bool {
	return c.HubConfig.ResolveEndpoints
}

func (c *config) HubEthAddr() string {
	return c.HubConfig.EthAddr
}

func (c *config) HubEndpoints() (endpoints []string) {
	return c.HubConfig.Endpoints
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

func (c *config) GPU() pb.GPUVendorType {
	t := strings.ToUpper(c.GPUConfig)
	v, ok := pb.GPUVendorType_value[t]
	if ok {
		return pb.GPUVendorType(v)
	}

	return pb.GPUVendorType_GPU_UNKNOWN
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

func (c *config) MetricsListenAddr() string {
	return c.MetricsListenAddrConfig
}

func (c *config) Plugins() plugin.Config {
	return c.PluginsConfig
}

func (c *config) validate() error {
	if len(c.HubConfig.EthAddr) == 0 {
		return errors.New("hub's ethereum address should be specified")
	}

	if !c.HubConfig.ResolveEndpoints && len(c.HubConfig.Endpoints) == 0 {
		return errors.New("`resolve_endpoints` is `false`, an array of hub's endpoints must be specified")
	}

	if c.HubConfig.ResolveEndpoints && len(c.HubConfig.Endpoints) > 0 {
		return errors.New("`resolve_endpoints` is `true`, only hub's ethereum address should be specified")
	}

	return nil
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

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Config represents a Miner configuration interface.
type Config interface {
	// HubEndpoints returns a string representation of a Hub endpoint to communicate with.
	HubEndpoints() []string
	// HubEthAddr returns hub's ethereum address.
	HubEthAddr() string
	// HubResolveEndpoints returns `true` if we need to resolve hub's endpoints via locator.
	HubResolveEndpoints() bool
	// HubResources returns resources allocated for a Hub.
	HubResources() *ResourcesConfig
	// Firewall returns firewall detection settings.
	Firewall() *FirewallConfig
	// PublicIPs returns all IPs that can be used to communicate with the miner.
	PublicIPs() []string
	// GPU returns GPU Tuner type
	GPU() pb.GPUVendorType
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
	// MetricsListenAddr returns the address that can be used by Prometheus to get
	// metrics.
	MetricsListenAddr() string
	// Plugins returns plugins settings.
	Plugins() plugin.Config
}
