package migration

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/dwh"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/multierror"
	"github.com/sonm-io/core/util/xgrpc"
)

type MarketShredder struct {
	api blockchain.API
	dwh sonm.DWHClient
}

func NewV1MarketShredder(ctx context.Context, bcCfg *blockchain.Config, dwhCfg dwh.YAMLConfig) (*MarketShredder, error) {
	api, err := blockchain.NewAPI(ctx, blockchain.WithConfig(bcCfg), blockchain.WithVersion(1))
	if err != nil {
		return nil, err
	}

	cc, err := xgrpc.NewClient(ctx, dwhCfg.Endpoint, credentials)
	if err != nil {
		return err
	}

	dwh := sonm.NewDWHClient(cc)
}

func NewMarketShredder(api blockchain.API, dwh sonm.DWHClient) *MarketShredder {
	return &MarketShredder{
		api: api,
		dwh: dwh,
	}
}

func (m *MarketShredder) WeDontNeedNoWaterLetTheMothefuckerBurn(ctx context.Context, pKey *ecdsa.PrivateKey) error {
	author := crypto.PubkeyToAddress(pKey.PublicKey)
	ordersRequest := &sonm.OrdersRequest{
		AuthorID: sonm.NewEthAddress(author),
	}
	orders, err := m.dwh.GetOrders(ctx, ordersRequest)
	if err != nil {
		return err
	}
	market := m.api.Market()
	merr := multierror.NewMultiError()
	for _, order := range orders.GetOrders() {
		if err := market.CancelOrder(ctx, pKey, order.GetOrder().GetId().Unwrap()); err != nil {
			merr = multierror.Append(merr, err)
		}
	}

	dealsRequest := &sonm.DealsRequest{
		AnyUserID: sonm.NewEthAddress(author),
	}
	deals, err := m.dwh.GetDeals(ctx, dealsRequest)
	if err != nil {
		return err
	}

	for _, deal := range deals.GetDeals() {
		// We only match masterID, but master cannot close deals.
		if deal.GetDeal().GetConsumerID().Unwrap() != author && deal.GetDeal().GetSupplierID().Unwrap() != author {
			continue
		}
		if err := market.CloseDeal(ctx, pKey, deal.GetDeal().GetId().Unwrap(), sonm.BlacklistType_BLACKLIST_NOBODY); err != nil {
			merr = multierror.Append(merr, err)
		}
	}
	return merr.ErrorOrNil()
}
