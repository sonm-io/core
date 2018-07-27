package price

import (
	"context"
	"math/big"
	"sync"
)

const (
	sonmPriceURL = "https://api.coinmarketcap.com/v1/ticker/sonm/"
)

type sonmPriceProvider struct {
	mu    sync.Mutex
	price *big.Int
}

// NewSonmPriceProvider returns price provider that keeps price per one SNM token,
// the value is Ether-graded.
func NewSonmPriceProvider() Provider { return &sonmPriceProvider{} }

func (p *sonmPriceProvider) Update(ctx context.Context) error {
	price, err := getPriceFromCMC(sonmPriceURL)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.price = price
	return nil
}

func (p *sonmPriceProvider) GetPrice() *big.Int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.price
}
