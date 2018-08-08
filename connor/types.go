package connor

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/connor/price"
	"github.com/sonm-io/core/proto"
)

type CorderFactoriy interface {
	FromOrder(order *sonm.Order) *Corder
	FromParams(price *big.Int, hashrate uint64, bench Benchmarks) *Corder
	FromSlice(orders []*sonm.Order) []*Corder
}

func NewCorderFactory(token string, benchmarkIndex int) CorderFactoriy {
	return &anyCorderFactory{
		tokenName:      token,
		benchmarkIndex: benchmarkIndex,
	}
}

type anyCorderFactory struct {
	benchmarkIndex int
	tokenName      string
}

func (a *anyCorderFactory) FromOrder(order *sonm.Order) *Corder {
	return &Corder{Order: order, benchmarkIndex: a.benchmarkIndex}
}

func (a *anyCorderFactory) FromParams(price *big.Int, hashrate uint64, bench Benchmarks) *Corder {
	bench.Values[a.benchmarkIndex] = hashrate

	ord := &sonm.Order{
		OrderType:     sonm.OrderType_BID,
		Price:         sonm.NewBigInt(price),
		Netflags:      &sonm.NetFlags{Flags: sonm.NetworkOutbound},
		IdentityLevel: sonm.IdentityLevel_ANONYMOUS,
		Tag:           []byte(fmt.Sprintf("connor_%s", strings.ToLower(a.tokenName))),
		Benchmarks:    bench.unwrap(),
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
		Price:     &sonm.Price{PerSecond: co.Order.GetPrice()},
		Blacklist: sonm.NewEthAddress(common.StringToAddress(co.Order.GetBlacklist())),
		Identity:  co.Order.IdentityLevel,
		Tag:       string(co.Tag),
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

func (co *Corder) restorePrice() *big.Int {
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

func (co *Corder) isReplaceable(newPrice *big.Int, delta float64) bool {
	currentPrice := big.NewFloat(0).SetInt(co.restorePrice())
	newFloatPrice := big.NewFloat(0).SetInt(newPrice)

	return isOrderReplaceable(currentPrice, newFloatPrice, delta)
}

type Benchmarks sonm.Benchmarks

func newZeroBenchmarks() Benchmarks {
	return Benchmarks{Values: make([]uint64, sonm.MinNumBenchmarks)}
}

func (b *Benchmarks) setGPUEth(v uint64) {
	b.Values[9] = v
}

func (b *Benchmarks) setGPUZcash(v uint64) {
	b.Values[10] = v
}

func (b *Benchmarks) setGPURedshift(v uint64) {
	b.Values[11] = v
}

func (b *Benchmarks) unwrap() *sonm.Benchmarks {
	v := sonm.Benchmarks(*b)
	return &v
}

func (b *Benchmarks) toMap() map[string]uint64 {
	// warn: this is shitty crutch, but we should refactor
	// CreateOrder method to omit this.

	v := b.unwrap()
	return map[string]uint64{
		"cpu-sysbench-multi":  v.CPUSysbenchMulti(),
		"cpu-sysbench-single": v.CPUSysbenchOne(),
		"cpu-cores":           v.CPUCores(),
		"ram-size":            v.RAMSize(),
		"storage-size":        v.StorageSize(),
		"net-download":        v.NetTrafficIn(),
		"net-upload":          v.NetTrafficOut(),
		"gpu-count":           v.GPUCount(),
		"gpu-mem":             v.GPUMem(),
		"gpu-eth-hashrate":    v.GPUEthHashrate(),
		"gpu-cash-hashrate":   v.GPUCashHashrate(),
		"gpu-redshift":        v.GPURedshift(),
	}
}

func benchmarksToMap(b *sonm.Benchmarks) map[string]uint64 {
	v := Benchmarks(*b)
	return v.toMap()
}

type taskStatus struct {
	*sonm.TaskStatusReply
	id string
}

type tokenParameters struct {
	corderFactory CorderFactoriy
	priceProvider price.Provider
	poolWatcher   interface{}
	// todo: how to deal with log parser?
	// the log parser is bound to container?
}
