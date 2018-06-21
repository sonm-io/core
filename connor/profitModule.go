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

// CALCULATE TOKENS
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
	return []*TokenMainData{
		{Symbol: "ETH"},
		{Symbol: "XMR"},
		{Symbol: "ZEC"},
	}
}
func (p *ProfitableModule) CollectTokensMiningProfit(t watchers.TokenWatcher) ([]*TokenMainData, error) {
	var tokensForCalc = p.getTokensForProfitCalculation()
	for _, token := range tokensForCalc {
		tokenData, err := t.GetTokenData(token.Symbol)
		if err != nil {
			return nil, fmt.Errorf("cannot get token data %v\r\n", err)
		}
		hashesPerSecond, divider, ok := p.getHashPowerAndDividerForToken(tokenData.Symbol, tokenData.NetHashPerSec)
		if !ok {
			p.c.logger.Info("cannot process tokenData", zap.String("token", tokenData.Symbol))
			continue
		}
		netHashesPerSec := int64(tokenData.NetHashPerSec)
		token.ProfitPerMonthUsd = p.CalculateMiningProfit(tokenData.PriceUSD, hashesPerSecond, float64(netHashesPerSec), tokenData.BlockReward, divider, tokenData.BlockTime)
		id, err := strconv.Atoi(tokenData.CmcID)
		if err != nil {
			return nil, err
		}
		if token.Symbol == p.c.cfg.UsingToken.Token {
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
func (p *ProfitableModule) CalculateMiningProfit(usd, hashesPerSecond, netHashesPerSecond, blockReward, div float64, blockTime int64) float64 {
	currentHashingPower := hashesPerSecond / div
	miningShare := currentHashingPower / (netHashesPerSecond + currentHashingPower)
	minedPerDay := miningShare * 86400 / float64(blockTime) * blockReward / div
	powerCostPerDayUSD := (powerConsumption * 24) / 1000 * costPerkWh
	returnPerDayUSD := (usd*minedPerDay - (usd * minedPerDay * 0.01)) - powerCostPerDayUSD
	perMonthUSD := float64(returnPerDayUSD * 30)
	return perMonthUSD
}

//Limit balance for Charge orders. Default value = 0.5
func (p *ProfitableModule) LimitChargeSNM(balance *big.Int, partCharge float64) *big.Int {
	limitChargeSNM := balance.Div(balance, big.NewInt(100))
	limitChargeSNM = limitChargeSNM.Mul(balance, big.NewInt(int64(partCharge*100)))
	return limitChargeSNM
}

//converting snmBalance = > USD Balance
func (p *ProfitableModule) ConvertingToUSDBalance(balanceSide *big.Int, snmPrice float64) float64 {
	bal := balanceSide.Mul(balanceSide, big.NewInt(int64(snmPrice*params.Ether)))
	bal = bal.Div(bal, big.NewInt(params.Ether))
	d, e := bal.DivMod(bal, big.NewInt(params.Ether), big.NewInt(0))
	f, _ := big.NewFloat(0).SetInt(e).Float64()
	res := float64(d.Int64()) + (f / params.Ether)
	return res
}
