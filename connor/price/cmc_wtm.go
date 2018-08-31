package price

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/params"
)

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
	v := p.calculate(
		big.NewFloat(0).SetInt(price),
		big.NewFloat(coinParams.BlockReward),
		big.NewFloat(coinParams.Difficulty),
		p.margin,
	)

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

type calculateFunc func(a, b, c *big.Float, m float64) *big.Int

func calculateEthPrice(price, reward, difficulty *big.Float, margin float64) *big.Int {
	weiPerSecPerHash := big.NewFloat(0).Quo(reward, difficulty)
	ethPerSecPerHash := big.NewFloat(0).Quo(weiPerSecPerHash, big.NewFloat(params.Ether))

	perSecPerHashUSD := big.NewFloat(0).Mul(price, ethPerSecPerHash)
	etherGradedPricePerSecPerHashUSD := big.NewFloat(0).Mul(perSecPerHashUSD, big.NewFloat(params.Ether))

	priceWithMargin := big.NewFloat(0).Mul(etherGradedPricePerSecPerHashUSD, big.NewFloat(margin))
	result, _ := priceWithMargin.Int(nil)
	return result
}

func calculateXmrPrice(price, reward, difficulty *big.Float, margin float64) *big.Int {
	hashrate := big.NewFloat(0).Quo(big.NewFloat(1), difficulty)

	xmrPerHashPerSec := big.NewFloat(0).Mul(hashrate, reward)
	usdPerHashPerSec := big.NewFloat(0).Mul(xmrPerHashPerSec, price)

	priceWithMargin := big.NewFloat(0).Mul(usdPerHashPerSec, big.NewFloat(margin))
	result, _ := priceWithMargin.Int(nil)
	return result
}
