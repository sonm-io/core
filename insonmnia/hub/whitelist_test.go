package hub

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func TestWhitelistSuperuser(t *testing.T) {
	w := whitelist{
		superusers: map[string]struct{}{"0x42": struct{}{}},
	}

	ctx := peer.NewContext(context.Background(), &peer.Peer{AuthInfo: util.EthAuthInfo{Wallet: common.HexToAddress("0x42")}})
	ctx = metadata.NewIncomingContext(ctx, metadata.New(map[string]string{"wallet": "0x42"}))
	allowed, _, err := w.Allowed(ctx, "docker.io", "hello-world", "")
	assert.NoError(t, err)
	assert.True(t, allowed)

	w.superusers = map[string]struct{}{}
	allowed, _, err = w.Allowed(ctx, "docker.io", "hello-world", "")
	assert.False(t, allowed)
}
