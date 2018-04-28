package main

import (
	"crypto/ecdsa"

	"github.com/sonm-io/core/blockchain"
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

	market, err := blockchain.NewAPI(blockchain.WithSidechainEndpoint(cfg.Ethereum.Endpoint), blockchain.WithGasPrice(0))
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
