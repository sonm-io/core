package price

import (
	"context"
	"fmt"
	"math/big"
	"time"
)

// Provider loads and calculates mining profit for specified token.
// The value is is USD per Second per hash.
type Provider interface {
	// GetPrice returns last known value of token's price.
	// Note that value is ether-graded (1e18).
	GetPrice() *big.Int
}

// Updateable provider should refresh their state
// with given interval.
type Updateable interface {
	Update(ctx context.Context) error
	Interval() time.Duration
}

type Factory interface {
	Config() interface{}
	ValidateConfig() error
	Init(margin float64) Provider
}

func NewFactory(t string) Factory {
	switch t {
	case "cmc":
		return &CoinMarketCapFactory{}
	case "static":
		return &StaticFactory{}
	// note: we should also support node as price provider
	default:
		return nil
	}
}

// todo: copy-paste form optimus/config.go:142
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

// SourceConfig configure different sources which van be used
// to obtain price for orders/deals.
type SourceConfig struct {
	Factory
}

func (m *SourceConfig) MarshalYAML() (interface{}, error) {
	return m.Config(), nil
}

func (m *SourceConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	factoryType, err := typeofInterface(unmarshal)
	if err != nil {
		return err
	}

	instance := NewFactory(factoryType)
	if instance == nil {
		return fmt.Errorf("unknown price provider type: %v", factoryType)
	}

	cfg := instance.Config()
	if err := unmarshal(cfg); err != nil {
		return err
	}

	if err := instance.ValidateConfig(); err != nil {
		return err
	}

	m.Factory = instance
	return nil
}
