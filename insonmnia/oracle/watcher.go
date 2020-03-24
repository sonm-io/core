package oracle

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance"
	"github.com/ethereum/go-ethereum/params"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"math/big"
	"time"
)

type PriceData struct {
	Price *big.Int
	Err   error
}

type PriceWatcher struct {
	parsePeriod time.Duration
	data        chan *PriceData
}

func NewPriceWatcher(parsePeriod time.Duration) *PriceWatcher {
	return &PriceWatcher{
		data:        make(chan *PriceData),
		parsePeriod: parsePeriod,
	}
}

func (p *PriceWatcher) Start(ctx context.Context) <-chan *PriceData {
	go func() {
		t := util.NewImmediateTicker(p.parsePeriod)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				price, err := p.loadCurrentPrice(ctx)
				if err != nil {
					p.data <- &PriceData{Price: nil, Err: err}
				}
				p.data <- &PriceData{Price: price, Err: nil}
			}
		}
	}()
	return p.data
}

func (p *PriceWatcher) loadCurrentPrice(ctx context.Context) (*big.Int, error) {
	logger := ctxlog.GetLogger(ctx)
	usdPrice, err := p.loadSNMPrice()
	if err != nil {
		return nil, err
	}
	logger.Debug("Download new price", zap.String("price", usdPrice.String()))
	return p.divideSNM(usdPrice), nil
}

func (p *PriceWatcher) divideSNM(price *big.Float) *big.Int {
	snmInOneUsd := big.NewFloat(0).Quo(big.NewFloat(1), price)
	snmInOneUsd.Mul(snmInOneUsd, big.NewFloat(params.Ether))
	intPrice, _ := snmInOneUsd.Int(nil)
	return intPrice
}

func (p *PriceWatcher) loadSNMPrice() (*big.Float, error) {
	client := binance.NewClient("", "")
	// futuresClient := binance.NewFuturesClient(apiKey, secretKey)
	resSNM, err := client.NewDepthService().Symbol("SNMBTC").Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("failed to load SNM price: %v", err)
	}
	resBTC, err := client.NewDepthService().Symbol("BTCUSDT").Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("failed to load BTC price: %v", err)
	}

	snmPrice, ok := big.NewFloat(0).SetString(resSNM.Asks[0].Price)
	if !ok {
		return nil, fmt.Errorf("failed to parse SNM price: %v", err)
	}
	btcPrice, ok := big.NewFloat(0).SetString(resBTC.Asks[0].Price)
	if !ok {
		return nil, fmt.Errorf("failed to parse BTC price: %v", err)
	}
	return btcPrice.Mul(btcPrice, snmPrice), nil
}
