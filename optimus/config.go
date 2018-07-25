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
	"go.uber.org/zap"
)

const (
	PolicySpotOnly OrderPolicy = iota
)

type nodeConfig struct {
	PrivateKey privateKey `yaml:"ethereum" json:"-"`
	Endpoint   auth.Addr  `yaml:"endpoint"`
}

type Config struct {
	Blockchain   *blockchain.Config         `yaml:"blockchain"`
	PrivateKey   privateKey                 `yaml:"ethereum" json:"-"`
	Logging      logging.Config             `yaml:"logging"`
	Node         nodeConfig                 `yaml:"node"`
	Workers      map[auth.Addr]workerConfig `yaml:"workers"`
	Benchmarks   benchmarks.Config          `yaml:"benchmarks"`
	Marketplace  marketplaceConfig          `yaml:"marketplace"`
	Optimization optimizationConfig
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

type workerConfig struct {
	PrivateKey     privateKey         `yaml:"ethereum" json:"-"`
	Epoch          time.Duration      `yaml:"epoch"`
	OrderPolicy    OrderPolicy        `yaml:"order_policy"`
	DryRun         bool               `yaml:"dry_run" default:"false"`
	Identity       sonm.IdentityLevel `yaml:"identity" required:"true"`
	PriceThreshold sonm.Price         `yaml:"price_threshold" required:"true"`
	StaleThreshold time.Duration      `yaml:"stale_threshold" default:"5m"`
	PreludeTimeout time.Duration      `yaml:"prelude_timeout" default:"30s"`
}

func (m *workerConfig) Validate() error {
	if m.PriceThreshold.GetPerSecond().Unwrap().Sign() <= 0 {
		return fmt.Errorf("price threshold must be a positive number")
	}

	return nil
}

type OrderPolicy int

func (m *OrderPolicy) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cfg string
	if err := unmarshal(&cfg); err != nil {
		return err
	}

	switch cfg {
	case "spot_only":
		*m = PolicySpotOnly
	default:
		return fmt.Errorf("unknown order policy: %s", cfg)
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

type optimizationConfig struct {
	Interval          time.Duration
	ClassifierFactory classifierFactory `yaml:"classifier" required:"true" json:"-"`
}

type classifierFactory func(log *zap.Logger) OrderClassifier

func newClassifierFactory(cfgUnmarshal func(interface{}) error) (classifierFactory, error) {
	ty, err := typeofInterface(cfgUnmarshal)
	if err != nil {
		return nil, err
	}

	switch ty {
	case "regression":
		cfg := struct {
			ModelFactory modelFactory `yaml:"model"`
		}{}

		if err := cfgUnmarshal(&cfg); err != nil {
			return nil, err
		}
		if cfg.ModelFactory == nil {
			return nil, fmt.Errorf("missing required field: `optimization/classifier/model`")
		}

		return func(log *zap.Logger) OrderClassifier {
			return newRegressionClassifier(cfg.ModelFactory, log)
		}, nil
	default:
		return nil, fmt.Errorf("unknown classifier: %s", ty)
	}
}

func (m *classifierFactory) UnmarshalYAML(unmarshal func(interface{}) error) error {
	factory, err := newClassifierFactory(unmarshal)
	if err != nil {
		return err
	}

	*m = factory

	return nil
}

func typeofInterface(unmarshal func(interface{}) error) (string, error) {
	raw := struct {
		Type string
	}{}

	if err := unmarshal(&raw); err != nil {
		return "", err
	}

	if raw.Type == "" {
		return "", fmt.Errorf(`"type" field is required`)
	}

	return raw.Type, nil
}
