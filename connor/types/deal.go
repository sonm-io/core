package types

import (
	"math/big"

	"github.com/sonm-io/core/proto"
)

type DealFactory interface {
	FromDeal(deal *sonm.Deal) *Deal
}

func NewDealFactory(benchmarkIndex int) DealFactory {
	return &anyDealFactory{benchmarkIndex: benchmarkIndex}
}

type anyDealFactory struct {
	benchmarkIndex int
}

func (a *anyDealFactory) FromDeal(deal *sonm.Deal) *Deal {
	return &Deal{
		Deal:           deal,
		benchmarkIndex: a.benchmarkIndex,
	}
}

type Deal struct {
	*sonm.Deal
	benchmarkIndex int
}

func (d *Deal) BenchmarkValue() uint64 {
	return d.GetBenchmarks().Get(d.benchmarkIndex)
}

func (d *Deal) Unwrap() *sonm.Deal {
	return d.Deal
}

func (d *Deal) RestorePrice() *big.Int {
	hashrate := big.NewInt(0).SetUint64(d.BenchmarkValue())
	return big.NewInt(0).Div(d.GetPrice().Unwrap(), hashrate)
}

func isDealReplaceable(currentPrice, newPrice *big.Float, delta float64) bool {
	diff := big.NewFloat(0).Mul(currentPrice, big.NewFloat(delta))
	lowerBound := big.NewFloat(0).Sub(currentPrice, diff)
	// deal should be replaced only if we hit lower bound of price
	// threshold (mining profit is less that we paying for the deal).
	return newPrice.Cmp(lowerBound) <= 0
}

func (d *Deal) IsReplaceable(actualPrice *big.Int, delta float64) bool {
	current := big.NewFloat(0).SetInt(d.RestorePrice())
	actual := big.NewFloat(0).SetInt(actualPrice)

	return isDealReplaceable(current, actual, delta)
}
