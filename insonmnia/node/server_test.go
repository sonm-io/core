package node

import (
	"context"
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func nopInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func nopStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return handler(srv, ss)
}

func newTestTLS(t *testing.T, privateKey *ecdsa.PrivateKey) credentials.TransportCredentials {
	_, tlsConfig, err := util.NewHitlessCertRotator(context.Background(), privateKey)
	require.NoError(t, err)

	return util.NewTLS(tlsConfig)
}

func newTestKey(t *testing.T) *ecdsa.PrivateKey {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	require.NotNil(t, key)

	return key
}

func servicesMock(c *gomock.Controller) *MockServices {
	marketServer := sonm.NewMockMarketServer(c)
	marketServer.EXPECT().GetOrderByID(gomock.Any(), gomock.Any()).AnyTimes().Return(&sonm.Order{}, nil)

	services := NewMockServices(c)
	services.EXPECT().Run(gomock.Any()).AnyTimes().Return(nil)
	services.EXPECT().Interceptor().Times(1).Return(nopInterceptor)
	services.EXPECT().StreamInterceptor().Times(1).Return(nopStreamInterceptor)
	services.EXPECT().RegisterGRPC(gomock.Any()).Times(1).Return(nil).Do(func(server *grpc.Server) error {
		sonm.RegisterMarketServer(server, marketServer)
		return nil
	})
	services.EXPECT().RegisterREST(gomock.Any()).Times(0).Return(nil)
	return services
}

func TestConnectWithoutTLS(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	marketServer := sonm.NewMockMarketServer(c)
	marketServer.EXPECT().GetOrderByID(gomock.Any(), gomock.Any()).Times(0)

	key := newTestKey(t)

	server, err := newServer(nodeConfig{}, servicesMock(c), WithGRPCServer(), WithGRPCSecure(newTestTLS(t, key), key))
	require.NoError(t, err)

	endpoints := server.LocalEndpoints()
	require.True(t, len(endpoints.GRPC) > 0)

	ctx := context.Background()

	go server.Serve(ctx)

	conn, err := xgrpc.NewClient(ctx, endpoints.GRPC[0].String(), nil)
	require.NoError(t, err)

	order, err := sonm.NewMarketClient(conn).GetOrderByID(ctx, &sonm.ID{Id: "1"})
	require.Error(t, err)
	require.Nil(t, order)

	stat, ok := status.FromError(err)
	assert.True(t, ok, "error must be provided by GRPC")
	assert.Equal(t, codes.Unavailable, stat.Code(), "must be unable to connect")
}

func TestConnectWithValidKeyWithoutWallet(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	key := newTestKey(t)
	transportCredentials := newTestTLS(t, key)

	server, err := newServer(nodeConfig{}, servicesMock(c), WithGRPCServer(), WithGRPCSecure(transportCredentials, key))
	require.NoError(t, err)

	endpoints := server.LocalEndpoints()
	require.True(t, len(endpoints.GRPC) > 0)

	ctx := context.Background()

	go server.Serve(ctx)

	conn, err := xgrpc.NewClient(ctx, endpoints.GRPC[0].String(), transportCredentials)
	require.NoError(t, err)

	order, err := sonm.NewMarketClient(conn).GetOrderByID(ctx, &sonm.ID{Id: "1"})
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, &sonm.Order{}, order)
}

func TestConnectWithInvalidKeyWithoutWallet(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	key := newTestKey(t)
	transportCredentials := newTestTLS(t, key)

	server, err := newServer(nodeConfig{}, servicesMock(c), WithGRPCServer(), WithGRPCSecure(transportCredentials, key))
	require.NoError(t, err)

	endpoints := server.LocalEndpoints()
	require.True(t, len(endpoints.GRPC) > 0)

	ctx := context.Background()

	go server.Serve(ctx)

	clientKey := newTestKey(t)
	clientTransportCredentials := newTestTLS(t, clientKey)

	conn, err := xgrpc.NewClient(ctx, endpoints.GRPC[0].String(), clientTransportCredentials)
	require.NoError(t, err)

	order, err := sonm.NewMarketClient(conn).GetOrderByID(ctx, &sonm.ID{Id: "1"})
	require.Error(t, err)
	require.Nil(t, order)

	stat, ok := status.FromError(err)
	assert.True(t, ok, "error must be provided by GRPC")
	assert.Equal(t, codes.Unavailable, stat.Code(), "must be unable to connect")
}

func TestConnectWithInvalidKeyWithWallet(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	key := newTestKey(t)
	transportCredentials := newTestTLS(t, key)

	server, err := newServer(nodeConfig{}, servicesMock(c), WithGRPCServer(), WithGRPCSecure(transportCredentials, key))
	require.NoError(t, err)

	endpoints := server.LocalEndpoints()
	require.True(t, len(endpoints.GRPC) > 0)

	ctx := context.Background()

	go server.Serve(ctx)

	clientKey := newTestKey(t)
	clientTransportCredentials := newTestTLS(t, clientKey)

	authenticator := auth.NewWalletAuthenticator(clientTransportCredentials, crypto.PubkeyToAddress(clientKey.PublicKey))
	cc, err := xgrpc.NewClient(ctx, endpoints.GRPC[0].String(), authenticator)
	require.NoError(t, err)

	order, err := sonm.NewMarketClient(cc).GetOrderByID(ctx, &sonm.ID{Id: "1"})
	require.Error(t, err)
	require.Nil(t, order)

	stat, ok := status.FromError(err)
	assert.True(t, ok, "error must be provided by GRPC")
	assert.Equal(t, codes.Unavailable, stat.Code(), "must be unable to connect ")
}

func TestConnectWithValidKeyWithWallet(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	key := newTestKey(t)
	transportCredentials := newTestTLS(t, key)

	server, err := newServer(nodeConfig{}, servicesMock(c), WithGRPCServer(), WithGRPCSecure(transportCredentials, key))
	require.NoError(t, err)

	endpoints := server.LocalEndpoints()
	require.True(t, len(endpoints.GRPC) > 0)

	ctx := context.Background()

	go server.Serve(ctx)

	clientTransportCredentials := newTestTLS(t, key)

	authenticator := auth.NewWalletAuthenticator(clientTransportCredentials, crypto.PubkeyToAddress(key.PublicKey))
	conn, err := xgrpc.NewClient(ctx, endpoints.GRPC[0].String(), authenticator)
	require.NoError(t, err)

	order, err := sonm.NewMarketClient(conn).GetOrderByID(ctx, &sonm.ID{Id: "1"})
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, &sonm.Order{}, order)
}
