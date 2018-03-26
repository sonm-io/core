package dealer

import (
	"context"
	"crypto/ecdsa"
	"io"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func newHubMock(ctrl *gomock.Controller) (sonm.HubClient, io.Closer) {
	hub := NewMockHubClient(ctrl)
	hub.EXPECT().ProposeDeal(gomock.Any(), gomock.Any()).AnyTimes().Return(&sonm.Empty{}, nil)
	hub.EXPECT().ApproveDeal(gomock.Any(), gomock.Any()).AnyTimes().Return(&sonm.Empty{}, nil)
	return hub, &mockConn{}
}

func newEthKey() *ecdsa.PrivateKey {
	k, _ := crypto.GenerateKey()
	return k
}

func TestDealer_Deal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	key := newEthKey()

	bc := blockchain.NewMockBlockchainer(ctrl)
	bc.EXPECT().OpenDealPending(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().Return(big.NewInt(1), nil)

	hub, closr := newHubMock(ctrl)
	defer closr.Close()

	dlr := NewDealer(key, hub, bc, time.Second)

	ask := &sonm.Order{Id: "askid", SupplierID: "test2"}
	bid := &sonm.Order{
		Id:             "bidid",
		SupplierID:     "test1",
		PricePerSecond: sonm.NewBigIntFromInt(100),
		Slot:           &sonm.Slot{Duration: 500},
	}

	id, err := dlr.Deal(ctx, bid, ask)
	assert.NoError(t, err)
	assert.True(t, id.Uint64() > 0)
}

func TestDealer_DealFailedAndCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	key := newEthKey()

	bc := blockchain.NewMockBlockchainer(ctrl)
	bc.EXPECT().OpenDealPending(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().Return(big.NewInt(1), nil)
	bc.EXPECT().CloseDealPending(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		MinTimes(1).Return(nil)

	hub := NewMockHubClient(ctrl)
	hub.EXPECT().ApproveDeal(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.New("test"))

	dlr := NewDealer(key, hub, bc, time.Second)

	ask := &sonm.Order{Id: "askid", SupplierID: "test2"}
	bid := &sonm.Order{
		Id:             "bidid",
		SupplierID:     "test1",
		PricePerSecond: sonm.NewBigIntFromInt(100),
		Slot:           &sonm.Slot{Duration: 500},
	}

	id, err := dlr.Deal(ctx, bid, ask)
	assert.Error(t, err)
	assert.Nil(t, id)

}

type mockConn struct{}

func (c *mockConn) Close() error { return nil }
