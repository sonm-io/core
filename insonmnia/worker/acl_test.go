package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func testCtx() context.Context {
	return log.WithLogger(context.Background(), zap.NewNop())
}

type testDealInfoSupplier struct {
	Deal *pb.Deal
}

func (m *testDealInfoSupplier) GetDealInfo(ctx context.Context, id *sonm.ID) (*pb.DealInfoReply, error) {
	return &pb.DealInfoReply{
		Deal: m.Deal,
	}, nil
}

func makeDealInfoSupplier(t *testing.T, buyerId string, dealID string) DealInfoSupplier {
	id, err := pb.NewBigIntFromString(dealID)
	require.NoError(t, err)
	return &testDealInfoSupplier{
		Deal: &sonm.Deal{
			Id:         id,
			ConsumerID: pb.NewEthAddress(common.HexToAddress(buyerId)),
		},
	}
}

func TestContextDealMetaData(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(map[string][]string{
		"deal": {"66"},
	}))

	md := newContextDealExtractor()
	dealID, err := md(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, structs.DealID("66"), dealID)
}

func startTaskDealExtractor() DealExtractor {
	return newRequestDealExtractor(func(request interface{}) (structs.DealID, error) {
		return structs.DealID(request.(*pb.StartTaskRequest).GetDealID().Unwrap().String()), nil
	})
}

func TestDealAuthorization(t *testing.T) {
	ctx := peer.NewContext(testCtx(), &peer.Peer{
		AuthInfo: auth.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: addr},
	})

	request := &sonm.StartTaskRequest{
		DealID: pb.NewBigIntFromInt(66),
	}

	extractor := startTaskDealExtractor()
	authorization := newDealAuthorization(ctx, makeDealInfoSupplier(t, addr.Hex(), "66"), extractor)

	require.NoError(t, authorization.Authorize(ctx, request))
}

func TestDealAuthorizationErrors(t *testing.T) {
	ctx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: auth.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: addr},
	})

	request := &sonm.StartTaskRequest{
		DealID: pb.NewBigIntFromInt(66),
	}

	extractor := startTaskDealExtractor()
	au := newDealAuthorization(context.Background(), makeDealInfoSupplier(t, "0x100500", "66"), extractor)

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
		DealID: pb.NewBigIntFromInt(66),
	}

	extractor := startTaskDealExtractor()
	au := newDealAuthorization(context.Background(), makeDealInfoSupplier(t, "0x100500", "66"), extractor)

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
	mul := newAnyOfAuth(&magicAuthorizer{ok: true}, &magicAuthorizer{ok: true}, &magicAuthorizer{ok: true})
	err := mul.Authorize(context.Background(), nil)
	assert.NoError(t, err)

	mul = newAnyOfAuth(&magicAuthorizer{ok: false}, &magicAuthorizer{ok: false}, &magicAuthorizer{ok: true})
	err = mul.Authorize(context.Background(), nil)
	assert.NoError(t, err)

	mul = newAnyOfAuth(&magicAuthorizer{ok: true}, &magicAuthorizer{ok: false}, &magicAuthorizer{ok: false})
	err = mul.Authorize(context.Background(), nil)
	assert.NoError(t, err)

	mul = newAnyOfAuth(&magicAuthorizer{ok: false}, &magicAuthorizer{ok: false}, &magicAuthorizer{ok: false})

	err = mul.Authorize(context.Background(), nil)
	assert.Error(t, err)
}
