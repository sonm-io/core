package price

import (
	"context"
	"fmt"
	"math/big"
	"sync"
)

type sonmPriceProvider struct {
	mu    sync.Mutex
	price *big.Int
}

// NewSonmPriceProvider returns price provider that keeps price per one SNM token,
// the value is Ether-graded.
func NewSonmPriceProvider() Provider { return &sonmPriceProvider{} }

func (p *sonmPriceProvider) Update(ctx context.Context) error {
	url := fmt.Sprintf("%s/%s", priceBaseURL, sonmURLPart)
	price, err := getPriceFromCMC(url)
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
