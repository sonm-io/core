package matcher

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func mockDWH(ctrl *gomock.Controller, t sonm.OrderType) sonm.DWHClient {
	orders := []*sonm.DWHOrder{
		{Order: &sonm.Order{OrderType: t, Id: sonm.NewBigIntFromInt(111), Price: sonm.NewBigIntFromInt(111)}},
		{Order: &sonm.Order{OrderType: t, Id: sonm.NewBigIntFromInt(222), Price: sonm.NewBigIntFromInt(222)}},
		{Order: &sonm.Order{OrderType: t, Id: sonm.NewBigIntFromInt(333), Price: sonm.NewBigIntFromInt(333)}},
	}

	dwh := sonm.NewMockDWHClient(ctrl)
	dwh.EXPECT().GetMatchingOrders(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&sonm.DWHOrdersReply{Orders: orders}, nil)
	return dwh
}

func TestMatcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key, _ := crypto.GenerateKey()

	eth := blockchain.NewMockAPI(ctrl)
	marketApi := blockchain.NewMockMarketAPI(ctrl)
	marketApi.EXPECT().GetOrderInfo(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&sonm.Order{OrderStatus: sonm.OrderStatus_ORDER_ACTIVE}, nil)
	marketApi.EXPECT().OpenDeal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
		Return(&sonm.Deal{Id: sonm.NewBigIntFromInt(123)}, nil)
	eth.EXPECT().Market().AnyTimes().Return(marketApi)

	m, err := NewMatcher(&Config{
		Key:        key,
		PollDelay:  time.Second,
		QueryLimit: 10,
		DWH:        mockDWH(ctrl, sonm.OrderType_ASK),
		Eth:        eth,
		Log:        zap.NewNop().Sugar(),
	})

	require.NoError(t, err)

	target := &sonm.Order{
		Id:        sonm.NewBigIntFromInt(1),
		OrderType: sonm.OrderType_BID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deal, err := m.CreateDealByOrder(ctx, target)
	require.NoError(t, err)
	require.NotNil(t, deal)
}

func TestMatcherFailedByTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key, _ := crypto.GenerateKey()

	eth := blockchain.NewMockAPI(ctrl)
	marketApi := blockchain.NewMockMarketAPI(ctrl)
	marketApi.EXPECT().GetOrderInfo(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&sonm.Order{OrderStatus: sonm.OrderStatus_ORDER_ACTIVE}, nil)
	marketApi.EXPECT().OpenDeal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
		Return(nil, fmt.Errorf("TEST: cannot create order"))
	eth.EXPECT().Market().AnyTimes().Return(marketApi)

	m, err := NewMatcher(&Config{
		Key:        key,
		PollDelay:  time.Second,
		QueryLimit: 10,
		DWH:        mockDWH(ctrl, sonm.OrderType_BID),
		Eth:        eth,
		Log:        zap.NewNop().Sugar(),
	})

	require.NoError(t, err)
	target := &sonm.Order{
		Id:        sonm.NewBigIntFromInt(1),
		OrderType: sonm.OrderType_ASK,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	_, err = m.CreateDealByOrder(ctx, target)
	require.Error(t, err)
	assert.EqualError(t, err, "context deadline exceeded")
}

func TestMatcherConfigValidate(t *testing.T) {
	_, err := NewMatcher(&Config{
		PollDelay: 0,
		Key:       nil,
		DWH:       nil,
		Eth:       nil,
	})

	require.Error(t, err)
}
