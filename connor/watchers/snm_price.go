package watchers

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
)

type snmPriceWatcher struct {
	mu   sync.Mutex
	url  string
	data float64
}

func (p *snmPriceWatcher) Update(ctx context.Context) error {
	data, err := loadSNMPriceUSD(p.url)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.data = data
	return nil
}

func (p *snmPriceWatcher) GetPrice() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.data
}

func loadSNMPriceUSD(url string) (float64, error) {
	body, err := fetchBody(url)
	if err != nil {
		return 0, err
	}
	var tickerSnm []*tokenData
	if err := json.Unmarshal(body, &tickerSnm); err != nil {
		return 0, err
	}

	return strconv.ParseFloat(tickerSnm[0].PriceUsd, 64)
}

func NewSNMPriceWatcher(url string) PriceWatcher {
	return &snmPriceWatcher{
		mu:   sync.Mutex{},
		url:  url,
		data: 0,
	}
}
