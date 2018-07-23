package connor

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/connor_v2/price"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
)

type Connor struct {
	cfg *Config
	log *zap.Logger
	key *ecdsa.PrivateKey

	snmPriceProvider   price.Provider
	tokenPriceProvider price.Provider
	orderManager       *engine
	marketClient       sonm.MarketClient
	dealsClient        sonm.DealManagementClient
}

func New(ctx context.Context, cfg *Config, log *zap.Logger) (*Connor, error) {
	log.Debug("building new instance", zap.Any("config", *cfg))

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return nil, fmt.Errorf("cannot load eth keys: %v", err)
	}

	_, TLSConfig, err := util.NewHitlessCertRotator(context.Background(), key)
	if err != nil {
		return nil, fmt.Errorf("cannot create cert TLS config: %v", err)
	}

	creds := auth.NewWalletAuthenticator(util.NewTLS(TLSConfig), crypto.PubkeyToAddress(key.PublicKey))
	cc, err := xgrpc.NewClient(ctx, cfg.Node.Endpoint.String(), creds)
	if err != nil {
		return nil, fmt.Errorf("cannot create connection to node: %v", err)
	}

	marketClient := sonm.NewMarketClient(cc)
	dealsClient := sonm.NewDealManagementClient(cc)

	return &Connor{
		key:                key,
		cfg:                cfg,
		log:                log,
		marketClient:       marketClient,
		dealsClient:        dealsClient,
		snmPriceProvider:   price.NewSonmPriceProvider(),
		tokenPriceProvider: price.NewProvider(cfg.Mining.Token),
		orderManager:       NewEngine(ctx, cfg.Engine, log, marketClient, dealsClient),
	}, nil
}

func (c *Connor) Serve(ctx context.Context) error {
	c.log.Debug("starting Connor instance")

	if err := c.loadInitialData(ctx); err != nil {
		return fmt.Errorf("initializind failed: %v", err)
	}

	c.log.Debug("price", zap.String("SNM", c.snmPriceProvider.GetPrice().String()),
		zap.String(c.cfg.Mining.Token, c.tokenPriceProvider.GetPrice().String()))

	c.orderManager.start(ctx)

	// restore two subsets of orders, then separate on non-exiting orders that
	// should be placed on market and active orders that should be watched
	// for deal opening.
	exitingOrders, err := c.marketClient.GetOrders(ctx, &sonm.Count{Count: 1000})
	if err != nil {
		return fmt.Errorf("cannot load orders from market: %v", err)
	}

	exitingDeals, err := c.dealsClient.List(ctx, &sonm.Count{Count: 1000})
	if err != nil {
		return fmt.Errorf("cannot load deals from market: %v", err)
	}

	exitingCorders := NewCordersSlice(exitingOrders.GetOrders(), c.cfg.Mining.Token)
	targetCorders := c.getTargetCorders()

	set := c.divideOrdersSets(exitingCorders, targetCorders)
	c.log.Debug("restoring exiting entities",
		zap.Int("orders_restore", len(set.toRestore)),
		zap.Int("orders_create", len(set.toCreate)),
		zap.Int("deals_restore", len(exitingDeals.GetDeal())))

	for _, ord := range set.toCreate {
		c.orderManager.CreateOrder(ord)
	}

	for _, ord := range set.toRestore {
		c.orderManager.RestoreOrder(ord)
	}

	for _, deal := range exitingDeals.GetDeal() {
		c.orderManager.RestoreDeal(deal)
	}

	<-ctx.Done()
	c.close()
	return nil
}

func (c *Connor) loadInitialData(ctx context.Context) error {
	if err := c.snmPriceProvider.Update(ctx); err != nil {
		return fmt.Errorf("cannot update SNM price: %v", err)
	}

	if err := c.tokenPriceProvider.Update(ctx); err != nil {
		return fmt.Errorf("cannot update %s price: %v", c.cfg.Mining.Token, err)
	}

	return nil
}

type ordersSets struct {
	toCreate  []*Corder
	toRestore []*Corder
}

func (c *Connor) divideOrdersSets(exitingCorders, targetCorders []*Corder) *ordersSets {
	byHashrate := map[uint64]*Corder{}
	for _, ord := range exitingCorders {
		byHashrate[ord.GetHashrate()] = ord
	}

	set := &ordersSets{
		toCreate:  make([]*Corder, 0),
		toRestore: make([]*Corder, 0),
	}

	for _, ord := range targetCorders {
		if ex, ok := byHashrate[ord.GetHashrate()]; ok {
			set.toRestore = append(set.toRestore, ex)
		} else {
			set.toCreate = append(set.toCreate, ord)
		}
	}

	return set
}

func (c *Connor) getTargetCorders() []*Corder {
	v := make([]*Corder, 0)

	for hashrate := c.cfg.Market.FromHashRate; hashrate <= c.cfg.Market.ToHashRate; hashrate += c.cfg.Market.Step {
		bigHashrate := big.NewInt(int64(hashrate))
		p := big.NewInt(0).Mul(bigHashrate, c.tokenPriceProvider.GetPrice())
		order, _ := NewCorderFromParams(c.cfg.Mining.Token, p, hashrate)
		v = append(v, order)
	}

	return v
}

func (c *Connor) close() {
	c.log.Debug("closing Connor")
}
