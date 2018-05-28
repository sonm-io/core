package node

import (
	"context"
	"crypto/ecdsa"
	"net"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/dwh"
	"github.com/sonm-io/core/insonmnia/npp"
	"github.com/sonm-io/core/insonmnia/npp/relay"
	"github.com/sonm-io/core/insonmnia/npp/rendezvous"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/netutil"
	"github.com/sonm-io/core/util/xgrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestNode(t *testing.T, key *ecdsa.PrivateKey) *Node {
	rvAddr, err := auth.NewAddr("3f46ed4f779fd378f630d8cd996796c69a7738d2@127.0.0.1:12345")
	require.NoError(t, err)

	relayAddr := netutil.TCPAddr{}
	relayAddr.TCPAddr = net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345}

	ctx := context.Background()
	nod, err := New(ctx, &Config{
		Node: nodeConfig{
			HttpBindPort:            0,
			BindPort:                0,
			AllowInsecureConnection: false,
		},
		NPP: npp.Config{
			Rendezvous: rendezvous.Config{
				Endpoints: []auth.Addr{*rvAddr},
			},
			Relay: relay.Config{
				Endpoints: []netutil.TCPAddr{relayAddr},
			},
		},
		DWH: dwh.YAMLConfig{
			Endpoint: "3f46ed4f779fd378f630d8cd996796c69a7738d2@127.0.0.1:12345",
		},
	}, key)
	require.NoError(t, err)

	return nod
}

func TestConnectWithoutTLS(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	nod := newTestNode(t, key)
	go nod.Serve()

	for {
		if len(nod.listeners) > 0 {
			break
		}
		// wait for grpc server up and running
		time.Sleep(50 * time.Millisecond)
	}

	ctx := context.Background()
	nodeEndpoint := nod.listeners[0].Addr().String()
	cc, err := xgrpc.NewClient(ctx, nodeEndpoint, nil)
	require.NoError(t, err)

	_, err = sonm.NewMarketClient(cc).GetOrderByID(ctx, &sonm.ID{Id: "1"})
	require.Error(t, err)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok, "error must be provided by GRPC")
	assert.Equal(t, codes.Unavailable, grpcErr.Code(), "must be unable to connect")
}

func TestConnectWithValidKeyWithoutWallet(t *testing.T) {
	nodeKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	nod := newTestNode(t, nodeKey)
	go nod.Serve()

	for {
		if len(nod.listeners) > 0 {
			break
		}
		// wait for grpc server up and running
		time.Sleep(50 * time.Millisecond)
	}

	ctx := context.Background()
	nodeEndpoint := nod.listeners[0].Addr().String()

	_, TLSConfig, err := util.NewHitlessCertRotator(ctx, nodeKey)
	require.NoError(t, err)

	cc, err := xgrpc.NewClient(ctx, nodeEndpoint, util.NewTLS(TLSConfig))
	require.NoError(t, err)

	_, err = sonm.NewMarketClient(cc).GetOrderByID(ctx, &sonm.ID{Id: "1"})
	require.NoError(t, err)
}

func TestConnectWithInvalidKeyWithoutWallet(t *testing.T) {
	nodeKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	nod := newTestNode(t, nodeKey)
	go nod.Serve()

	for {
		if len(nod.listeners) > 0 {
			break
		}
		// wait for grpc server up and running
		time.Sleep(50 * time.Millisecond)
	}

	ctx := context.Background()
	nodeEndpoint := nod.listeners[0].Addr().String()

	clientKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	_, TLSConfig, err := util.NewHitlessCertRotator(ctx, clientKey)
	require.NoError(t, err)

	cc, err := xgrpc.NewClient(ctx, nodeEndpoint, util.NewTLS(TLSConfig))
	require.NoError(t, err)

	_, err = sonm.NewMarketClient(cc).GetOrderByID(ctx, &sonm.ID{Id: "1"})
	require.Error(t, err)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok, "error must be provided by GRPC")
	assert.Equal(t, codes.Unavailable, grpcErr.Code(), "must be unable to connect")
}

func TestConnectWithInvalidKeyWithWallet(t *testing.T) {
	nodeKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	nod := newTestNode(t, nodeKey)
	go nod.Serve()

	for {
		if len(nod.listeners) > 0 {
			break
		}
		// wait for grpc server up and running
		time.Sleep(50 * time.Millisecond)
	}

	ctx := context.Background()
	nodeEndpoint := nod.listeners[0].Addr().String()

	clientKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	_, TLSConfig, err := util.NewHitlessCertRotator(ctx, clientKey)
	require.NoError(t, err)

	creds := auth.NewWalletAuthenticator(util.NewTLS(TLSConfig), crypto.PubkeyToAddress(clientKey.PublicKey))
	cc, err := xgrpc.NewClient(ctx, nodeEndpoint, creds)
	require.NoError(t, err)

	_, err = sonm.NewMarketClient(cc).GetOrderByID(ctx, &sonm.ID{Id: "1"})
	require.Error(t, err)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok, "error must be provided by GRPC")
	assert.Equal(t, codes.Unavailable, grpcErr.Code(), "must be unable to connect ")
}

func TestConnectWithValidKeyWithWallet(t *testing.T) {
	nodeKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	nod := newTestNode(t, nodeKey)
	go nod.Serve()

	for {
		if len(nod.listeners) > 0 {
			break
		}
		// wait for grpc server up and running
		time.Sleep(50 * time.Millisecond)
	}

	ctx := context.Background()
	nodeEndpoint := nod.listeners[0].Addr().String()

	_, TLSConfig, err := util.NewHitlessCertRotator(ctx, nodeKey)
	require.NoError(t, err)

	creds := auth.NewWalletAuthenticator(util.NewTLS(TLSConfig), crypto.PubkeyToAddress(nodeKey.PublicKey))
	cc, err := xgrpc.NewClient(ctx, nodeEndpoint, creds)
	require.NoError(t, err)

	_, err = sonm.NewMarketClient(cc).GetOrderByID(ctx, &sonm.ID{Id: "1"})
	require.NoError(t, err)
}
