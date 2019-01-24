package migration

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/dwh"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/multierror"
	"github.com/sonm-io/core/util/xgrpc"
	"google.golang.org/grpc/credentials"
)

type MarketShredder interface {
	WeDontNeedNoWaterLetTheMothefuckerBurn(ctx context.Context, pKey *ecdsa.PrivateKey) error
}

type nilMarketShredder struct {
}

func (m *nilMarketShredder) WeDontNeedNoWaterLetTheMothefuckerBurn(ctx context.Context, pKey *ecdsa.PrivateKey) error {
	return nil
}

type marketShredder struct {
	api blockchain.API
	dwh sonm.DWHClient
}

type MigrationConfig struct {
	Blockchain   *blockchain.Config
	MigrationDWH *dwh.YAMLConfig
	Enabled      *bool `default:"true"`
	Version      uint
}

func NewV1MarketShredder(ctx context.Context, cfg *MigrationConfig, credentials credentials.TransportCredentials) (MarketShredder, error) {
	if !*cfg.Enabled {
		return &nilMarketShredder{}, nil
	}
	if cfg.MigrationDWH == nil {
		return nil, fmt.Errorf("dwh config is required for enabled migrartion")
	}
	api, err := blockchain.NewAPI(ctx, blockchain.WithConfig(cfg.Blockchain), blockchain.WithVersion(1))
	if err != nil {
		return nil, err
	}

	cc, err := xgrpc.NewClient(ctx, cfg.MigrationDWH.Endpoint, credentials)
	if err != nil {
		return nil, err
	}

	dwh := sonm.NewDWHClient(cc)
	return NewMarketShredder(api, dwh), nil
}

func NewMarketShredder(api blockchain.API, dwh sonm.DWHClient) *marketShredder {
	return &marketShredder{
		api: api,
		dwh: dwh,
	}
}

func (m *marketShredder) WeDontNeedNoWaterLetTheMothefuckerBurn(ctx context.Context, pKey *ecdsa.PrivateKey) error {
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
