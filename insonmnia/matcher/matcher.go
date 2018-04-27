package matcher

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

type Matcher interface {
	CreateDealByOrder(ctx context.Context, order *sonm.Order) (*sonm.Deal, error)
}

type Config struct {
	Key        *ecdsa.PrivateKey
	PollDelay  time.Duration
	DWH        sonm.DWHClient
	Eth        blockchain.API
	QueryLimit uint64
}

type matcher struct {
	cfg *Config
}

func NewMatcher(cfg *Config) Matcher {
	return &matcher{cfg: cfg}
}

func (m *matcher) CreateDealByOrder(ctx context.Context, order *sonm.Order) (*sonm.Deal, error) {
	tk := util.NewImmediateTicker(m.cfg.PollDelay)
	defer tk.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-tk.C:
			// 1. check if order is actual
			ctxlog.G(ctx).Info("check that order exists", zap.String("id", order.GetId().Unwrap().String()))
			if err := m.checkIfOrderExists(ctx, order.GetId().Unwrap()); err != nil {
				return nil, err
			}

			// 2. get matching orders from dwh
			ctxlog.G(ctx).Info("search for matching order via DWH", zap.String("id", order.GetId().Unwrap().String()))
			matchingOrders, err := m.getMatchingOrders(ctx, order.GetId().Unwrap())
			if err != nil {
				return nil, err
			}

			// 3. iterate over sorted orders
			for _, dealWithMe := range matchingOrders {
				fmt.Printf(" >>>>>>> %v | %v\r\n", order, dealWithMe)

				bid, ask, err := m.reorderOrders(order, dealWithMe)
				if err != nil {
					return nil, err
				}

				// 4. try to open deal
				fmt.Printf(" >>>>>>> %v | %v\r\n", bid, ask)
				ctxlog.G(ctx).Info("opening deal", zap.String("bid", bid.GetId().Unwrap().String()), zap.String("ask", ask.GetId().Unwrap().String()))
				deal, err := m.openDeal(ctx, bid, ask)
				if err == nil {
					return deal, nil
				}

				// 5. if deal is not created - wait for timeout and goto 1
				ctxlog.G(ctx).Warn("cannot open deal",
					zap.Error(err),
					zap.String("bid", bid.GetId().Unwrap().String()),
					zap.String("ask", ask.GetId().Unwrap().String()))
			}
		}
	}
}

func (m *matcher) checkIfOrderExists(ctx context.Context, id *big.Int) error {
	order, err := m.cfg.Eth.GetOrderInfo(ctx, id)
	if err != nil {
		return err
	}

	if order.GetOrderStatus() != sonm.OrderStatus_ORDER_ACTIVE {
		return errors.New("order is not active")
	}

	return nil
}

func (m *matcher) getMatchingOrders(ctx context.Context, id *big.Int) ([]*sonm.Order, error) {
	dwhReply, err := m.cfg.DWH.GetMatchingOrders(ctx, &sonm.MatchingOrdersRequest{
		Id:    sonm.NewBigInt(id),
		Limit: m.cfg.QueryLimit,
	})

	if err != nil {
		return nil, err
	}

	orders := make([]*sonm.Order, len(dwhReply.GetOrders()))
	for i, ord := range dwhReply.GetOrders() {
		fmt.Printf(" >>>>>>. %#v\r\n", ord)
		orders[i] = ord.GetOrder()
	}

	return orders, nil
}

func (m *matcher) openDeal(ctx context.Context, bid, ask *sonm.Order) (*sonm.Deal, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-m.cfg.Eth.OpenDeal(ctx, m.cfg.Key, ask.GetId().Unwrap(), bid.GetId().Unwrap()):
		return res.Deal, res.Err
	}
}

func (m *matcher) reorderOrders(one, two *sonm.Order) (bid, ask *sonm.Order, err error) {
	// just a sanity check, orders must have different types
	if one.GetOrderType() == two.GetOrderType() {
		return nil, nil, errors.New("orders must have different types")
	}

	if one.GetOrderType() == sonm.OrderType_BID {
		bid, ask = one, two
	} else {
		ask, bid = one, two
	}

	return
}
