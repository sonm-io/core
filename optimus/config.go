package optimus

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/logging"
	"go.uber.org/zap"
)

type Config struct {
	PrivateKey   privateKey                 `yaml:"ethereum" json:"-"`
	Logging      logging.Config             `yaml:"logging"`
	Workers      map[auth.Addr]workerConfig `yaml:"workers"`
	Benchmarks   benchmarks.Config          `yaml:"benchmarks"`
	Marketplace  marketplaceConfig          `yaml:"marketplace"`
	Optimization optimizationConfig         `yaml:"optimization"`
}

type workerConfig struct {
	Epoch time.Duration `yaml:"epoch"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
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

	key, err := cfg.LoadKey(accounts.Silent())
	if err != nil {
		return err
	}

	*m = privateKey(*key)
	return nil
}

type marketplaceConfig struct {
	Interval time.Duration
	Endpoint auth.Addr
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
			ModelFactory modelFactory  `yaml:"model"`
			Sigmoid      sigmoidConfig `yaml:"logistic"`
		}{}

		if err := cfgUnmarshal(&cfg); err != nil {
			return nil, err
		}
		if cfg.ModelFactory == nil {
			return nil, fmt.Errorf("missing required field: `optimization/classifier/model`")
		}

		sigmoid := newSigmoid(cfg.Sigmoid)

		return func(log *zap.Logger) OrderClassifier {
			return newRegressionClassifier(cfg.ModelFactory, sigmoid, time.Now, log)
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
