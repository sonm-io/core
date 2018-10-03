package worker

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/configor"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/dwh"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/matcher"
	"github.com/sonm-io/core/insonmnia/npp"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/insonmnia/worker/plugin"
	"github.com/sonm-io/core/insonmnia/worker/salesman"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/debug"
)

type ResourcesConfig struct {
	Cgroup    string                `required:"true" yaml:"cgroup"`
	Resources *specs.LinuxResources `required:"false" yaml:"resources"`
}

type WhitelistConfig struct {
	Url string `yaml:"url"`
	// Deprecated: use PrivilegedIdentityLevel instead. Breaking issue #1470.
	Enabled                 *bool              `yaml:"enabled" default:"true" required:"true"`
	PrivilegedAddresses     []string           `yaml:"privileged_addresses"`
	RefreshPeriod           uint               `yaml:"refresh_period" default:"60"`
	PrivilegedIdentityLevel sonm.IdentityLevel `yaml:"privileged_identity_level" default:"identified"`
}

type DevConfig struct {
	DisableMasterApproval bool `yaml:"disable_master_approval"`
}

type Config struct {
	Endpoint          string              `yaml:"endpoint" required:"true"`
	Logging           logging.Config      `yaml:"logging"`
	Resources         *ResourcesConfig    `yaml:"resources" required:"false"`
	Blockchain        *blockchain.Config  `yaml:"blockchain"`
	NPP               npp.Config          `yaml:"npp"`
	SSH               *SSHConfig          `yaml:"ssh" required:"false" `
	PublicIPs         []string            `yaml:"public_ip_addrs" required:"false" `
	Plugins           plugin.Config       `yaml:"plugins"`
	Storage           state.StorageConfig `yaml:"store"`
	Benchmarks        benchmarks.Config   `yaml:"benchmarks"`
	Whitelist         WhitelistConfig     `yaml:"whitelist"`
	MetricsListenAddr string              `yaml:"metrics_listen_addr" default:"127.0.0.1:14000"`
	DWH               dwh.YAMLConfig      `yaml:"dwh"`
	Matcher           *matcher.YAMLConfig `yaml:"matcher"`
	Salesman          salesman.YAMLConfig `yaml:"salesman"`
	Master            common.Address      `yaml:"master" required:"true"`
	Development       *DevConfig          `yaml:"development"`
	Admin             *common.Address     `yaml:"admin"`
	Debug             *debug.Config       `yaml:"debug"`
	Superusers        SuperusersConfig    `yaml: "superusers"`
}

// NewConfig creates a new Worker config from the specified YAML file.
func NewConfig(path string) (*Config, error) {
	cfg := &Config{}

	err := configor.Load(cfg, path)

	if err != nil {
		return nil, err
	}
	// TODO: drop in next major version. Breaking issue #1470.
	if !*cfg.Whitelist.Enabled {
		cfg.Whitelist.PrivilegedIdentityLevel = sonm.IdentityLevel_ANONYMOUS
	}

	return cfg, nil
}

type SuperusersConfig struct {
	UpdatePeriod time.Duration `yaml:"update_period" default:"1m"`
}
