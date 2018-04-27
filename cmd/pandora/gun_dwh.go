package main

import (
	"context"
	"crypto/ecdsa"

	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
)

type DWHExtConfig struct {
	Logging     LoggingConfig  `config:"logging"`
	Ethereum    EthereumConfig `config:"ethereum"`
	DWHEndpoint string         `config:"dwh_endpoint"`
}

type dwhExt struct {
	privateKey *ecdsa.PrivateKey
	dwh        sonm.DWHClient
	log        *zap.Logger
}

func NewDWHGun(cfg DWHExtConfig) (Gun, error) {
	privateKey := PrivateKey(cfg.Ethereum)

	log, err := NewLogger(cfg.Logging)
	if err != nil {
		return nil, err
	}

	credentials := TransportCredentials(privateKey)

	conn, err := xgrpc.NewClient(context.Background(), cfg.DWHEndpoint, credentials)
	if err != nil {
		return nil, err
	}
	dwh := sonm.NewDWHClient(conn)

	ext := &dwhExt{
		privateKey: privateKey,
		dwh:        dwh,
		log:        log,
	}
	return newGun(ext, log), nil
}
