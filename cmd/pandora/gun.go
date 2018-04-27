package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/proto"
	"github.com/yandex/pandora/core"
	"github.com/yandex/pandora/core/aggregate/netsample"
	"go.uber.org/zap"
)

type GunConfig struct {
	EthereumEndpoint    string
	EthereumAccountPath string
	EthereumAccountPass string
	LoggingLevel        string
}

func NewDefaultGunConfig() GunConfig {
	return GunConfig{}
}

type Gun struct {
	aggregator core.Aggregator

	counter    int
	privateKey *ecdsa.PrivateKey
	market     blockchain.MarketAPI
	log        *zap.Logger
}

func NewGun(cfg GunConfig) (*Gun, error) {
	ethConfig := accounts.EthConfig{
		Keystore:   cfg.EthereumAccountPath,
		Passphrase: cfg.EthereumAccountPass,
	}

	market, err := blockchain.NewAPI(blockchain.WithEthEndpoint(cfg.EthereumEndpoint), blockchain.WithGasPrice(0))
	if err != nil {
		return nil, err
	}

	level, err := logging.NewLevelFromString(cfg.LoggingLevel)
	if err != nil {
		return nil, err
	}

	log := logging.BuildLogger(*level)

	m := &Gun{
		privateKey: PrivateKey(ethConfig),
		market:     market,
		log:        log,
	}

	return m, nil
}

func (m *Gun) Bind(aggregator core.Aggregator) {
	m.aggregator = aggregator
}

func (m *Gun) Shoot(ctx context.Context, ammo core.Ammo) {
	sample := netsample.Acquire("REQUEST")

	var err error
	switch ammo := ammo.(type) {
	case *OrderInfoAmmo:
		err = m.executeGetOrderInfo(ctx, ammo)
	case *OrderPlaceAmmo:
		err = m.executePlaceOrder(ctx, ammo)
	default:
		panic(fmt.Sprintf("unknown ammo type: %T", ammo))
	}

	if err == nil {
		sample.SetProtoCode(200)
	} else {
		m.log.Warn("failed to process ammo", zap.Error(err))
		sample.SetProtoCode(500)
	}

	m.aggregator.Report(sample)
}

func (m *Gun) executeGetOrderInfo(ctx context.Context, ammo *OrderInfoAmmo) error {
	order, err := m.market.GetOrderInfo(ctx, big.NewInt(ammo.OrderID))
	if err != nil {
		return err
	}

	m.log.Debug("OK", zap.String("ammo", "GetOrderInfo"), zap.Any("order", *order))

	return nil
}

func (m *Gun) executePlaceOrder(ctx context.Context, ammo *OrderPlaceAmmo) error {
	orderOrError := <-m.market.PlaceOrder(ctx, m.privateKey, m.order())
	if orderOrError.Err != nil {
		return orderOrError.Err
	}

	m.log.Debug("OK", zap.String("ammo", "PlaceOrder"), zap.Any("order", orderOrError.Order))

	return nil
}

func (m *Gun) order() *sonm.Order {
	order := &sonm.Order{
		OrderType:      sonm.OrderType_ASK,
		OrderStatus:    sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID:       sonm.NewEthAddress(crypto.PubkeyToAddress(m.privateKey.PublicKey)),
		CounterpartyID: &sonm.EthAddress{},
		Duration:       3600 + uint64(rand.Int63n(3600)),
		Price:          sonm.NewBigIntFromInt(1000 + rand.Int63n(1000)),
		Netflags:       sonm.NetflagsToUint([3]bool{true, true, (rand.Int() % 2) == 0}),
		IdentityLevel:  sonm.IdentityLevel_ANONYMOUS,
		Blacklist:      &sonm.EthAddress{},
		Tag:            []byte("00000"),
		Benchmarks: &sonm.Benchmarks{
			Values: []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
	}

	return order
}
