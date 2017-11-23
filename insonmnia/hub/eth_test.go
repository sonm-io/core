package hub

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func makeTestKey() (string, *ecdsa.PrivateKey) {
	key, _ := ethcrypto.GenerateKey()
	addr := util.PubKeyToAddr(key.PublicKey)
	return addr, key
}

func TestEth_CheckDealExists(t *testing.T) {
	addr, key := makeTestKey()
	bC := blockchain.NewMockBlockchainer(gomock.NewController(t))
	bC.EXPECT().GetDeals(addr).AnyTimes().Return([]*big.Int{big.NewInt(1), big.NewInt(2)}, nil)
	bC.EXPECT().GetDealInfo(big.NewInt(1)).AnyTimes().Return(&pb.Deal{SupplierID: addr, Status: pb.DealStatus_ACCEPTED}, nil)
	bC.EXPECT().GetDealInfo(big.NewInt(2)).AnyTimes().Return(&pb.Deal{SupplierID: addr, Status: pb.DealStatus_CLOSED}, nil)
	bC.EXPECT().GetDealInfo(big.NewInt(3)).AnyTimes().Return(&pb.Deal{SupplierID: "anotherEthAddress", Status: pb.DealStatus_CLOSED}, nil)

	eeth := &eth{
		ctx: context.Background(),
		key: key,
		bc:  bC,
	}

	exists, err := eeth.CheckDealExists("1")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = eeth.CheckDealExists("2")
	assert.NoError(t, err)
	assert.False(t, exists)

	exists, err = eeth.CheckDealExists("3")
	assert.NoError(t, err)
	assert.False(t, exists)

	exists, err = eeth.CheckDealExists("qwerty")
	assert.Error(t, err)
}

func TestEth_WaitForDealCreated(t *testing.T) {
	addr, key := makeTestKey()
	bC := blockchain.NewMockBlockchainer(gomock.NewController(t))
	bC.EXPECT().GetDeals("client-addr").AnyTimes().Return(
		[]*big.Int{
			big.NewInt(100),
			big.NewInt(200),
		},
		nil)

	bC.EXPECT().GetDealInfo(big.NewInt(100)).AnyTimes().Return(
		&pb.Deal{
			SupplierID:        addr,
			BuyerID:           "client-addr",
			Status:            pb.DealStatus_ACCEPTED,
			SpecificationHash: "aaa",
		},
		nil)
	bC.EXPECT().GetDealInfo(big.NewInt(200)).AnyTimes().Return(
		&pb.Deal{
			SupplierID:        addr,
			BuyerID:           "client-addr",
			Status:            pb.DealStatus_PENDING,
			SpecificationHash: "bbb",
		},
		nil)

	eeth := &eth{
		ctx: context.Background(),
		key: key,
		bc:  bC,
	}

	req, err := structs.NewDealRequest(&pb.DealRequest{
		AskId:    addr,
		BidId:    "client-addr",
		SpecHash: "bbb",
		Order:    &pb.Order{Slot: &pb.Slot{}},
	})
	assert.NoError(t, err)

	found, err := eeth.WaitForDealCreated(req)
	assert.NoError(t, err)
	assert.Equal(t, "bbb", found.SpecificationHash)
	assert.Equal(t, "client-addr", found.BuyerID)
	assert.Equal(t, addr, found.SupplierID)
}

func TestEth_CheckDealExists2(t *testing.T) {
	addr, key := makeTestKey()
	bC := blockchain.NewMockBlockchainer(gomock.NewController(t))
	bC.EXPECT().GetDeals("client-addr").AnyTimes().Return(
		[]*big.Int{
			big.NewInt(100),
		},
		nil)

	bC.EXPECT().GetDealInfo(big.NewInt(100)).AnyTimes().Return(
		&pb.Deal{
			SupplierID:        addr,
			BuyerID:           "client-addr",
			Status:            pb.DealStatus_CLOSED,
			SpecificationHash: "aaa",
		},
		nil)

	eeth := &eth{
		ctx: context.Background(),
		key: key,
		bc:  bC,
	}

	req, err := structs.NewDealRequest(&pb.DealRequest{
		AskId:    addr,
		BidId:    "client-addr",
		SpecHash: "aaa",
		Order:    &pb.Order{Slot: &pb.Slot{}},
	})
	assert.NoError(t, err)

	found, err := eeth.WaitForDealCreated(req)
	assert.Nil(t, found)
	assert.Error(t, err)
	assert.EqualError(t, err, "context deadline exceeded")
}
