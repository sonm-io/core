package node

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
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

func (t *tokenAPI) Deposit(ctx context.Context, amount *sonm.BigInt) (*sonm.Empty, error) {
	if err := t.remotes.eth.MasterchainToken().ApproveAtLeast(ctx, t.remotes.key, blockchain.GatekeeperMasterchainAddr(), amount.Unwrap()); err != nil {
		return nil, fmt.Errorf("cannot change allowance: %s", err)
	}

	if err := t.remotes.eth.MasterchainGate().PayIn(ctx, t.remotes.key, amount.Unwrap()); err != nil {
		return nil, err
	}

	return &sonm.Empty{}, nil
}

func (t *tokenAPI) Withdraw(ctx context.Context, amount *sonm.BigInt) (*sonm.Empty, error) {
	if err := t.remotes.eth.SidechainToken().IncreaseApproval(ctx, t.remotes.key, blockchain.GatekeeperSidechainAddr(), amount.Unwrap()); err != nil {
		return nil, fmt.Errorf("cannot change allowance: %s", err)
	}

	if err := t.remotes.eth.SidechainGate().PayIn(ctx, t.remotes.key, amount.Unwrap()); err != nil {
		return nil, err
	}

	return &sonm.Empty{}, nil
}

func newTokenManagementAPI(opts *remoteOptions) sonm.TokenManagementServer {
	return &tokenAPI{remotes: opts}
}
