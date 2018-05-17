package miner

import (
	"context"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func walletCtx(addr common.Address) context.Context {
	peerCtx := peer.NewContext(context.Background(), &peer.Peer{
		AuthInfo: auth.EthAuthInfo{TLS: credentials.TLSInfo{}, Wallet: addr},
	})

	ctx := metadata.NewIncomingContext(peerCtx, metadata.New(map[string]string{}))
	return ctx
}

func TestWhitelistSuperuser(t *testing.T) {
	w := whitelist{
		superusers: map[string]struct{}{addr.Hex(): {}},
	}

	ctx := walletCtx(addr)
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

	ctx := walletCtx(addr)
	reader := strings.NewReader(data)
	w := whitelist{}
	w.fillFromJsonReader(ctx, reader)
	allowed, _, err := w.Allowed(ctx, "", "sonm/eth-claymore@sha256:b5f9a9e47fa319607ed339789ef6692d4937ae5910b86e0ab929d035849e491e", "")
	assert.True(t, allowed)
	assert.NoError(t, err)
}
