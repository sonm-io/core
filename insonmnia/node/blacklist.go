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
	return &blacklistAPI{remotes: opts}
}

func (b *blacklistAPI) List(ctx context.Context, addr *sonm.EthAddress) (*sonm.BlacklistReply, error) {
	if addr.IsZero() {
		return nil, errors.New("cannot get blacklist for empty address")
	}

	req := &sonm.BlacklistRequest{UserID: addr}
	return b.remotes.dwh.GetBlacklist(ctx, req)
}

func (b *blacklistAPI) Remove(ctx context.Context, addr *sonm.EthAddress) (*sonm.Empty, error) {
	if err := b.remotes.eth.Blacklist().Remove(ctx, b.remotes.key, addr.Unwrap()); err != nil {
		return nil, fmt.Errorf("cannot remove address from blacklist: %v", err)
	}

	return &sonm.Empty{}, nil
}
