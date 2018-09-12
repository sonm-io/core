package price

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/params"
)

type CoinMarketCapConfig struct {
	WhatToMineID int           `yaml:"what_to_mine_id" required:"true"`
	URL          string        `yaml:"url" required:"true"`
	Interval     time.Duration `yaml:"update_interval" default:"10m"`
}

type CoinMarketCapFactory struct {
	CoinMarketCapConfig
}

func (m *CoinMarketCapFactory) Config() interface{} {
	return &m.CoinMarketCapConfig
}

func (m *CoinMarketCapFactory) ValidateConfig() error {
	if m.Interval < time.Second {
		return fmt.Errorf("update interval cannot be less that one second")
	}

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

type cmcPriceProvider struct {
	mu     sync.Mutex
	price  *big.Int
	cfg    *CoinMarketCapConfig
	margin float64

	calculate calculateFunc
}

func NewCMCProvider(cfg *CoinMarketCapConfig, margin float64) Provider {
	prov := &cmcPriceProvider{cfg: cfg, margin: margin}

	switch cfg.WhatToMineID {
	case ethWtmID:
		prov.calculate = calculateEthPrice
	case moneroEtmID:
		prov.calculate = calculateXmrPrice
	}

	return prov
}

func (p *cmcPriceProvider) Interval() time.Duration { return p.cfg.Interval }

func (p *cmcPriceProvider) Update(ctx context.Context) error {
	// 1. load price for 1 token in USD
	price, err := getPriceFromCMC(p.cfg.URL)
	if err != nil {
		return err
	}

	// 2. load network parameters
	coinParams, err := getTokenParamsFromWTM(p.cfg.WhatToMineID)
	if err != nil {
		return err
	}

	// 3. calculate token price per hash per second
	v := p.calculate(price, coinParams, p.margin)

	p.mu.Lock()
	defer p.mu.Unlock()

	p.price = v
	return nil
}

func (p *cmcPriceProvider) GetPrice() *big.Int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.price
}

type calculateFunc func(tokenPrice *big.Int, coinParams *coinParams, m float64) *big.Int

func calculateEthPrice(tokenPrice *big.Int, coin *coinParams, margin float64) *big.Int {
	price := big.NewFloat(0).SetInt(tokenPrice)
	reward := big.NewFloat(coin.BlockReward)
	difficulty := big.NewFloat(coin.Difficulty)

	weiPerSecPerHash := big.NewFloat(0).Quo(reward, difficulty)
	ethPerSecPerHash := big.NewFloat(0).Quo(weiPerSecPerHash, big.NewFloat(params.Ether))

	perSecPerHashUSD := big.NewFloat(0).Mul(price, ethPerSecPerHash)
	etherGradedPricePerSecPerHashUSD := big.NewFloat(0).Mul(perSecPerHashUSD, big.NewFloat(params.Ether))

	priceWithMargin := big.NewFloat(0).Mul(etherGradedPricePerSecPerHashUSD, big.NewFloat(margin))
	result, _ := priceWithMargin.Int(nil)
	return result
}

func calculateXmrPrice(tokenPrice *big.Int, coin *coinParams, margin float64) *big.Int {
	price := big.NewFloat(0).SetInt(tokenPrice)

	reward := big.NewFloat(coin.BlockReward * 720)
	netHash := big.NewFloat(float64(coin.Nethash))
	dailyXMR := big.NewFloat(0).Quo(reward, netHash)

	dailyUSD := big.NewFloat(0).Mul(dailyXMR, price)
	secondUSD := big.NewFloat(0).Quo(dailyUSD, big.NewFloat(86400))
	withMargin := big.NewFloat(0).Mul(secondUSD, big.NewFloat(margin))

	v, _ := withMargin.Int(nil)
	return v
}
