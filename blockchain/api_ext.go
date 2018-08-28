package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
	"golang.org/x/sync/errgroup"
)

type niceMarketAPI struct {
	MarketAPI
	profiles  ProfileRegistryAPI
	blacklist BlacklistAPI
}

func (m *niceMarketAPI) OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bidID *big.Int) (*sonm.Deal, error) {
	orders, err := m.GetOrderInfoMulti(ctx, key, askID, bidID)
	if err != nil {
		return nil, err
	}
	ask, bid := orders[0], orders[1]

	if ask.GetOrderType() != sonm.OrderType_ASK {
		return nil, fmt.Errorf("ask must have ASK type, but it is %s", ask.GetOrderType())
	}
	if bid.GetOrderType() != sonm.OrderType_BID {
		return nil, fmt.Errorf("bid must have BID type, but it is %s", bid.GetOrderType())
	}
	if ask.GetOrderStatus() != sonm.OrderStatus_ORDER_ACTIVE {
		return nil, fmt.Errorf("ask order must be active, but it is %s", ask.GetOrderStatus())
	}
	if bid.GetOrderStatus() != sonm.OrderStatus_ORDER_ACTIVE {
		return nil, fmt.Errorf("bid order must be active, but it is %s", bid.GetOrderStatus())
	}
	if ask.GetPrice().Cmp(bid.GetPrice()) > 0 {
		return nil, fmt.Errorf("bid price %s must be >= ask price %s", bid.GetPrice(), ask.GetPrice())
	}
	if ask.GetDuration() < bid.GetDuration() {
		return nil, fmt.Errorf("bid duration %d must be <= ask duration %d", bid.GetDuration(), ask.GetDuration())
	}
	if !ask.GetNetflags().ConverseImplication(bid.GetNetflags()) {
		return nil, fmt.Errorf("bid netflags %v must fit in ask netflags %v", bid.GetNetflags().ToBoolSlice(), ask.GetNetflags().ToBoolSlice())
	}
	numBenchmarks, err := m.GetNumBenchmarks(ctx)
	if err != nil {
		return nil, err
	}
	for id := 0; id < int(numBenchmarks); id++ {
		if bid.GetBenchmarks().Get(id) > ask.GetBenchmarks().Get(id) {
			return nil, fmt.Errorf("benchmark matching failed: id=%d bid benchmark %d must be <= ask benchmark %d", id, bid.GetBenchmarks().Get(id), ask.GetBenchmarks().Get(id))
		}
	}

	masters, err := m.GetMasterMulti(ctx, ask, bid)
	if err != nil {
		return nil, err
	}
	askMaster, bidMaster := masters[0], masters[1]

	askCounterparty := ask.GetCounterpartyID().Unwrap()
	if (askCounterparty != common.Address{} && askCounterparty != bidMaster) {
		return nil, fmt.Errorf("ask counterparty %s doesn't match with bid master %s", askCounterparty.Hex(), bidMaster.Hex())
	}
	bidCounterparty := bid.GetCounterpartyID().Unwrap()
	if (bidCounterparty != common.Address{} && bidCounterparty != askMaster) {
		return nil, fmt.Errorf("bid counterparty %s doesn't match with ask master %s", bidCounterparty.Hex(), askMaster.Hex())
	}

	identities, err := m.GetProfileLevelMulti(ctx, ask.GetAuthorID().Unwrap(), bid.GetAuthorID().Unwrap())
	if err != nil {
		return nil, err
	}
	askIdentity, bidIdentity := identities[0], identities[1]

	// TODO: Here the market bug, but we need to live with it #1293.
	if bidIdentity < ask.GetIdentityLevel() {
		return nil, fmt.Errorf("bid identity %s must be >= ask author identity %s", bidIdentity.String(), ask.GetIdentityLevel().String())
	}
	if askIdentity < bid.GetIdentityLevel() {
		return nil, fmt.Errorf("ask identity %s must be >= bid author identity %s", askIdentity.String(), bid.GetIdentityLevel().String())
	}

	blacklist := blacklistVerify{}

	wg, blacklistCtx := errgroup.WithContext(ctx)
	wg.Go(func() (err error) {
		blacklist.AskMasterInBidBlacklist, err = m.blacklist.Check(blacklistCtx, common.HexToAddress(bid.GetBlacklist()), askMaster)
		return err
	})
	wg.Go(func() (err error) {
		blacklist.AskAuthorInBidBlacklist, err = m.blacklist.Check(blacklistCtx, common.HexToAddress(bid.GetBlacklist()), ask.GetAuthorID().Unwrap())
		return err
	})
	wg.Go(func() (err error) {
		blacklist.AskMasterInBidAuthorBlacklist, err = m.blacklist.Check(blacklistCtx, bid.GetAuthorID().Unwrap(), askMaster)
		return err
	})
	wg.Go(func() (err error) {
		blacklist.AskAuthorInBidAuthorBlacklist, err = m.blacklist.Check(blacklistCtx, bid.GetAuthorID().Unwrap(), ask.GetAuthorID().Unwrap())
		return err
	})
	wg.Go(func() (err error) {
		blacklist.BidAuthorInAskBlacklist, err = m.blacklist.Check(blacklistCtx, common.HexToAddress(ask.GetBlacklist()), bid.GetAuthorID().Unwrap())
		return err
	})
	wg.Go(func() (err error) {
		blacklist.BidAuthorInAskMasterBlacklist, err = m.blacklist.Check(blacklistCtx, askMaster, bid.GetAuthorID().Unwrap())
		return err
	})
	wg.Go(func() (err error) {
		blacklist.BidAuthorInAskAuthorBlacklist, err = m.blacklist.Check(blacklistCtx, ask.GetAuthorID().Unwrap(), bid.GetAuthorID().Unwrap())
		return err
	})
	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("blacklist check failed: %v", err)
	}
	if err := blacklist.Verify(bid, ask, bidMaster, askMaster); err != nil {
		return nil, fmt.Errorf("blacklist check failed: %v", err)
	}

	return m.MarketAPI.OpenDeal(ctx, key, askID, bidID)
}

func (m *niceMarketAPI) GetOrderInfoMulti(ctx context.Context, key *ecdsa.PrivateKey, orderIDs ...*big.Int) ([]*sonm.Order, error) {
	wg, ctx := errgroup.WithContext(ctx)

	orders := make([]*sonm.Order, len(orderIDs))
	for id := range orderIDs {
		id := id
		wg.Go(func() error {
			v, err := m.GetOrderInfo(ctx, orderIDs[id])
			if err != nil {
				return err
			}

			orders[id] = v

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (m *niceMarketAPI) GetMasterMulti(ctx context.Context, orders ...*sonm.Order) ([]common.Address, error) {
	wg, ctx := errgroup.WithContext(ctx)

	masters := make([]common.Address, len(orders))
	for id := range orders {
		id := id
		wg.Go(func() error {
			v, err := m.GetMaster(ctx, orders[id].GetAuthorID().Unwrap())
			if err != nil {
				return err
			}

			masters[id] = v

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return masters, nil
}

func (m *niceMarketAPI) GetProfileLevelMulti(ctx context.Context, authorIDs ...common.Address) ([]sonm.IdentityLevel, error) {
	wg, ctx := errgroup.WithContext(ctx)

	identities := make([]sonm.IdentityLevel, len(authorIDs))
	for id := range authorIDs {
		id := id
		wg.Go(func() error {
			v, err := m.profiles.GetProfileLevel(ctx, authorIDs[id])
			if err != nil {
				return err
			}

			identities[id] = v

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return identities, nil
}

type blacklistVerify struct {
	AskMasterInBidBlacklist       bool
	AskAuthorInBidBlacklist       bool
	AskMasterInBidAuthorBlacklist bool
	AskAuthorInBidAuthorBlacklist bool
	BidAuthorInAskBlacklist       bool
	BidAuthorInAskMasterBlacklist bool
	BidAuthorInAskAuthorBlacklist bool
}

func (m *blacklistVerify) Verify(bid, ask *sonm.Order, bidMaster, askMaster common.Address) error {
	if m.AskMasterInBidBlacklist {
		return fmt.Errorf("ASK master %s is in BID blacklist %s", askMaster.Hex(), common.HexToAddress(bid.GetBlacklist()).Hex())
	}
	if m.AskAuthorInBidBlacklist {
		return fmt.Errorf("ASK author %s in in BID blacklist %s", ask.GetAuthorID().Unwrap().Hex(), common.HexToAddress(bid.GetBlacklist()).Hex())
	}
	if m.AskMasterInBidAuthorBlacklist {
		return fmt.Errorf("ASK master %s is in BID author blacklist %s", askMaster.Hex(), bid.GetAuthorID().Unwrap().Hex())
	}
	if m.AskAuthorInBidAuthorBlacklist {
		return fmt.Errorf("ASK author %s is in BID author blacklist %s", ask.GetAuthorID().Unwrap(), bid.GetAuthorID().Unwrap().Hex())
	}
	if m.BidAuthorInAskBlacklist {
		return fmt.Errorf("BID author %s is in ASK blacklist %s", bid.GetAuthorID().Unwrap().Hex(), common.HexToAddress(ask.GetBlacklist()).Hex())
	}
	if m.BidAuthorInAskMasterBlacklist {
		return fmt.Errorf("BID author %s is in ASK master blacklist %s", bid.GetAuthorID().Unwrap().Hex(), askMaster.Hex())
	}
	if m.BidAuthorInAskAuthorBlacklist {
		return fmt.Errorf("BID author %s is in ASK author blacklist %s", bid.GetAuthorID().Unwrap().Hex(), ask.GetAuthorID().Unwrap().Hex())
	}

	return nil
}
