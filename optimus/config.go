package optimus

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/debug"
)

type nodeConfig struct {
	PrivateKey privateKey `yaml:"ethereum" json:"-"`
	Endpoint   auth.Addr  `yaml:"endpoint"`
}

type OptimizationConfig struct {
	Model optimizationMethodFactory `yaml:"model"`
}

type Config struct {
	Restrictions *RestrictionsConfig         `yaml:"restrictions"`
	Blockchain   *blockchain.Config          `yaml:"blockchain"`
	PrivateKey   privateKey                  `yaml:"ethereum" json:"-"`
	Logging      logging.Config              `yaml:"logging"`
	Node         nodeConfig                  `yaml:"node"`
	Workers      map[auth.Addr]*workerConfig `yaml:"workers"`
	Benchmarks   benchmarks.Config           `yaml:"benchmarks"`
	Marketplace  marketplaceConfig           `yaml:"marketplace"`
	Debug        *debug.Config               `yaml:"debug"`
}

func (m *Config) Validate() error {
	for _, cfg := range m.Workers {
		if err := cfg.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	if err := configor.Load(cfg, path); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

type simulationConfig struct {
	Orders []*sonm.BigInt `yaml:"orders"`
}

type RestrictionsConfig struct {
	Name     string  `yaml:"cgroup_name" default:"/optimus"`
	CPUCount float64 `yaml:"cpu_count" default:"1.0"`
	// Upper RSS threshold in megabytes.
	//
	// Zero value, which is default, means no limit.
	MemoryLimit uint64 `yaml:"memory_limit"`
}

type workerConfig struct {
	PrivateKey     privateKey         `yaml:"ethereum" json:"-"`
	Epoch          time.Duration      `yaml:"epoch"`
	OrderDuration  time.Duration      `yaml:"order_duration_threshold" default:"24h"`
	DryRun         bool               `yaml:"dry_run" default:"false"`
	Identity       sonm.IdentityLevel `yaml:"identity" required:"true"`
	PriceThreshold priceThreshold     `yaml:"price_threshold" required:"true"`
	StaleThreshold time.Duration      `yaml:"stale_threshold" default:"5m"`
	PreludeTimeout time.Duration      `yaml:"prelude_timeout" default:"30s"`
	Optimization   OptimizationConfig `yaml:"optimization"`
	Simulation     *simulationConfig  `yaml:"simulation"`
	PlanPolicy     *planPolicy        `yaml:"plan_policy" default:"entire_machine"`
	VerboseLog     bool               `yaml:"verbose"`
}

func (m *workerConfig) Validate() error {
	if m.OrderDuration < 0 {
		return fmt.Errorf("order duration threshold must be non-negative")
	}

	if m.Optimization.Model.OptimizationMethodFactory == nil {
		m.Optimization.Model = optimizationMethodFactory{OptimizationMethodFactory: &defaultOptimizationMethodFactory{}}
	}

	if m.PlanPolicy == nil {
		m.PlanPolicy = &planPolicy{Type: planPolicyEntireMachine}
	}

	return nil
}

type privateKey ecdsa.PrivateKey

func (m *privateKey) Unwrap() *ecdsa.PrivateKey {
	key := ecdsa.PrivateKey(*m)
	return &key
}

func (m *privateKey) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cfg accounts.EthConfig
	if err := unmarshal(&cfg); err != nil {
		return err
	}

	key, err := cfg.LoadKey()
	if err != nil {
		return err
	}

	*m = privateKey(*key)
	return nil
}

type marketplaceConfig struct {
	PrivateKey privateKey    `yaml:"ethereum" json:"-"`
	Endpoint   auth.Addr     `yaml:"endpoint"`
	Interval   time.Duration `yaml:"interval"`
	MinPrice   *sonm.Price   `yaml:"min_price" default:"0.0001 USD/h"`
}

func typeofInterface(unmarshal func(interface{}) error) (string, error) {
	raw := struct {
		Type string `yaml:"type"`
	}{}

	if err := unmarshal(&raw); err != nil {
		return "", err
	}

	if raw.Type == "" {
		return "", fmt.Errorf(`"type" field is required`)
	}

	return raw.Type, nil
}
