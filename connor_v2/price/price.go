package price

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/params"
)

const (
	retryCount   = 3
	retryTimeout = 1 * time.Second
	// append ticker name and also *trailing slash*
	priceBaseURL = "https://api.coinmarketcap.com/v1/ticker"
	sonmURLPart  = "sonm/"
	zcashURLPart = "zcash/"
	ethURLPart   = "ethereum/"
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

func NewProvider(token string) Provider {
	switch token {
	case "NULL":
		return newNullPriceProvider()
	case "ETH":
		return newEthPriceProvider()
	case "ZEC":
		return newZecPriceProvider()
	default:
		// should never happens
		panic("cannot get price updater for token " + token)
	}
}

type nullPriceProvider struct{}

func (p *nullPriceProvider) Update(ctx context.Context) error {
	return nil
}

func (p *nullPriceProvider) GetPrice() *big.Int {
	return big.NewInt(5e10)
}

func newNullPriceProvider() Provider { return &nullPriceProvider{} }

type ethPriceProvider struct {
	mu    sync.Mutex
	price *big.Int
}

func newEthPriceProvider() Provider { return &ethPriceProvider{} }

func (p *ethPriceProvider) Update(ctx context.Context) error {
	// 1. load price for 1 token in USD
	url := fmt.Sprintf("%s/%s", priceBaseURL, ethURLPart)
	price, err := getPriceFromCMC(url)
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
	result, _ := etherGradedPricePerSecPerHashUSD.Int(nil)
	return result
}

func (p *ethPriceProvider) GetPrice() *big.Int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.price
}

type zecPriceProvider struct {
	mu    sync.Mutex
	price *big.Int
}

func newZecPriceProvider() Provider { return &zecPriceProvider{price: big.NewInt(1)} }

func (p *zecPriceProvider) Update(ctx context.Context) error {
	return nil
}

func (p *zecPriceProvider) calculate(price, reward, difficulty *big.Float) *big.Int {
	// YourHashrate / NetHashRate / BlockTime * 86400 * BlockReward
	// todo: check formula, calculated results is different than whatTiMine's one
	return nil
}

func (p *zecPriceProvider) GetPrice() *big.Int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.price
}
