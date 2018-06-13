package node

import (
	"context"

	"github.com/pkg/errors"
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
		return nil, errors.WithMessage(err, "cannot remove address from blacklist")
	}

	return &sonm.Empty{}, nil
}
