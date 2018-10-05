package optimus

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

// Blacklist is a thing that can be asked to determine whether a specific ETH
// address is in the "owner" blacklist and vise-verse.
// This is used to be sure that the order created via Optimus can be bought
// by the user this order was created for.
type blacklist struct {
	owner     common.Address
	blacklist map[common.Address]struct{}
	dwh       sonm.DWHClient
	log       *zap.SugaredLogger
}

func newBlacklist(owner common.Address, dwh sonm.DWHClient, log *zap.SugaredLogger) *blacklist {
	return &blacklist{
		owner:     owner,
		blacklist: map[common.Address]struct{}{},
		dwh:       dwh,
		log:       log.With(zap.String("addr", owner.Hex())),
	}
}

// IsAllowed checks whether the given "addr" is allowed for this blacklist.
// This method returns true both if an "owner" is in the "addr" blacklist
// and vice-versa.
// The blacklist needs to be updated before calling this method.
func (m *blacklist) IsAllowed(addr common.Address) bool {
	_, ok := m.blacklist[addr]
	return !ok
}

func (m *blacklist) Update(ctx context.Context) error {
	m.log.Debug("updating blacklist")

	blacklist, err := m.dwh.GetBlacklistsContainingUser(ctx, &sonm.BlacklistRequest{
		UserID: sonm.NewEthAddress(m.owner),
	})
	if err != nil {
		return err
	}

	m.blacklist = map[common.Address]struct{}{}
	for _, addr := range blacklist.Blacklists {
		m.blacklist[addr.Unwrap()] = struct{}{}
	}

	m.log.Infow("blacklist has been updated", zap.Any("blacklist", m.blacklist))

	return nil
}

type multiBlacklist struct {
	blacklists []*blacklist
}

func newMultiBlacklist(blacklists ...*blacklist) *multiBlacklist {
	return &multiBlacklist{
		blacklists: blacklists,
	}
}

func (m *multiBlacklist) IsAllowed(addr common.Address) bool {
	for _, blacklist := range m.blacklists {
		if !blacklist.IsAllowed(addr) {
			return false
		}
	}

	return true
}

func (m *multiBlacklist) Update(ctx context.Context) error {
	for _, blacklist := range m.blacklists {
		if err := blacklist.Update(ctx); err != nil {
			return err
		}
	}

	return nil
}

type emptyBlacklist struct{}

func newEmptyBlacklist() *emptyBlacklist {
	return &emptyBlacklist{}
}

func (emptyBlacklist) Update(ctx context.Context) error {
	return nil
}

func (emptyBlacklist) IsAllowed(addr common.Address) bool {
	return true
}
