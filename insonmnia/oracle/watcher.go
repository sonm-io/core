package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/params"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

const (
	snmPriceTickerURL string = "https://api.coinmarketcap.com/v1/ticker/sonm/"
)

type tokenData struct {
	ID       string `json:"id"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	PriceUsd string `json:"price_usd"`
}

type PriceData struct {
	price *big.Int
	err   error
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
					p.data <- &PriceData{price: nil, err: err}
				}
				p.data <- &PriceData{price: price, err: nil}
			}
		}
	}()
	return p.data
}

func (p *PriceWatcher) loadCurrentPrice(ctx context.Context) (*big.Int, error) {
	logger := ctxlog.GetLogger(ctx)
	usdPrice, err := p.loadSNMPrice(snmPriceTickerURL)
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

func (p *PriceWatcher) loadSNMPrice(url string) (*big.Float, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var tickerSnm []*tokenData
	err = json.Unmarshal(body, &tickerSnm)
	if err != nil {
		return nil, err
	}
	if len(tickerSnm) < 1 {
		return nil, fmt.Errorf("loading ticker is abused")
	}
	f, _, err := new(big.Float).Parse(tickerSnm[0].PriceUsd, 10)
	if err != nil {
		return nil, err
	}
	return f, nil
}
