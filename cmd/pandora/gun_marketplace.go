package main

import (
	"context"
	"crypto/ecdsa"

	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

type MarketplaceExtConfig struct {
	Logging  LoggingConfig  `config:"logging"`
	Ethereum EthereumConfig `config:"ethereum"`
}

type marketplaceExt struct {
	privateKey *ecdsa.PrivateKey
	market     blockchain.MarketAPI
	log        *zap.Logger
}

func NewMarketplaceGun(cfg MarketplaceExtConfig) (Gun, error) {
	privateKey := PrivateKey(cfg.Ethereum)

	registryAddr, err := util.HexToAddress(cfg.Ethereum.Registry)
	if err != nil {
		return nil, err
	}

	market, err := blockchain.NewAPI(context.Background(),
		blockchain.WithSidechainEndpoint(cfg.Ethereum.Endpoint),
		blockchain.WithContractRegistry(registryAddr))
	if err != nil {
		return nil, err
	}

	log, err := NewLogger(cfg.Logging)
	if err != nil {
		return nil, err
	}

	ext := &marketplaceExt{
		privateKey: privateKey,
		market:     market.Market(),
		log:        log,
	}
	return newGun(ext, log), nil
}
