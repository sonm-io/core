package node

import (
	"context"
	"errors"
	"fmt"

	"github.com/sonm-io/core/proto"
)

type blacklistAPI struct {
	remotes *remoteOptions
}

func newBlacklistAPI(opts *remoteOptions) sonm.BlacklistServer {
	return &blacklistAPI{
		remotes: opts,
	}
}

func (m *blacklistAPI) List(ctx context.Context, addr *sonm.EthAddress) (*sonm.BlacklistReply, error) {
	if addr.IsZero() {
		return nil, errors.New("address is empty")
	}

	return m.remotes.dwh.GetBlacklist(ctx, &sonm.BlacklistRequest{UserID: addr})
}

func (m *blacklistAPI) Remove(ctx context.Context, addr *sonm.EthAddress) (*sonm.Empty, error) {
	if err := m.remotes.eth.Blacklist().Remove(ctx, m.remotes.key, addr.Unwrap()); err != nil {
		return nil, fmt.Errorf("failed to remove address from blacklist: %v", err)
	}

	return &sonm.Empty{}, nil
}
