package main

import (
	"context"
	"crypto/ecdsa"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

var (
	privateKeyGlobal         *ecdsa.PrivateKey
	privateKeyOnce           sync.Once
	transportCredentials     credentials.TransportCredentials
	transportCredentialsOnce sync.Once
)

func PrivateKey(cfg EthereumConfig) *ecdsa.PrivateKey {
	if cfg.AccountType == "random" {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			panic(err)
		}
		return privateKey
	}

	privateKeyOnce.Do(func() {
		ethConfig := accounts.EthConfig{
			Keystore:   cfg.AccountPath,
			Passphrase: cfg.AccountPass,
		}

		privateKey, err := ethConfig.LoadKey()
		if err != nil {
			panic(err)
		}

		privateKeyGlobal = privateKey
	})

	return privateKeyGlobal
}

func TransportCredentials(privateKey *ecdsa.PrivateKey) credentials.TransportCredentials {
	transportCredentialsOnce.Do(func() {
		_, TLSConfig, err := util.NewHitlessCertRotator(context.Background(), privateKey)
		if err != nil {
			panic(err)
		}

		transportCredentials = util.NewTLS(TLSConfig)
	})

	return transportCredentials
}

func NewLogger(cfg LoggingConfig) (*zap.Logger, error) {
	level, err := logging.NewLevelFromString(cfg.Level)
	if err != nil {
		return nil, err
	}

	return logging.BuildLogger(logging.Config{
		Level:  level,
		Output: logging.StdoutLogOutput,
	})
}
