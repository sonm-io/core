package node

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
)

type tokenAPI struct {
	remotes *remoteOptions
}

func (t *tokenAPI) TestTokens(ctx context.Context, _ *sonm.Empty) (*sonm.Empty, error) {
	if _, err := t.remotes.eth.TestToken().GetTokens(ctx, t.remotes.key); err != nil {
		return nil, err
	}

	return &sonm.Empty{}, nil
}

func (t *tokenAPI) Balance(ctx context.Context, _ *sonm.Empty) (*sonm.BalanceReply, error) {
	addr := crypto.PubkeyToAddress(t.remotes.key.PublicKey)

	live, err := t.remotes.eth.MasterchainToken().BalanceOf(ctx, addr)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get live token balance")
	}

	side, err := t.remotes.eth.SidechainToken().BalanceOf(ctx, addr)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get side token balance")
	}

	return &sonm.BalanceReply{
		LiveBalance: sonm.NewBigInt(live),
		SideBalance: sonm.NewBigInt(side),
	}, nil
}

func newTokenManagementAPI(opts *remoteOptions) sonm.TokenManagementServer {
	return &tokenAPI{remotes: opts}
}
