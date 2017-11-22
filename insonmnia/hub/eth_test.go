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

func TestETH_CheckDealExists(t *testing.T) {
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

func TestXXX(t *testing.T) {
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
			SpecificationHash: "qwerty",
		},
		nil)
	bC.EXPECT().GetDealInfo(big.NewInt(200)).AnyTimes().Return(
		&pb.Deal{
			SupplierID:        addr,
			BuyerID:           "client-addr",
			Status:            pb.DealStatus_PENDING,
			SpecificationHash: "qwerty",
		},
		nil)

	eeth := &eth{
		ctx: context.Background(),
		key: key,
		bc:  bC,
	}

	//
	req, err := structs.NewDealRequest(&pb.DealRequest{
		AskId:    addr,
		BidId:    "client-addr",
		SpecHash: "qwerty",
		Order:    &pb.Order{Slot: &pb.Slot{}},
	})
	assert.NoError(t, err)

	// call blocked for 30 secs
	found, err := eeth.WaitForDealCreated(req)

	assert.NoError(t, err)
	assert.True(t, found)
}
