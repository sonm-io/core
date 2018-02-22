package hub

import (
	"strings"
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"go.uber.org/zap/zapcore"
)

type LoggingConfig struct {
	Level string `required:"true" default:"debug"`
}

type GatewayConfig struct {
	Ports []uint16 `required:"true" yaml:"ports"`
}

type LocatorConfig struct {
	Endpoint     string        `yaml:"endpoint" required:"true"`
	UpdatePeriod time.Duration `yaml:"update_period" required:"true" default:"10s"`
}

type MarketConfig struct {
	Endpoint     string        `required:"true" yaml:"endpoint"`
	UpdatePeriod time.Duration `default:"60s" yaml:"update_period_sec"`
}

type StoreConfig struct {
	Type     string `required:"true" default:"boltdb" yaml:"type"`
	Endpoint string `required:"true" default:"/tmp/sonm/boltdb" yaml:"endpoint"`
	Bucket   string `required:"true" default:"sonm" yaml:"bucket"`
	KeyFile  string `yaml:"key_file"`
	CertFile string `yaml:"cert_file"`
}

type ClusterConfig struct {
	Store                        StoreConfig   `yaml:"store"`
	Failover                     bool          `yaml:"failover"`
	Endpoint                     string        `yaml:"endpoint"`
	AnnounceEndpoint             string        `yaml:"announce_endpoint"`
	LeaderKey                    string        `yaml:"leader_key" default:"sonm/hub/leader"`
	MemberListKey                string        `yaml:"member_list_key" default:"sonm/hub/list"`
	SynchronizableEntitiesPrefix string        `yaml:"sync_prefix" default:"sonm/hub/sync"`
	LeaderTTL                    time.Duration `yaml:"leader_ttl" default:"20s"`
	AnnouncePeriod               time.Duration `yaml:"announce_period" default:"10s"`
	AnnounceTTL                  time.Duration `yaml:"announce_ttl" default:"20s"`
	MemberGCPeriod               time.Duration `yaml:"member_gc_period" default:"15s"`
}

type WhitelistConfig struct {
	Url                 string   `yaml:"url"`
	Enabled             *bool    `yaml:"enabled" default:"true" required:"true"`
	PrivilegedAddresses []string `yaml:"privileged_addresses"`
	RefreshPeriod       uint     `yaml:"refresh_period" default:"60"`
}

type Config struct {
	Endpoint          string             `required:"true" yaml:"endpoint"`
	GatewayConfig     *GatewayConfig     `yaml:"gateway"`
	Logging           LoggingConfig      `yaml:"logging"`
	Eth               accounts.EthConfig `yaml:"ethereum"`
	Locator           LocatorConfig      `yaml:"locator"`
	Market            MarketConfig       `yaml:"market"`
	Cluster           ClusterConfig      `yaml:"cluster"`
	Whitelist         WhitelistConfig    `yaml:"whitelist"`
	MetricsListenAddr string             `yaml:"metrics_listen_addr" default:"127.0.0.1:14000"`
}

func (c *Config) LogLevel() zapcore.Level {
	return logging.ParseLogLevel(c.Logging.Level)
}

// NewConfig loads a hub config from the specified YAML file.
func NewConfig(path string) (*Config, error) {
	conf := &Config{}
	err := configor.Load(conf, path)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func (c *Config) EndpointIP() string {
	return strings.Split(c.Endpoint, ":")[0]
}
