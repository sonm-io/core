package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"time"

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
	logger.Debug("Download new price", zap.Float64("price", usdPrice))
	return p.divideSNM(usdPrice), nil
}

func (p *PriceWatcher) divideSNM(price float64) *big.Int {
	return big.NewInt(int64(1 / price * 1e18))
}

func (p *PriceWatcher) loadSNMPrice(url string) (float64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var tickerSnm []*tokenData
	err = json.Unmarshal(body, &tickerSnm)
	if err != nil {
		return 0, err
	}
	if len(tickerSnm) < 1 {
		return 0, fmt.Errorf("loading ticker is abused")
	}
	return strconv.ParseFloat(tickerSnm[0].PriceUsd, 64)
}
