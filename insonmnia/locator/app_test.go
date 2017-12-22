package locator

import (
	"crypto/ecdsa"
	"crypto/tls"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/net/context"

	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/require"
)

var (
	key = getTestKey()
)

func getTestKey() *ecdsa.PrivateKey {
	k, _ := crypto.GenerateKey()
	return k
}

func TestStorage_PutExternal(t *testing.T) {
	app := NewApp(TestConfig("localhost:9090"), key)
	err := app.Init()
	require.NoError(t, err)

	go func() {
		if err := app.Serve(); err != nil {
			t.Errorf("App server failed: %s", err)
		}
	}()

	cert, key, err := util.GenerateCert(key)
	require.NoError(t, err)

	crt, err := tls.X509KeyPair(cert, key)
	require.NoError(t, err)

	creds := util.NewTLS(&tls.Config{Certificates: []tls.Certificate{crt}, InsecureSkipVerify: true})

	conn, err := util.MakeGrpcClient(context.Background(), "localhost:9090", creds)
	require.NoError(t, err)

	locatorClient := pb.NewLocatorClient(conn)

	_, err = locatorClient.Announce(context.Background(), &pb.AnnounceRequest{IpAddr: []string{"42.42.42.42"}})
	require.NoError(t, err, "Failed to securely announce")
}
