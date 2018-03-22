package miner

import (
	"github.com/jinzhu/configor"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
	"go.uber.org/zap/zapcore"
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
	Level       string `required:"true" default:"debug"`
	parsedLevel zapcore.Level
}

type ResourcesConfig struct {
	Cgroup    string                `required:"true" yaml:"cgroup"`
	Resources *specs.LinuxResources `required:"false" yaml:"resources"`
}

type LocatorConfig struct {
	Endpoint string `required:"true" yaml:"endpoint"`
}

type DevConfig struct {
	DevAddr  string `yaml:"listen"`
	Insecure bool   `yaml:"insecure"`
}

type storeConfig struct {
	Path string `required:"true" yaml:"path" default:"/var/lib/sonm/worker.boltdb"`
}

type config struct {
	HubConfig               HubConfig           `required:"true" yaml:"hub"`
	FirewallConfig          *FirewallConfig     `required:"false" yaml:"firewall"`
	Eth                     *accounts.EthConfig `yaml:"ethereum"`
	SSHConfig               *SSHConfig          `required:"false" yaml:"ssh"`
	LoggingConfig           LoggingConfig       `yaml:"logging"`
	LocatorConfig           *LocatorConfig      `required:"true" yaml:"locator"`
	PublicIPsConfig         []string            `required:"false" yaml:"public_ip_addrs"`
	MetricsListenAddrConfig string              `yaml:"metrics_listen_addr" default:"127.0.0.1:14001"`
	PluginsConfig           plugin.Config       `yaml:"plugins"`
	StoreConfig             storeConfig         `yaml:"store"`
	DevConfig               *DevConfig          `yaml:"yes_i_want_to_use_dev-only_features"`
}

func (c *config) LogLevel() zapcore.Level {
	return c.LoggingConfig.parsedLevel
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

func (c *config) SSH() *SSHConfig {
	return c.SSHConfig
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

func (c *config) Dev() *DevConfig {
	return c.DevConfig
}

func (c *config) Store() string {
	return c.StoreConfig.Path
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

	if err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
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
	// SSH returns settings for built-in ssh server
	SSH() *SSHConfig
	// Store returns path to boltdb which keeps Worker's state.
	Store() string
	// ETH returns ethereum configuration
	ETH() *accounts.EthConfig
	// LocatorEndpoint returns locator endpoint.
	LocatorEndpoint() string
	// MetricsListenAddr returns the address that can be used by Prometheus to get
	// metrics.
	MetricsListenAddr() string
	// Plugins returns plugins settings.
	Plugins() plugin.Config
	// DevAddr to listen on. For dev purposes only!
	Dev() *DevConfig
}
