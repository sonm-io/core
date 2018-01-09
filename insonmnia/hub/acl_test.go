package hub

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func makeDefaultOrder(t *testing.T, buyerId string) *structs.Order {
	slot := &pb.Slot{
		Duration:  uint64(structs.MinSlotDuration.Seconds()),
		Resources: &pb.Resources{},
	}

	order, err := structs.NewOrder(&pb.Order{
		ByuerID:        buyerId,
		Slot:           slot,
		PricePerSecond: pb.NewBigIntFromInt(1),
	})
	assert.NoError(t, err)
	return order
}

func makeHubWithOrder(t *testing.T, buyerId string, dealId DealID) *Hub {
	order := makeDefaultOrder(t, buyerId)
	return &Hub{
		deals: map[DealID]*DealMeta{dealId: {Order: *order}},
	}
}

func TestFieldDealMetaData(t *testing.T) {
	request := &sonm.HubStartTaskRequest{
		Deal: &sonm.Deal{
			Id: "0x42",
		},
	}

	md := fieldDealMetaData{}
	dealID, err := md.Deal(context.Background(), request)
	require.NoError(t, err)
	assert.Equal(t, DealID("0x42"), dealID)
}

func TestFieldDealMetaDataErrorsOnInvalidType(t *testing.T) {
	type Request struct {
		Deal string
	}
	request := &Request{
		Deal: "0x42",
	}

	md := fieldDealMetaData{}
	dealID, err := md.Deal(context.Background(), request)
	assert.Error(t, err)
	assert.Equal(t, DealID(""), dealID)
}

func TestFieldDealMetaDataErrorsOnInvalidInnerType(t *testing.T) {
	type Deal struct {
		Id int
	}
	type Request struct {
		Deal *Deal
	}
	request := &Request{
		Deal: &Deal{
			Id: 42,
		},
	}

	md := fieldDealMetaData{}
	dealID, err := md.Deal(context.Background(), request)
	assert.Error(t, err)
	assert.Equal(t, DealID(""), dealID)
}

func TestFieldDealMetaDataWallet(t *testing.T) {
	peerCtx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: auth.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
	})

	ctx := metadata.NewIncomingContext(peerCtx, metadata.MD(map[string][]string{
		"wallet": {"0x42"},
	}))

	md := contextDealMetaData{}
	dealID, err := md.Wallet(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, "0x42", dealID)
}

func TestFieldDealMetaDataWalletErrorsOnEmptyMD(t *testing.T) {
	peerCtx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: auth.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
	})

	ctx := metadata.NewIncomingContext(peerCtx, metadata.MD(map[string][]string{}))

	md := contextDealMetaData{}
	dealID, err := md.Wallet(ctx, nil)
	assert.Error(t, err)
	assert.Equal(t, "", dealID)
}

func TestContextDealMetaData(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(map[string][]string{
		"deal": {"0x42"},
	}))

	md := contextDealMetaData{}
	dealID, err := md.Deal(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, DealID("0x42"), dealID)
}

func TestDealAuthorization(t *testing.T) {
	wallet, err := util.NewSelfSignedWallet(key)
	require.NoError(t, err)

	access := util.NewWalletAccess(wallet)

	peerCtx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: auth.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
	})

	requestMD, err := access.GetRequestMetadata(nil)
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(peerCtx, metadata.MD(map[string][]string{
		"wallet": {requestMD["wallet"]},
	}))

	request := &sonm.HubStartTaskRequest{
		Deal: &sonm.Deal{
			Id: "0x42",
		},
	}

	md := fieldDealMetaData{}
	auth := newDealAuthorization(context.Background(), makeHubWithOrder(t, addr.Hex(), "0x42"), &md)

	require.NoError(t, auth.Authorize(ctx, request))
}

func TestDealAuthorizationErrors(t *testing.T) {
	wallet, err := util.NewSelfSignedWallet(key)
	require.NoError(t, err)

	access := util.NewWalletAccess(wallet)

	peerCtx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: auth.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
	})

	requestMD, err := access.GetRequestMetadata(nil)
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(peerCtx, metadata.MD(map[string][]string{
		"wallet": {requestMD["wallet"]},
	}))

	request := &sonm.HubStartTaskRequest{
		Deal: &sonm.Deal{
			Id: "0x42",
		},
	}

	md := fieldDealMetaData{}
	au := newDealAuthorization(context.Background(), makeHubWithOrder(t, "0x100500", "0x42"), &md)

	require.Error(t, au.Authorize(ctx, request))
}

func TestDealAuthorizationErrorsOnInvalidWallet(t *testing.T) {
	peerCtx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: auth.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
	})

	ctx := metadata.NewIncomingContext(peerCtx, metadata.MD(map[string][]string{
		"wallet": {"0x100500"},
	}))

	request := &sonm.HubStartTaskRequest{
		Deal: &sonm.Deal{
			Id: "0x42",
		},
	}

	md := fieldDealMetaData{}
	au := newDealAuthorization(context.Background(), makeHubWithOrder(t, "0x100500", "0x42"), &md)

	require.Error(t, au.Authorize(ctx, request))
}
