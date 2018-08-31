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
	// Update loads actual price from external source
	Update(ctx context.Context) error
	// GetPrice returns last known value of token's price.
	// Note that value is ether-graded (1e18).
	GetPrice() *big.Int
}

type Factory interface {
	Config() interface{}
	Init(margin float64) Provider
	validateConfig() error
}

// Updateable provider should refresh their state
// with given interval.
type Updateable interface {
	UpdateInterval() time.Duration
}

type CoinMarketCapConfig struct {
	WhatToMineID int           `yaml:"what_to_mine_id" required:"true"`
	URL          string        `yaml:"url" required:"true"`
	Interval     time.Duration `yaml:"update_interval" default:"10m"`
}

func (c *CoinMarketCapConfig) UpdateInterval() time.Duration { return c.Interval }

type CoinMarketCapFactory struct {
	CoinMarketCapConfig
}

func (m *CoinMarketCapFactory) Config() interface{} {
	return &m.CoinMarketCapConfig
}

func (m *CoinMarketCapFactory) validateConfig() error {
	supported := []int{ethWtmID, moneroEtmID}
	for _, id := range supported {
		if id == m.CoinMarketCapConfig.WhatToMineID {
			return nil
		}
	}

	return fmt.Errorf("unsupported whattomine id: %d", m.CoinMarketCapConfig.WhatToMineID)
}

func (m *CoinMarketCapFactory) Init(margin float64) Provider {
	return NewCMCProvider(&m.CoinMarketCapConfig, margin)
}

type StaticProviderConfig struct {
	Price int64 `yaml:"price" required:"true"`
}

type StaticFactory struct {
	StaticProviderConfig
}

func (m *StaticFactory) Config() interface{} {
	return &m.StaticProviderConfig
}

func (m *StaticFactory) validateConfig() error {
	if m.StaticProviderConfig.Price < 0 {
		return fmt.Errorf("price value should be positive")
	}

	return nil
}

func (m *StaticFactory) Init(Margin float64) Provider {
	return NewStaticProvider(&m.StaticProviderConfig)
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

	m.Factory = instance
	return nil
}
