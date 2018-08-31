package price

import (
	"context"
	"math/big"
)

type staticProvider struct {
	value *big.Int
}

func NewStaticProvider(cfg *StaticProviderConfig) Provider {
	return &staticProvider{value: big.NewInt(cfg.Price)}
}

func (p *staticProvider) Update(ctx context.Context) error {
	return nil
}

func (p *staticProvider) GetPrice() *big.Int {
	return p.value
}
