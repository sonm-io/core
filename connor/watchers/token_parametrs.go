package watchers

/*
	Watcher for Token's parameters: Ethereum, Monero, Zcash.
	Parameters: CmcID, Symbol, BlockTime, BlockReward, NetHashPerSec, PriceUSD
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/pkg/errors"
)

type tokenPriceWatcher struct {
	mu     sync.Mutex
	urlCMC string
	urlCRC string
	data   map[string]*TokenParameters
}

// tokenSnapshot represents several coin parameters related to mining
type tokenSnapshot struct {
	Data struct {
		General struct {
			ID                 string  `json:"ID"`
			Symbol             string  `json:"Symbol"`
			BlockTime          int64   `json:"BlockTime"`
			NetHashesPerSecond float64 `json:"NetHashesPerSecond"`
			BlockReward        float64 `json:"BlockReward"`
		} `json:"General"`
	} `json:"Data"`
}

// tokenData represents tokenData price in different currencies
type tokenData struct {
	ID       string `json:"id"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	PriceUsd string `json:"price_usd"`
}

func getTokensForUpdate() []*tokenData {
	return []*tokenData{
		//Symbol FROM coinmarketcup.com. ID FROM cryptocompare.com
		{Name: "Ethereum", Symbol: "ETH", ID: "7605"},
		{Name: "Monero", Symbol: "XMR", ID: "5038"},
		{Name: "ZCash", Symbol: "ZEC", ID: "24854"},
	}
}

func NewTokenPriceWatcher(urlCmc, urlCrypto string) TokenWatcher {
	return &tokenPriceWatcher{
		urlCMC: urlCmc,
		urlCRC: urlCrypto,
		data:   make(map[string]*TokenParameters),
	}
}

func (p *tokenPriceWatcher) GetTokenData(symbol string) (*TokenParameters, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	d, ok := p.data[symbol]
	if !ok {
		return nil, errors.New("no token with given symbol")
	}

	return d, nil
}

func (p *tokenPriceWatcher) Update(ctx context.Context) error {
	var (
		tokenPrices    []*tokenData   //priceUSD, Symbol
		tokensSnapshot *tokenSnapshot //Id, blockTime, blockReward, hashesPerSec
	)

	body, err := fetchBody(p.urlCMC)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &tokenPrices); err != nil {
		return err
	}

	tokensToWorkWith := getTokensForUpdate()
	tokensToWorkWithMap := make(map[string]tokenData)

	for _, t := range tokensToWorkWith {
		tokensToWorkWithMap[t.Symbol] = tokenData{
			ID:       t.ID,
			Symbol:   t.Symbol,
			PriceUsd: t.PriceUsd,
			Name:     t.Name,
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, t := range tokenPrices {
		if _, ok := tokensToWorkWithMap[t.Symbol]; ok {
			priceUSD, err := strconv.ParseFloat(t.PriceUsd, 64)
			if err != nil {
				return errors.Wrapf(err, "cannot parse USD price from \"%s\"", t.PriceUsd)
			}
			p.data[t.Symbol] = &TokenParameters{
				Symbol:   t.Symbol,
				PriceUSD: priceUSD,
			}
		}
	}

	/* SNAPSHOT */
	for _, token := range tokensToWorkWith {
		body, err := fetchBody(p.urlCRC + token.ID)
		if err != nil {
			return nil
		}
		if err := json.Unmarshal(body, &tokensSnapshot); err != nil {
			return fmt.Errorf("cannot unmarshal data for token %s: %v", token.ID, err)
		}

		params, ok := p.data[tokensSnapshot.Data.General.Symbol]
		if !ok {
			return fmt.Errorf("cannot get token by symbol \"%s\"", tokensSnapshot.Data.General.Symbol)
		}

		params.CmcID = tokensSnapshot.Data.General.ID
		params.BlockTime = tokensSnapshot.Data.General.BlockTime
		params.BlockReward = tokensSnapshot.Data.General.BlockReward
		params.NetHashPerSec = tokensSnapshot.Data.General.NetHashesPerSecond
	}

	return nil
}

type TokenParameters struct {
	CmcID         string
	Symbol        string
	BlockTime     int64
	BlockReward   float64
	NetHashPerSec float64
	PriceUSD      float64
}
