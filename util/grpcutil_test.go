package util

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"testing"
	"time"

	"github.com/sonm-io/core/util/xgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/require"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func TestTLSGenCerts(t *testing.T) {
	priv, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}
	certPEM, keyPEM, err := GenerateCert(priv, time.Second*20)
	if err != nil {
		t.Fatalf("%v", err)
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("%v", err)
	}
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("%v", err)
	}
	_, err = checkCert(x509Cert)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSecureGRPCConnect(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	serPriv, err := ethcrypto.GenerateKey()
	require.NoError(err)
	rot, serTLS, err := NewHitlessCertRotator(ctx, serPriv)
	require.NoError(err)
	defer rot.Close()
	serCreds := NewTLS(serTLS)
	server := xgrpc.NewServer(nil, xgrpc.Credentials(serCreds))
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(err)
	defer lis.Close()
	go func() {
		server.Serve(lis)
	}()

	t.Run("ClientWithTLS", func(t *testing.T) {
		clientPriv, err := ethcrypto.GenerateKey()
		require.NoError(err)
		rot, clientTLS, err := NewHitlessCertRotator(ctx, clientPriv)
		require.NoError(err)
		defer rot.Close()
		var clientCreds = NewTLS(clientTLS)
		conn, err := xgrpc.NewClient(ctx, lis.Addr().String(), clientCreds, grpc.WithTimeout(time.Second), grpc.WithBlock())
		require.NoError(err)
		defer conn.Close()

		err = grpc.Invoke(ctx, "/DummyService/dummyMethod", nil, nil, conn)
		require.NotNil(err)
		st, ok := status.FromError(err)
		require.True(ok)
		require.True(st.Code() == codes.Unimplemented)
	})

	t.Run("ClientWithoutTLS", func(t *testing.T) {
		conn, err := xgrpc.NewClient(ctx, lis.Addr().String(), nil, grpc.WithBlock(), grpc.WithTimeout(time.Second))
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
