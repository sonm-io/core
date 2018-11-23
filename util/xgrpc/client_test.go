package xgrpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSecureGRPCConnect(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	serPriv, err := crypto.GenerateKey()
	require.NoError(err)
	rot, serTLS, err := util.NewHitlessCertRotator(ctx, serPriv)
	require.NoError(err)
	defer rot.Close()
	serCreds := util.NewTLS(serTLS)
	server := NewServer(nil, Credentials(serCreds))
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(err)
	defer lis.Close()
	go func() {
		server.Serve(lis)
	}()

	t.Run("ClientWithTLS", func(t *testing.T) {
		clientPriv, err := crypto.GenerateKey()
		require.NoError(err)
		rot, clientTLS, err := util.NewHitlessCertRotator(ctx, clientPriv)
		require.NoError(err)
		defer rot.Close()
		var clientCreds = util.NewTLS(clientTLS)
		conn, err := NewClient(ctx, lis.Addr().String(), clientCreds, grpc.WithTimeout(time.Second), grpc.WithBlock())
		require.NoError(err)
		defer conn.Close()

		err = grpc.Invoke(ctx, "/DummyService/dummyMethod", nil, nil, conn)
		require.NotNil(err)
		st, ok := status.FromError(err)
		require.True(ok)
		require.True(st.Code() == codes.Unimplemented)
	})

	t.Run("ClientWithoutTLS", func(t *testing.T) {
		conn, err := NewClient(ctx, lis.Addr().String(), nil, grpc.WithBlock(), grpc.WithTimeout(time.Second))
		if err != nil {
			// On Linux we can have an error here due to failed TLS Handshake
			// It's expected behavior
			return
		}
		// If we got here, error must occur after the first call
		require.NotNil(conn)
		defer conn.Close()
		err = grpc.Invoke(ctx, "/DummyService/dummyMethod", nil, nil, conn)
		require.NotNil(err)
		st, ok := status.FromError(err)
		require.True(ok)
		require.True(st.Code() == codes.Unavailable)
	})
}
