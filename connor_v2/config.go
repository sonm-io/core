package connor

import (
	"fmt"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/logging"
)

type miningConfig struct {
	Token         string         `yaml:"token" required:"true"`
	Image         string         `yaml:"image" required:"true"`
	Wallet        common.Address `yaml:"wallet" required:"true"`
	PoolReportURL string         `yaml:"pool_report" required:"false"`
	// todo(sshaman1101): parse PoolTrackingURL and replace URL part with wallet address.
	PoolTrackingURL string `yaml:"pool_tracking" required:"false"`
}

type marketConfig struct {
	// todo: (sshaman1101): allow to set multiple subsets for order placing
	FromHashRate uint64 `yaml:"from_hashrate" required:"true"`
	ToHashRate   uint64 `yaml:"to_hashrate" required:"true"`
	Step         uint64 `yaml:"step" required:"true"`
}

type nodeConfig struct {
	Endpoint auth.Addr `json:"endpoint"`
}

type engineConfig struct {
	ConnectionTimeout   time.Duration `yaml:"connection_timeout" default:"30s"`
	OrderWatchInterval  time.Duration `yaml:"order_watch_interval" default:"10s"`
	TaskStartInterval   time.Duration `yaml:"task_start_interval" default:"15s"`
	TaskTrackInterval   time.Duration `yaml:"task_track_interval" default:"15s"`
	TaskRestoreInterval time.Duration `yaml:"task_restore_interval" default:"10s"`
}

type Config struct {
	Node   nodeConfig         `yaml:"node"`
	Eth    accounts.EthConfig `yaml:"ethereum"`
	Market marketConfig       `yaml:"market"`
	Mining miningConfig       `yaml:"mining"`
	Log    logging.Config     `yaml:"log"`
	Engine engineConfig       `yaml:"engine"`

	Metrics string `yaml:"metrics" default:"127.0.0.1:14005"`
}

func (c *Config) validate() error {
	availableTokens := map[string]bool{
		"ETH":  true,
		"ZEC":  true,
		"NULL": true, // null token is for testing purposes
	}

	if _, ok := availableTokens[c.Mining.Token]; !ok {
		return fmt.Errorf("unsupported token \"%s\"", c.Mining.Token)
	}

	named, err := reference.ParseNormalizedNamed(c.Mining.Image)
	if err != nil {
		return fmt.Errorf("cannot parse image name: %v", err)
	}

	c.Mining.Image = named.String()
	return nil
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	if err := configor.Load(cfg, path); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil

}
