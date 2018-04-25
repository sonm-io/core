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
)

func mockDWH(ctrl *gomock.Controller, t sonm.OrderType) sonm.DWHClient {
	orders := []*sonm.DWHOrder{
		{Order: &sonm.Order{OrderType: t, Id: "111", Price: sonm.NewBigIntFromInt(111)}},
		{Order: &sonm.Order{OrderType: t, Id: "222", Price: sonm.NewBigIntFromInt(222)}},
		{Order: &sonm.Order{OrderType: t, Id: "333", Price: sonm.NewBigIntFromInt(333)}},
	}

	dwh := sonm.NewMockDWHClient(ctrl)
	dwh.EXPECT().GetMatchingOrders(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&sonm.DWHOrdersReply{Orders: orders}, nil)
	return dwh
}

func mockEth(ctrl *gomock.Controller) (blockchain.API, chan blockchain.DealOrError) {
	api := blockchain.NewMockAPI(ctrl)

	ch := make(chan blockchain.DealOrError)

	api.EXPECT().GetOrderInfo(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&sonm.Order{OrderStatus: sonm.OrderStatus_ORDER_ACTIVE}, nil)
	api.EXPECT().OpenDeal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
		Return(ch)

	return api, ch
}

func TestMatcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key, _ := crypto.GenerateKey()
	eth, dealChan := mockEth(ctrl)

	m := NewMatcher(&Config{
		Key:        key,
		PollDelay:  time.Second,
		QueryLimit: 10,
		DWH:        mockDWH(ctrl, sonm.OrderType_ASK),
		Eth:        eth,
	})

	target := &sonm.Order{
		Id:        "1",
		OrderType: sonm.OrderType_BID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		dealChan <- blockchain.DealOrError{Deal: &sonm.Deal{Id: "123"}, Err: nil}
	}()

	deal, err := m.CreateDealByOrder(ctx, target)
	require.NoError(t, err)
	require.NotNil(t, deal)
}

func TestMatcherFailedByTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key, _ := crypto.GenerateKey()
	eth, dealChan := mockEth(ctrl)

	m := NewMatcher(&Config{
		Key:        key,
		PollDelay:  time.Second,
		QueryLimit: 10,
		DWH:        mockDWH(ctrl, sonm.OrderType_BID),
		Eth:        eth,
	})

	target := &sonm.Order{
		Id:        "1",
		OrderType: sonm.OrderType_ASK,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(time.Second)
			dealChan <- blockchain.DealOrError{Deal: nil, Err: fmt.Errorf("TEST_%d: cannot create order", i)}
		}
	}()

	_, err := m.CreateDealByOrder(ctx, target)
	require.Error(t, err)
	assert.EqualError(t, err, "context deadline exceeded")
}
