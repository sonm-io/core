package hub

import (
	"strings"
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/npp"
	"go.uber.org/zap/zapcore"
)

type LoggingConfig struct {
	Level       string `required:"true" default:"debug"`
	parsedLevel zapcore.Level
}

type MarketConfig struct {
	Endpoint     string        `required:"true" yaml:"endpoint"`
	UpdatePeriod time.Duration `default:"60s" yaml:"update_period_sec"`
}

type StoreConfig struct {
	Endpoint string `required:"true" default:"/tmp/sonm/boltdb" yaml:"endpoint"`
	Bucket   string `required:"true" default:"sonm" yaml:"bucket"`
}

type WhitelistConfig struct {
	Url                 string   `yaml:"url"`
	Enabled             *bool    `yaml:"enabled" default:"true" required:"true"`
	PrivilegedAddresses []string `yaml:"privileged_addresses"`
	RefreshPeriod       uint     `yaml:"refresh_period" default:"60"`
}

type Config struct {
	Endpoint          string `required:"true" yaml:"endpoint"`
	AnnounceEndpoint  string
	Logging           LoggingConfig      `yaml:"logging"`
	Eth               accounts.EthConfig `yaml:"ethereum"`
	Market            MarketConfig       `yaml:"market"`
	Store             StoreConfig        `yaml:"store"`
	Whitelist         WhitelistConfig    `yaml:"whitelist"`
	MetricsListenAddr string             `yaml:"metrics_listen_addr" default:"127.0.0.1:14000"`
	NPP               npp.Config
}

func (c *Config) LogLevel() zapcore.Level {
	return c.Logging.parsedLevel
}

// NewConfig loads a hub config from the specified YAML file.
func NewConfig(path string) (*Config, error) {
	conf := &Config{}
	err := configor.Load(conf, path)
	if err != nil {
		return nil, err
	}

	lvl, err := logging.ParseLogLevel(conf.Logging.Level)
	if err != nil {
		return nil, err
	}
	conf.Logging.parsedLevel = lvl

	return conf, nil
}

func (c *Config) EndpointIP() string {
	return strings.Split(c.Endpoint, ":")[0]
}
