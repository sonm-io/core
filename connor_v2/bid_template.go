package connor

import (
	"math/big"

	"github.com/sonm-io/core/proto"
)

type bidTemplate func(price *big.Int, hashrate uint64) *sonm.BidOrder

func nullCoinTemplate(price *big.Int, _ uint64) *sonm.BidOrder {
	return &sonm.BidOrder{
		Price: &sonm.Price{
			PerSecond: sonm.NewBigInt(price),
		},
		Tag: "connor_v2_nullCoin",
		Counterparty: &sonm.EthAddress{
			Address: nil,
		},

		Resources: &sonm.BidResources{
			Network: &sonm.BidNetwork{
				Overlay:  false,
				Outbound: true,
				Incoming: false,
			},
			Benchmarks: map[string]uint64{
				"ram-size":            67108864, // 64Mb
				"cpu-cores":           1,
				"cpu-sysbench-single": 100,
				"cpu-sysbench-multi":  100,
				"net-download":        1048576,
				"net-upload":          1048576,
				"gpu-count":           0,
				"gpu-mem":             0,
			},
		},
	}
}

func ethTemplate(price *big.Int, hashrate uint64) *sonm.BidOrder {
	return &sonm.BidOrder{
		Price: &sonm.Price{
			PerSecond: sonm.NewBigInt(price),
		},
		Tag: "connor_v2_eth",
		Counterparty: &sonm.EthAddress{
			Address: nil,
		},

		Resources: &sonm.BidResources{
			Network: &sonm.BidNetwork{
				Overlay:  false,
				Outbound: true,
				Incoming: false,
			},
			Benchmarks: map[string]uint64{
				"ram-size":            268435456, // 256Mb
				"cpu-cores":           1,
				"cpu-sysbench-single": 100,
				"cpu-sysbench-multi":  100,
				"net-download":        1048576,
				"net-upload":          1048576,
				"gpu-count":           1,
				"gpu-mem":             3900000000,
			},
		},
	}
}

func newBidTemplate(token string) bidTemplate {
	switch token {
	case "NULL":
		return nullCoinTemplate
	case "ETH":
		return ethTemplate
	case "ZEC":
		// todo: find proper params for zec and xmr
		return nullCoinTemplate
	case "XMR":
		return nullCoinTemplate
	default:
		// should never happens
		return nullCoinTemplate
	}
}
