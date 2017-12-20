package hub

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

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
		AuthInfo: util.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
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
		AuthInfo: util.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
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
		AuthInfo: util.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
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
	auth := newDealAuthorization(context.Background(), &md)
	auth.(*dealAuthorization).allowedWallets[DealID("0x42")] = addr.Hex()

	require.NoError(t, auth.Authorize(ctx, request))
}

func TestDealAuthorizationErrors(t *testing.T) {
	wallet, err := util.NewSelfSignedWallet(key)
	require.NoError(t, err)

	access := util.NewWalletAccess(wallet)

	peerCtx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: util.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
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
	auth := newDealAuthorization(context.Background(), &md)
	auth.(*dealAuthorization).allowedWallets[DealID("0x42")] = "0x100500"

	require.Error(t, auth.Authorize(ctx, request))
}

func TestDealAuthorizationErrorsOnInvalidWallet(t *testing.T) {
	peerCtx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: util.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
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
	auth := newDealAuthorization(context.Background(), &md)
	auth.(*dealAuthorization).allowedWallets[DealID("0x42")] = "0x100500"

	require.Error(t, auth.Authorize(ctx, request))
}
