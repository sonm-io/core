package hub

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func TestWhitelistSuperuser(t *testing.T) {
	w := whitelist{
		superusers: map[string]struct{}{addr.Hex(): {}},
	}

	wallet, err := util.NewSelfSignedWallet(key)
	require.NoError(t, err)

	peerCtx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: util.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
	})

	ctx := metadata.NewIncomingContext(peerCtx, metadata.New(map[string]string{
		"wallet": wallet.Message,
	}))

	allowed, _, err := w.Allowed(ctx, "docker.io", "hello-world", "")
	assert.NoError(t, err)
	assert.True(t, allowed)

	w.superusers = map[string]struct{}{}
	allowed, _, err = w.Allowed(ctx, "docker.io", "hello-world", "")
	assert.False(t, allowed)
}
