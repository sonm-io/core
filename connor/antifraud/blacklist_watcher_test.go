package antifraud

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// avoid using the mockgen because it cannot properly mock stream method
// that are also declared in the proto/node.sonm.go
type blacklistClientMock struct{}

func (blacklistClientMock) List(ctx context.Context, in *sonm.EthAddress, opts ...grpc.CallOption) (*sonm.BlacklistReply, error) {
	return nil, nil
}
func (blacklistClientMock) Remove(ctx context.Context, in *sonm.EthAddress, opts ...grpc.CallOption) (*sonm.Empty, error) {
	return &sonm.Empty{}, nil
}

func newTestBlacklistWatcher() blacklistWatcher {
	return blacklistWatcher{
		address:     common.HexToAddress("0x950B346f1028cbf76a6ed721786eBcfb13DAc4Ec"),
		nextPeriod:  minStep,
		lastSuccess: time.Now(),
		client:      blacklistClientMock{},
		log:         zap.NewNop(),
	}
}

func TestBlackListWatcher(t *testing.T) {
	w := newTestBlacklistWatcher()

	assert.False(t, w.isBlacklisted(), "should not be blacklisted by default")
	w.Success()
	assert.False(t, w.isBlacklisted(), "should not be blacklisted after success")

	w.Failure()
	assert.True(t, w.isBlacklisted(), "should be blacklisted when first failure detected")

	w.Success()
	assert.True(t, w.isBlacklisted(), "should still be blacklisted after failure")

	// assume that un-blacklisting time is come
	w.unBlacklistAt = time.Now().Add(-1 * time.Second)
	err := w.TryUnblacklist(context.Background())
	require.NoError(t, err)
	assert.False(t, w.isBlacklisted(), "should not be blacklisted after removal")
}
