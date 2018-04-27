package hub

import (
	"context"
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/miner"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func testCtx() context.Context {
	logger := logging.BuildLogger(*logging.NewLevel(zapcore.DebugLevel))
	return log.WithLogger(context.Background(), logger)
}

func makeHubWithOrder(t *testing.T, ctx context.Context, buyerId string, dealId structs.DealID) *Hub {
	return &Hub{
		ctx: ctx,
		worker: &miner.Miner{
			Deals: map[structs.DealID]*structs.DealMeta{dealId: {Deal: &sonm.Deal{ConsumerID: pb.NewEthAddress(common.HexToAddress(buyerId))}}},
		},
	}
}

func TestFieldDealMetaData(t *testing.T) {
	request := &sonm.StartTaskRequest{
		Deal: &sonm.Deal{
			Id: "0x42",
		},
	}

	md := newFieldDealExtractor()
	dealID, err := md(context.Background(), request)
	require.NoError(t, err)
	assert.Equal(t, structs.DealID("0x42"), dealID)
}

func TestFieldDealMetaDataErrorsOnInvalidType(t *testing.T) {
	type Request struct {
		Deal string
	}
	request := &Request{
		Deal: "0x42",
	}

	md := newFieldDealExtractor()
	dealID, err := md(context.Background(), request)
	assert.Error(t, err)
	assert.Equal(t, structs.DealID(""), dealID)
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

	md := newFieldDealExtractor()
	dealID, err := md(context.Background(), request)
	assert.Error(t, err)
	assert.Equal(t, structs.DealID(""), dealID)
}

func TestContextDealMetaData(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(map[string][]string{
		"deal": {"0x42"},
	}))

	md := newContextDealExtractor()
	dealID, err := md(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, structs.DealID("0x42"), dealID)
}

func TestDealAuthorization(t *testing.T) {
	ctx := peer.NewContext(testCtx(), &peer.Peer{
		AuthInfo: auth.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: addr},
	})

	request := &sonm.StartTaskRequest{
		Deal: &sonm.Deal{
			Id: "0x42",
		},
	}

	md := newFieldDealExtractor()
	authorization := newDealAuthorization(ctx, makeHubWithOrder(t, ctx, addr.Hex(), "0x42"), md)

	require.NoError(t, authorization.Authorize(ctx, request))
}

func TestDealAuthorizationErrors(t *testing.T) {
	ctx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: auth.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: addr},
	})

	request := &sonm.StartTaskRequest{
		Deal: &sonm.Deal{
			Id: "0x42",
		},
	}

	md := newFieldDealExtractor()
	au := newDealAuthorization(context.Background(), makeHubWithOrder(t, ctx, "0x100500", "0x42"), md)

	require.Error(t, au.Authorize(ctx, request))
}

func TestDealAuthorizationErrorsOnInvalidWallet(t *testing.T) {
	peerCtx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: auth.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
	})

	ctx := metadata.NewIncomingContext(peerCtx, metadata.MD(map[string][]string{
		"wallet": {"0x100500"},
	}))

	request := &sonm.StartTaskRequest{
		Deal: &sonm.Deal{
			Id: "0x42",
		},
	}

	md := newFieldDealExtractor()
	au := newDealAuthorization(context.Background(), makeHubWithOrder(t, ctx, "0x100500", "0x42"), md)

	require.Error(t, au.Authorize(ctx, request))
}

type magicAuthorizer struct {
	ok bool
}

func (ma *magicAuthorizer) Authorize(ctx context.Context, request interface{}) error {
	if ma.ok {
		return nil
	}

	return errors.New("sorry")
}

func TestMultiAuth(t *testing.T) {
	mul := newMultiAuth(&magicAuthorizer{ok: true}, &magicAuthorizer{ok: true}, &magicAuthorizer{ok: true})
	err := mul.Authorize(context.Background(), nil)
	assert.NoError(t, err)

	mul = newMultiAuth(&magicAuthorizer{ok: false}, &magicAuthorizer{ok: false}, &magicAuthorizer{ok: true})
	err = mul.Authorize(context.Background(), nil)
	assert.NoError(t, err)

	mul = newMultiAuth(&magicAuthorizer{ok: true}, &magicAuthorizer{ok: false}, &magicAuthorizer{ok: false})
	err = mul.Authorize(context.Background(), nil)
	assert.NoError(t, err)

	mul = newMultiAuth(&magicAuthorizer{ok: false}, &magicAuthorizer{ok: false}, &magicAuthorizer{ok: false})

	err = mul.Authorize(context.Background(), nil)
	assert.Error(t, err)
}
