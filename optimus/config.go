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
)

type Config struct {
	PrivateKey   privateKey `yaml:"ethereum" json:"-"`
	Logging      logging.Config
	Workers      map[auth.Addr]WorkerConfig `yaml:"workers"`
	Benchmarks   benchmarks.Config          `yaml:"benchmarks"`
	Marketplace  marketplaceConfig
	Optimization optimizationConfig
}

type WorkerConfig struct {
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
	Interval   time.Duration
	Classifier newClassifier `json:"-"`
}

type newClassifier func() OrderClassifier

func (m *newClassifier) UnmarshalYAML(unmarshal func(interface{}) error) error {
	ty, err := typeofInterface(unmarshal)
	if err != nil {
		return err
	}

	switch ty {
	case "regression":
		cfg := struct {
			Model   newModel
			Sigmoid sigmoidConfig `yaml:"logistic"`
		}{}

		if err := unmarshal(&cfg); err != nil {
			return err
		}

		sigmoid := newSigmoid(cfg.Sigmoid)

		*m = func() OrderClassifier {
			return newRegressionClassifier(cfg.Model, sigmoid, time.Now)
		}
	default:
		return fmt.Errorf("unknown classifier: %s", ty)
	}

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
