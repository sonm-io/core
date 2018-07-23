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
	m := baseBenchmark()
	switch token {
	case "ETH":
		m["gpu-eth-hashrate"] = hashrate
	case "ZEC":
		m["gpu-cash-hashrate"] = hashrate
	case "NULL":
		m["gpu-redshift"] = hashrate
	}

	bench, err := sonm.NewBenchmarksFromMap(m)
	if err != nil {
		return nil, err
	}

	ord := &sonm.Order{
		OrderType:     sonm.OrderType_BID,
		Price:         sonm.NewBigInt(price),
		Netflags:      &sonm.NetFlags{Flags: sonm.NetworkOutbound},
		IdentityLevel: sonm.IdentityLevel_ANONYMOUS,
		// Blacklist:     "",
		Tag:        []byte(fmt.Sprintf("connor_v2_%s", strings.ToLower(token))),
		Benchmarks: bench,
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

func (co *Corder) getBenchmarkMap() map[string]uint64 {
	m := baseBenchmark()

	switch co.token {
	case "ETH":
		m["gpu-eth-hashrate"] = co.GetHashrate()
		m["gpu-mem"] = 3900e6
	case "ZEC":
		m["gpu-cash-hashrate"] = co.GetHashrate()
		m["gpu-mem"] = 900e6 // todo: I should find the right value for this
	case "NULL":
		m["gpu-redshift"] = co.GetHashrate()
		m["gpu-mem"] = 9e6
	}
	return m
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
			Benchmarks: co.getBenchmarkMap(),
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

func baseBenchmark() map[string]uint64 {
	return map[string]uint64{
		// todo: should be tuned in future
		"cpu-sysbench-multi":  100,
		"cpu-sysbench-single": 100,
		"cpu-cores":           1,
		"ram-size":            256000000,
		"storage-size":        0,
		"net-download":        1000000,
		"net-upload":          1000000,
		"gpu-count":           0,
		"gpu-mem":             0,
		"gpu-eth-hashrate":    0,
		"gpu-cash-hashrate":   0,
		"gpu-redshift":        0,
	}
}
