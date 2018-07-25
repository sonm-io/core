package connor

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
)

type Corder struct {
	*sonm.Order
	token string
}

func NewCorderFromOrder(order *sonm.Order, token string) *Corder {
	return &Corder{Order: order, token: token}
}

func NewCorderFromParams(token string, price *big.Int, hashrate uint64) (*Corder, error) {
	bench := newBenchmark()
	switch token {
	case "ETH":
		bench.setGPUMemory(3000e6)
		bench.setGPUEth(hashrate)
	case "ZEC":
		// todo: I should find the right value for this
		bench.setGPUMemory(900e6)
		bench.setGPUZcash(hashrate)
	case "NULL":
		bench.setGPUMemory(1e6)
		bench.setGPURedshift(hashrate)
	}

	ord := &sonm.Order{
		OrderType:     sonm.OrderType_BID,
		Price:         sonm.NewBigInt(price),
		Netflags:      &sonm.NetFlags{Flags: sonm.NetworkOutbound},
		IdentityLevel: sonm.IdentityLevel_ANONYMOUS,
		// Blacklist:     "",
		Tag:        []byte(fmt.Sprintf("connor_v2_%s", strings.ToLower(token))),
		Benchmarks: bench.unwrap(),
	}

	return &Corder{Order: ord, token: token}, nil
}

func (co *Corder) GetHashrate() uint64 {
	switch co.token {
	case "ETH":
		return co.GetBenchmarks().GPUEthHashrate()
	case "ZEC":
		return co.GetBenchmarks().GPUCashHashrate()
	case "NULL":
		return co.GetBenchmarks().GPURedshift()
	default:
		return 0
	}
}

func (co *Corder) AsBID() *sonm.BidOrder { // todo: tests
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
			Benchmarks: benchmarkMap(co.Order.Benchmarks),
		},
	}
}

func NewCordersSlice(orders []*sonm.Order, token string) []*Corder {
	v := make([]*Corder, len(orders))
	for i, ord := range orders {
		v[i] = NewCorderFromOrder(ord, token)
	}

	return v
}

type Benchmarks sonm.Benchmarks

func (b *Benchmarks) setGPUMemory(v uint64) {
	b.Values[8] = v
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

func newBenchmark() *Benchmarks {
	return &Benchmarks{
		Values: []uint64{
			100,       // "cpu-sysbench-multi"
			100,       // "cpu-sysbench-single"
			1,         // "cpu-cores"
			256000000, // "ram-size"
			0,         // "storage-size"
			1000000,   // "net-download"
			1000000,   // "net-upload"
			0,         // "gpu-count"
			0,         // "gpu-mem"
			0,         // "gpu-eth-hashrate"
			0,         // "gpu-cash-hashrate"
			0,         // "gpu-redshift"
		},
	}
}

func benchmarkMap(b *sonm.Benchmarks) map[string]uint64 {
	v := Benchmarks(*b)
	return v.toMap()
}

type taskStatus struct {
	*sonm.TaskStatusReply
	id string
}
