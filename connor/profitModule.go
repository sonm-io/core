package connor

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/params"
	"github.com/sonm-io/core/connor/database"
	"github.com/sonm-io/core/connor/watchers"
	"go.uber.org/zap"
)

const (
	hashingPower     = 1
	costPerkWh       = 0.0
	powerConsumption = 0.0
)

type ProfitableModule struct {
	c *Connor
}

func NewProfitableModules(c *Connor) *ProfitableModule {
	return &ProfitableModule{
		c: c,
	}
}

type powerAndDivider struct {
	power float64
	div   float64
}

func (p *ProfitableModule) getHashPowerAndDividerForToken(s string, hp float64) (float64, float64, bool) {
	var tokenHashPower = map[string]powerAndDivider{
		"ETH": {div: 1, power: hashingPower * 1000000.0},
		"XMR": {div: 1, power: 1},
		"ZEC": {div: 1, power: 1},
	}
	k, ok := tokenHashPower[s]
	if !ok {
		return .0, .0, false
	}
	return k.power, k.div, true
}

type TokenMainData struct {
	Symbol            string
	ProfitPerDaySnm   float64
	ProfitPerMonthSnm float64
	ProfitPerMonthUsd float64
}

func (p *ProfitableModule) getTokensForProfitCalculation() []*TokenMainData {
	// todo: make configurable
	return []*TokenMainData{
		{Symbol: "ETH"},
		{Symbol: "XMR"},
		{Symbol: "ZEC"},
	}
}

func (p *ProfitableModule) CollectTokensMiningProfit(t watchers.TokenWatcher) ([]*TokenMainData, error) {
	p.c.logger.Debug("CollectTokensMiningProfit")

	var tokensForCalc = p.getTokensForProfitCalculation()
	for _, token := range tokensForCalc {
		tokenData, err := t.GetTokenData(token.Symbol)
		if err != nil {
			return nil, fmt.Errorf("cannot get token data: %v", err)
		}

		hashesPerSecond, divider, ok := p.getHashPowerAndDividerForToken(tokenData.Symbol, tokenData.NetHashPerSec)
		if !ok {
			p.c.logger.Info("cannot process tokenData", zap.String("token", tokenData.Symbol))
			continue
		}

		netHashesPerSec := int64(tokenData.NetHashPerSec)
		token.ProfitPerMonthUsd, err = p.CalculateMiningProfit(tokenData.PriceUSD, hashesPerSecond, float64(netHashesPerSec), tokenData.BlockReward, divider, tokenData.BlockTime)
		id, err := strconv.Atoi(tokenData.CmcID)
		if err != nil {
			return nil, err
		}
		if token.Symbol == p.c.cfg.Mining.Token {
			p.c.db.SaveProfitToken(&database.TokenDb{
				ID:              int64(id),
				Name:            token.Symbol,
				UsdPrice:        tokenData.PriceUSD,
				NetHashesPerSec: tokenData.NetHashPerSec,
				BlockTime:       tokenData.BlockTime,
				BlockReward:     tokenData.BlockReward,
				ProfitPerMonth:  token.ProfitPerMonthUsd,
				DateTime:        time.Now(),
			})
		}
	}
	return tokensForCalc, nil
}

// TODO(sshaman1101): tests
func (p *ProfitableModule) CalculateMiningProfit(usd, hashesPerSecond, netHashesPerSecond, blockReward, div float64, blockTime int64) (float64, error) {
	if div == 0 {
		return 0, fmt.Errorf("the current div is 0")
	}
	currentHashingPower := hashesPerSecond / div

	miningShare := currentHashingPower / (netHashesPerSecond + currentHashingPower)
	minedPerDay := miningShare * 86400 / float64(blockTime) * blockReward / div
	powerCostPerDayUSD := (powerConsumption * 24) / 1000 * costPerkWh
	returnPerDayUSD := (usd*minedPerDay - (usd * minedPerDay * 0.01)) - powerCostPerDayUSD
	perMonthUSD := float64(returnPerDayUSD * 30)
	return perMonthUSD, nil
}

// Limit balance for Charge orders. Default value = 0.5
// TODO(sshaman1101): tests
func (p *ProfitableModule) LimitChargeSNM(balance *big.Int, partCharge float64) *big.Int {
	limitChargeSNM := balance.Div(balance, big.NewInt(100))
	limitChargeSNM = limitChargeSNM.Mul(balance, big.NewInt(int64(partCharge*100)))
	return limitChargeSNM
}

// ConvertSNMBalanceToUSD converts amount of tokens (Ether-graded value)
// to human-friendly price in USD.
//
// balance * price / ether = balance in USD, e.g:
// balance = 40e18 SNM
// price = 0.22 $/SNM
// balance in USD = 8.8$
func (p *ProfitableModule) ConvertSNMBalanceToUSD(balance *big.Int, usdForOneSNM float64) float64 {
	bal := big.NewFloat(0).SetInt(balance)
	price := big.NewFloat(usdForOneSNM)

	ether := big.NewFloat(params.Ether)
	priceInUSDInPowOfEther := big.NewFloat(0).Mul(bal, price)

	// divide result by ether to get total amount of funds in "normal",
	// not the 1e18-graded power.
	result, _ := big.NewFloat(0).Quo(priceInUSDInPowOfEther, ether).Float64()

	return result
}
