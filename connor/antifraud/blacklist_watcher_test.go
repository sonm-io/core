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
// that are also declared in the proto/node.pb.go
type blacklistClientMock struct{}

func (blacklistClientMock) List(ctx context.Context, in *sonm.EthAddress, opts ...grpc.CallOption) (*sonm.BlacklistReply, error) {
	return nil, nil
}
func (blacklistClientMock) Remove(ctx context.Context, in *sonm.EthAddress, opts ...grpc.CallOption) (*sonm.Empty, error) {
	return &sonm.Empty{}, nil
}

func TestBlackListWatcher(t *testing.T) {

	w := blacklistWatcher{
		address:     common.HexToAddress("0x950B346f1028cbf76a6ed721786eBcfb13DAc4Ec"),
		currentStep: minStep,
		client:      blacklistClientMock{},
		log:         zap.NewNop(),
	}

	assert.False(t, w.Blacklisted(), "should not be blacklisted by default")
	w.Success()
	assert.False(t, w.Blacklisted(), "should not be blacklisted after success")

	w.Failure()
	assert.True(t, w.Blacklisted(), "should be blacklisted when first failure detected")

	w.Success()
	assert.True(t, w.Blacklisted(), "should still be blacklisted after failure")

	// assume that un-blacklisting time is come
	w.till = time.Now().Add(-1 * time.Second)
	err := w.TryUnblacklist(context.Background())
	require.NoError(t, err)
	assert.False(t, w.Blacklisted(), "should not be blacklisted after removal")
}
