package hub

import (
	"context"
	"crypto/ecdsa"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func walletCtx(t *testing.T, key *ecdsa.PrivateKey) context.Context {
	wallet, err := util.NewSelfSignedWallet(key)
	require.NoError(t, err)

	peerCtx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: util.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: common.Address{}},
	})

	ctx := metadata.NewIncomingContext(peerCtx, metadata.New(map[string]string{
		"wallet": wallet.Message,
	}))
	return ctx
}
func TestWhitelistSuperuser(t *testing.T) {
	w := whitelist{
		superusers: map[string]struct{}{addr.Hex(): {}},
	}

	ctx := walletCtx(t, key)
	allowed, _, err := w.Allowed(ctx, "docker.io", "hello-world", "")
	assert.NoError(t, err)
	assert.True(t, allowed)

	w.superusers = map[string]struct{}{}
	allowed, _, err = w.Allowed(ctx, "docker.io", "hello-world", "")
	assert.False(t, allowed)
}

func TestWhitelistAllowed(t *testing.T) {
	data := `
{
  "docker.io/sonm/eth-claymore": {
    "allowed_hashes": [
      "sha256:b5f9a9e47fa319607ed339789ef6692d4937ae5910b86e0ab929d035849e491e"
    ]
  }
}`

	ctx := walletCtx(t, key)
	reader := strings.NewReader(data)
	w := whitelist{}
	w.fillFromJsonReader(ctx, reader)
	allowed, _, err := w.Allowed(ctx, "", "sonm/eth-claymore@sha256:b5f9a9e47fa319607ed339789ef6692d4937ae5910b86e0ab929d035849e491e", "")
	assert.True(t, allowed)
	assert.NoError(t, err)
}
