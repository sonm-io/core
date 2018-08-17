package price

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/params"
)

const (
	retryCount   = 3
	retryTimeout = 1 * time.Second
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

type ProviderConfig struct {
	Margin float64
	URL    string
}

type nullPriceProvider struct {
	cfg *ProviderConfig
}

func NewNullPriceProvider(cfg *ProviderConfig) Provider {
	return &nullPriceProvider{cfg: cfg}
}

func (p *nullPriceProvider) Update(ctx context.Context) error {
	return nil
}

func (p *nullPriceProvider) GetPrice() *big.Int {
	v, _ := big.NewFloat(0).Mul(big.NewFloat(5e5), big.NewFloat(p.cfg.Margin)).Int(nil)
	return v
}

type ethPriceProvider struct {
	mu    sync.Mutex
	price *big.Int
	cfg   *ProviderConfig
}

func NewEthPriceProvider(cfg *ProviderConfig) Provider {
	return &ethPriceProvider{cfg: cfg}
}

func (p *ethPriceProvider) Update(ctx context.Context) error {
	// 1. load price for 1 token in USD
	price, err := getPriceFromCMC(p.cfg.URL)
	if err != nil {
		return err
	}

	// 2. load network parameters
	coinParams, err := getTokenParamsFromWTM(ethWtmID)
	if err != nil {
		return err
	}

	// 3. calculate token price per hash per second
	v := p.calculate(
		big.NewFloat(0).SetInt(price),
		big.NewFloat(coinParams.BlockReward),
		big.NewFloat(coinParams.Difficulty),
	)

	p.mu.Lock()
	defer p.mu.Unlock()

	p.price = v
	return nil
}

func (p *ethPriceProvider) calculate(price, reward, difficulty *big.Float) *big.Int {
	weiPerSecPerHash := big.NewFloat(0).Quo(reward, difficulty)
	ethPerSecPerHash := big.NewFloat(0).Quo(weiPerSecPerHash, big.NewFloat(params.Ether))

	perSecPerHashUSD := big.NewFloat(0).Mul(price, ethPerSecPerHash)
	etherGradedPricePerSecPerHashUSD := big.NewFloat(0).Mul(perSecPerHashUSD, big.NewFloat(params.Ether))

	priceWithMargin := big.NewFloat(0).Mul(etherGradedPricePerSecPerHashUSD, big.NewFloat(p.cfg.Margin))
	result, _ := priceWithMargin.Int(nil)
	return result
}

func (p *ethPriceProvider) GetPrice() *big.Int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.price
}

type xmrPriceProvider struct {
	mu    sync.Mutex
	price *big.Int
	cfg   *ProviderConfig
}

func NewXmrPriceProvider(cfg *ProviderConfig) Provider {
	return &xmrPriceProvider{cfg: cfg}
}

func (p *xmrPriceProvider) Update(ctx context.Context) error {
	// 1. load price for 1 token in USD
	price, err := getPriceFromCMC(p.cfg.URL)
	if err != nil {
		return err
	}

	// 2. load network parameters
	coinParams, err := getTokenParamsFromWTM(moneroEtmID)
	if err != nil {
		return err
	}

	// 3. calculate token price per hash per second
	v := p.calculate(
		big.NewFloat(0).SetInt(price),
		big.NewFloat(coinParams.BlockReward),
		big.NewFloat(coinParams.Difficulty),
	)

	p.mu.Lock()
	defer p.mu.Unlock()

	p.price = v
	return nil
}

func (p *xmrPriceProvider) calculate(price, reward, difficulty *big.Float) *big.Int {
	hashrate := big.NewFloat(0).Quo(big.NewFloat(1), difficulty)

	xmrPerHashPerSec := big.NewFloat(0).Mul(hashrate, reward)
	usdPerHashPerSec := big.NewFloat(0).Mul(xmrPerHashPerSec, price)

	priceWithMargin := big.NewFloat(0).Mul(usdPerHashPerSec, big.NewFloat(p.cfg.Margin))
	result, _ := priceWithMargin.Int(nil)
	return result
}

func (p *xmrPriceProvider) GetPrice() *big.Int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.price
}
