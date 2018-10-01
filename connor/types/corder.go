package types

import (
	"fmt"
	"math/big"
	"time"

	"github.com/cnf/structhash"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
)

const (
	orderCancelDelayStep = 5 * time.Second
	orderCancelMaxDelay  = 5 * time.Minute
)

type CorderFactory interface {
	FromOrder(order *sonm.Order) *Corder
	FromParams(price *big.Int, hashrate uint64, bench Benchmarks) *Corder
	FromSlice(orders []*sonm.Order) []*Corder
}

func NewCorderFactory(tag string, benchmarkIndex int, counterparty common.Address) CorderFactory {
	return &anyCorderFactory{
		orderTag:       tag,
		benchmarkIndex: benchmarkIndex,
		counterparty:   counterparty,
	}
}

type anyCorderFactory struct {
	benchmarkIndex int
	orderTag       string
	counterparty   common.Address
}

func (a *anyCorderFactory) FromOrder(order *sonm.Order) *Corder {
	return &Corder{Order: order, benchmarkIndex: a.benchmarkIndex}
}

func (a *anyCorderFactory) FromParams(price *big.Int, hashrate uint64, bench Benchmarks) *Corder {
	bench.Values[a.benchmarkIndex] = hashrate

	ord := &sonm.Order{
		OrderType:      sonm.OrderType_BID,
		Price:          sonm.NewBigInt(price),
		Netflags:       &sonm.NetFlags{Flags: sonm.NetworkOutbound},
		IdentityLevel:  sonm.IdentityLevel_ANONYMOUS,
		Tag:            []byte(a.orderTag),
		Benchmarks:     bench.unwrap(),
		CounterpartyID: sonm.NewEthAddress(a.counterparty),
	}

	return &Corder{Order: ord, benchmarkIndex: a.benchmarkIndex}
}

func (a *anyCorderFactory) FromSlice(orders []*sonm.Order) []*Corder {
	v := make([]*Corder, len(orders))
	for i, ord := range orders {
		v[i] = a.FromOrder(ord)
	}

	return v
}

type Corder struct {
	*sonm.Order
	benchmarkIndex int
}

func (co *Corder) GetHashrate() uint64 {
	return co.GetBenchmarks().Get(co.benchmarkIndex)
}

func (co *Corder) AsBID() *sonm.BidOrder {
	return &sonm.BidOrder{
		Price:        &sonm.Price{PerSecond: co.Order.GetPrice()},
		Blacklist:    sonm.NewEthAddress(common.HexToAddress(co.Order.GetBlacklist())),
		Identity:     co.Order.IdentityLevel,
		Tag:          string(co.Tag),
		Counterparty: co.CounterpartyID,
		Resources: &sonm.BidResources{
			Network: &sonm.BidNetwork{
				Overlay:  co.Order.GetNetflags().GetOverlay(),
				Outbound: co.Order.GetNetflags().GetOutbound(),
				Incoming: co.Order.GetNetflags().GetIncoming(),
			},
			Benchmarks: benchmarksToMap(co.Order.Benchmarks),
		},
	}
}

func (co *Corder) RestorePrice() *big.Int {
	hashrate := big.NewInt(0).SetUint64(co.GetHashrate())
	return big.NewInt(0).Div(co.GetPrice().Unwrap(), hashrate)
}

func isOrderReplaceable(currentPrice, newPrice *big.Float, delta float64) bool {
	diff := big.NewFloat(0).Mul(currentPrice, big.NewFloat(delta))

	upperBound := big.NewFloat(0).Add(currentPrice, diff)
	lowerBound := big.NewFloat(0).Sub(currentPrice, diff)

	upperHit := newPrice.Cmp(upperBound) >= 0
	lowerHit := newPrice.Cmp(lowerBound) <= 0

	return upperHit || lowerHit
}

func (co *Corder) IsReplaceable(newPrice *big.Int, delta float64) bool {
	currentPrice := big.NewFloat(0).SetInt(co.RestorePrice())
	newFloatPrice := big.NewFloat(0).SetInt(newPrice)

	return isOrderReplaceable(currentPrice, newFloatPrice, delta)
}

func (co *Corder) Hash() string {
	s := struct {
		Benchmarks   []uint64
		Counterparty common.Address
		Netflags     uint64
	}{
		Benchmarks:   co.Benchmarks.Values,
		Counterparty: co.GetCounterpartyID().Unwrap(),
		Netflags:     co.GetNetflags().GetFlags(),
	}

	return fmt.Sprintf("%x", structhash.Sha1(s, 1))
}

type CorderCancelTuple struct {
	Corder *Corder
	Delay  time.Duration
}

func (c *CorderCancelTuple) WithIncreasedDelay() *CorderCancelTuple {
	c.Delay *= 2
	if c.Delay > orderCancelMaxDelay {
		c.Delay = orderCancelMaxDelay
	}
	return c
}

func NewCorderCancelTuple(c *Corder) *CorderCancelTuple {
	return &CorderCancelTuple{Corder: c, Delay: orderCancelDelayStep}
}
