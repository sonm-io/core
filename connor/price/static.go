package price

import (
	"fmt"
	"math/big"
)

type StaticProviderConfig struct {
	Price int64 `yaml:"price" required:"true"`
}

type StaticFactory struct {
	StaticProviderConfig
}

func (m *StaticFactory) Config() interface{} {
	return &m.StaticProviderConfig
}

func (m *StaticFactory) ValidateConfig() error {
	if m.StaticProviderConfig.Price < 0 {
		return fmt.Errorf("price value should be positive")
	}

	return nil
}

func (m *StaticFactory) Init(Margin float64) Provider {
	return NewStaticProvider(&m.StaticProviderConfig)
}

type staticProvider struct {
	value *big.Int
}

func NewStaticProvider(cfg *StaticProviderConfig) Provider {
	return &staticProvider{value: big.NewInt(cfg.Price)}
}

func (p *staticProvider) GetPrice() *big.Int {
	return p.value
}
