package blockchain

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func privateKey(t *testing.T) *ecdsa.PrivateKey {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	require.NotNil(t, key)

	return key
}

func TestErrNiceOpenDealInvalidAskType(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	market := NewMockMarketAPI(controller)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(42)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_BID,
		OrderStatus: sonm.OrderStatus_ORDER_INACTIVE,
	}, nil)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(88)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_BID,
		OrderStatus: sonm.OrderStatus_ORDER_INACTIVE,
	}, nil)

	profiles := NewMockProfileRegistryAPI(controller)
	blacklist := NewMockBlacklistAPI(controller)

	niceMarket := &niceMarketAPI{
		MarketAPI: market,
		profiles:  profiles,
		blacklist: blacklist,
	}

	deal, err := niceMarket.OpenDeal(context.Background(), privateKey(t), big.NewInt(42), big.NewInt(88))

	require.Error(t, err)
	require.Nil(t, deal)

	assert.EqualError(t, err, "ask must have ASK type, but it is BID")
}

func TestErrNiceOpenDealInvalidBidType(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	market := NewMockMarketAPI(controller)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(42)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_ASK,
		OrderStatus: sonm.OrderStatus_ORDER_INACTIVE,
	}, nil)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(88)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_ASK,
		OrderStatus: sonm.OrderStatus_ORDER_INACTIVE,
	}, nil)

	profiles := NewMockProfileRegistryAPI(controller)
	blacklist := NewMockBlacklistAPI(controller)

	niceMarket := &niceMarketAPI{
		MarketAPI: market,
		profiles:  profiles,
		blacklist: blacklist,
	}

	deal, err := niceMarket.OpenDeal(context.Background(), privateKey(t), big.NewInt(42), big.NewInt(88))

	require.Error(t, err)
	require.Nil(t, deal)

	assert.EqualError(t, err, "bid must have BID type, but it is ASK")
}

func TestErrNiceOpenDealInactiveAsk(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	market := NewMockMarketAPI(controller)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(42)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_ASK,
		OrderStatus: sonm.OrderStatus_ORDER_INACTIVE,
	}, nil)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(88)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_BID,
		OrderStatus: sonm.OrderStatus_ORDER_INACTIVE,
	}, nil)

	profiles := NewMockProfileRegistryAPI(controller)
	blacklist := NewMockBlacklistAPI(controller)

	niceMarket := &niceMarketAPI{
		MarketAPI: market,
		profiles:  profiles,
		blacklist: blacklist,
	}

	deal, err := niceMarket.OpenDeal(context.Background(), privateKey(t), big.NewInt(42), big.NewInt(88))

	require.Error(t, err)
	require.Nil(t, deal)

	assert.EqualError(t, err, "ask order must be active, but it is ORDER_INACTIVE")
}

func TestErrNiceOpenDealInactiveBid(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	market := NewMockMarketAPI(controller)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(42)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_ASK,
		OrderStatus: sonm.OrderStatus_ORDER_ACTIVE,
	}, nil)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(88)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_BID,
		OrderStatus: sonm.OrderStatus_ORDER_INACTIVE,
	}, nil)

	profiles := NewMockProfileRegistryAPI(controller)
	blacklist := NewMockBlacklistAPI(controller)

	niceMarket := &niceMarketAPI{
		MarketAPI: market,
		profiles:  profiles,
		blacklist: blacklist,
	}

	deal, err := niceMarket.OpenDeal(context.Background(), privateKey(t), big.NewInt(42), big.NewInt(88))

	require.Error(t, err)
	require.Nil(t, deal)

	assert.EqualError(t, err, "bid order must be active, but it is ORDER_INACTIVE")
}

func TestErrNiceOpenDealBidPriceLesserThanAskPrice(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	market := NewMockMarketAPI(controller)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(42)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_ASK,
		OrderStatus: sonm.OrderStatus_ORDER_ACTIVE,
		Price:       sonm.NewBigIntFromInt(500),
	}, nil)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(88)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_BID,
		OrderStatus: sonm.OrderStatus_ORDER_ACTIVE,
		Price:       sonm.NewBigIntFromInt(499),
	}, nil)

	profiles := NewMockProfileRegistryAPI(controller)
	blacklist := NewMockBlacklistAPI(controller)

	niceMarket := &niceMarketAPI{
		MarketAPI: market,
		profiles:  profiles,
		blacklist: blacklist,
	}

	deal, err := niceMarket.OpenDeal(context.Background(), privateKey(t), big.NewInt(42), big.NewInt(88))

	require.Error(t, err)
	require.Nil(t, deal)

	assert.EqualError(t, err, "bid price 499 must be >= ask price 500")
}

func TestErrNiceOpenDealBidDurationLesserThanAskDuration(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	market := NewMockMarketAPI(controller)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(42)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_ASK,
		OrderStatus: sonm.OrderStatus_ORDER_ACTIVE,
		Price:       sonm.NewBigIntFromInt(500),
		Duration:    100500,
	}, nil)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(88)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_BID,
		OrderStatus: sonm.OrderStatus_ORDER_ACTIVE,
		Price:       sonm.NewBigIntFromInt(500),
		Duration:    100501,
	}, nil)

	profiles := NewMockProfileRegistryAPI(controller)
	blacklist := NewMockBlacklistAPI(controller)

	niceMarket := &niceMarketAPI{
		MarketAPI: market,
		profiles:  profiles,
		blacklist: blacklist,
	}

	deal, err := niceMarket.OpenDeal(context.Background(), privateKey(t), big.NewInt(42), big.NewInt(88))

	require.Error(t, err)
	require.Nil(t, deal)

	assert.EqualError(t, err, "bid duration 100501 must be <= ask duration 100500")
}

func TestErrNiceOpenDealBidNetflagsNotFitInAskNetflags(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	market := NewMockMarketAPI(controller)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(42)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_ASK,
		OrderStatus: sonm.OrderStatus_ORDER_ACTIVE,
		Netflags:    sonm.NetFlagsFromBoolSlice([]bool{true, true, false}),
	}, nil)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(88)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_BID,
		OrderStatus: sonm.OrderStatus_ORDER_ACTIVE,
		Netflags:    sonm.NetFlagsFromBoolSlice([]bool{true, true, true}),
	}, nil)

	profiles := NewMockProfileRegistryAPI(controller)
	blacklist := NewMockBlacklistAPI(controller)

	niceMarket := &niceMarketAPI{
		MarketAPI: market,
		profiles:  profiles,
		blacklist: blacklist,
	}

	deal, err := niceMarket.OpenDeal(context.Background(), privateKey(t), big.NewInt(42), big.NewInt(88))

	require.Error(t, err)
	require.Nil(t, deal)

	assert.EqualError(t, err, "bid netflags [true true true] must fit in ask netflags [true true false]")
}

func TestErrNiceOpenDealBidBenchmarksNotFitInAskBenchmarks(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	market := NewMockMarketAPI(controller)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(42)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_ASK,
		OrderStatus: sonm.OrderStatus_ORDER_ACTIVE,
		Benchmarks:  &sonm.Benchmarks{Values: []uint64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200}},
	}, nil)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(88)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_BID,
		OrderStatus: sonm.OrderStatus_ORDER_ACTIVE,
		Benchmarks:  &sonm.Benchmarks{Values: []uint64{100, 200, 300, 400, 542, 600, 700, 800, 900, 1000, 1100, 1200}},
	}, nil)
	market.EXPECT().GetNumBenchmarks(gomock.Any()).Times(1).Return(uint64(12), nil)

	profiles := NewMockProfileRegistryAPI(controller)
	blacklist := NewMockBlacklistAPI(controller)

	niceMarket := &niceMarketAPI{
		MarketAPI: market,
		profiles:  profiles,
		blacklist: blacklist,
	}

	deal, err := niceMarket.OpenDeal(context.Background(), privateKey(t), big.NewInt(42), big.NewInt(88))

	require.Error(t, err)
	require.Nil(t, deal)

	assert.EqualError(t, err, "benchmark matching failed: id=4 bid benchmark 542 must be <= ask benchmark 500")
}

func TestErrNiceOpenDealAskCounterpartyIsSetButMismatchWithBid(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	market := NewMockMarketAPI(controller)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(42)).Times(1).Return(&sonm.Order{
		OrderType:      sonm.OrderType_ASK,
		OrderStatus:    sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID:       sonm.NewEthAddress(common.HexToAddress("0x0000000000000000000000000000000000000042")),
		CounterpartyID: sonm.NewEthAddress(common.HexToAddress("0x0000000000000000000000000000000000000001")),
	}, nil)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(88)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_BID,
		OrderStatus: sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID:    sonm.NewEthAddress(common.HexToAddress("0x0000000000000000000000000000000000000088")),
	}, nil)
	market.EXPECT().GetNumBenchmarks(gomock.Any()).Times(1).Return(uint64(12), nil)
	market.EXPECT().GetMaster(gomock.Any(), common.HexToAddress("0x0000000000000000000000000000000000000042")).Times(1).Return(common.HexToAddress("0x0000000000000000000000000000000000000142"), nil)
	market.EXPECT().GetMaster(gomock.Any(), common.HexToAddress("0x0000000000000000000000000000000000000088")).Times(1).Return(common.HexToAddress("0x0000000000000000000000000000000000000188"), nil)

	profiles := NewMockProfileRegistryAPI(controller)
	blacklist := NewMockBlacklistAPI(controller)

	niceMarket := &niceMarketAPI{
		MarketAPI: market,
		profiles:  profiles,
		blacklist: blacklist,
	}

	deal, err := niceMarket.OpenDeal(context.Background(), privateKey(t), big.NewInt(42), big.NewInt(88))

	require.Error(t, err)
	require.Nil(t, deal)

	assert.EqualError(t, err, "ask counterparty 0x0000000000000000000000000000000000000001 doesn't match with bid master 0x0000000000000000000000000000000000000188")
}

func TestErrNiceOpenDealBidCounterpartyIsSetButMismatchWithAsk(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	market := NewMockMarketAPI(controller)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(42)).Times(1).Return(&sonm.Order{
		OrderType:   sonm.OrderType_ASK,
		OrderStatus: sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID:    sonm.NewEthAddress(common.HexToAddress("0x0000000000000000000000000000000000000042")),
	}, nil)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(88)).Times(1).Return(&sonm.Order{
		OrderType:      sonm.OrderType_BID,
		OrderStatus:    sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID:       sonm.NewEthAddress(common.HexToAddress("0x0000000000000000000000000000000000000088")),
		CounterpartyID: sonm.NewEthAddress(common.HexToAddress("0x0000000000000000000000000000000000000001")),
	}, nil)
	market.EXPECT().GetNumBenchmarks(gomock.Any()).Times(1).Return(uint64(12), nil)
	market.EXPECT().GetMaster(gomock.Any(), common.HexToAddress("0x0000000000000000000000000000000000000042")).Times(1).Return(common.HexToAddress("0x0000000000000000000000000000000000000142"), nil)
	market.EXPECT().GetMaster(gomock.Any(), common.HexToAddress("0x0000000000000000000000000000000000000088")).Times(1).Return(common.HexToAddress("0x0000000000000000000000000000000000000188"), nil)

	profiles := NewMockProfileRegistryAPI(controller)
	blacklist := NewMockBlacklistAPI(controller)

	niceMarket := &niceMarketAPI{
		MarketAPI: market,
		profiles:  profiles,
		blacklist: blacklist,
	}

	deal, err := niceMarket.OpenDeal(context.Background(), privateKey(t), big.NewInt(42), big.NewInt(88))

	require.Error(t, err)
	require.Nil(t, deal)

	assert.EqualError(t, err, "bid counterparty 0x0000000000000000000000000000000000000001 doesn't match with ask master 0x0000000000000000000000000000000000000142")
}

func TestErrNiceOpenDealBidIdentityMismatch(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	market := NewMockMarketAPI(controller)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(42)).Times(1).Return(&sonm.Order{
		OrderType:     sonm.OrderType_ASK,
		OrderStatus:   sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID:      sonm.NewEthAddress(common.HexToAddress("0x0000000000000000000000000000000000000042")),
		IdentityLevel: sonm.IdentityLevel_ANONYMOUS,
	}, nil)
	market.EXPECT().GetOrderInfo(gomock.Any(), big.NewInt(88)).Times(1).Return(&sonm.Order{
		OrderType:     sonm.OrderType_BID,
		OrderStatus:   sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID:      sonm.NewEthAddress(common.HexToAddress("0x0000000000000000000000000000000000000088")),
		IdentityLevel: sonm.IdentityLevel_REGISTERED,
	}, nil)
	market.EXPECT().GetNumBenchmarks(gomock.Any()).Times(1).Return(uint64(12), nil)
	market.EXPECT().GetMaster(gomock.Any(), common.HexToAddress("0x0000000000000000000000000000000000000042")).Times(1).Return(common.HexToAddress("0x0000000000000000000000000000000000000000"), nil)
	market.EXPECT().GetMaster(gomock.Any(), common.HexToAddress("0x0000000000000000000000000000000000000088")).Times(1).Return(common.HexToAddress("0x0000000000000000000000000000000000000000"), nil)

	profiles := NewMockProfileRegistryAPI(controller)
	profiles.EXPECT().GetProfileLevel(gomock.Any(), common.HexToAddress("0x0000000000000000000000000000000000000042")).Return(sonm.IdentityLevel_ANONYMOUS, nil)
	profiles.EXPECT().GetProfileLevel(gomock.Any(), common.HexToAddress("0x0000000000000000000000000000000000000088")).Return(sonm.IdentityLevel_REGISTERED, nil)
	blacklist := NewMockBlacklistAPI(controller)

	niceMarket := &niceMarketAPI{
		MarketAPI: market,
		profiles:  profiles,
		blacklist: blacklist,
	}

	deal, err := niceMarket.OpenDeal(context.Background(), privateKey(t), big.NewInt(42), big.NewInt(88))

	require.Error(t, err)
	require.Nil(t, deal)

	assert.EqualError(t, err, "ask identity ANONYMOUS must be >= bid author identity REGISTERED")
}
