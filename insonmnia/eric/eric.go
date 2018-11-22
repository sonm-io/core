package eric

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/blockchain"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"math/big"
	"sync"
)

type Eric struct {
	key    *ecdsa.PrivateKey
	logger *zap.Logger
	bch    blockchain.API
	cfg    *Config

	payoutSettings map[string]struct {
		mu sync.Mutex
		blockchain.AutoPayoutSetting
	}
}

func NewEric() (*Eric, error) {
	return &Eric{}, nil
}

func (e *Eric) Start(ctx context.Context) error {
	// load configs

	events, err := e.bch.Events().GetEvents(ctx, e.bch.Events().GetAutoPayoutFilter(
		[]common.Address{e.bch.ContractRegistry().AutoPayout()}, big.NewInt(0)))
	if err != nil {
		return err
	}

	// listen routine

	ctx, cancel := context.WithCancel(ctx)

	errGroup := errgroup.Group{}
	errGroup.Go(func() error {
		err := o.watchPriceRoutine(ctx)
		if err != nil {
			o.logger.Error("price watching routine failed", zap.Error(err))
			cancel()
		}
		return err
	})
	errGroup.Go(func() error {
		err := o.submitPriceRoutine(ctx)
		if err != nil {
			o.logger.Error("price submission routine failed", zap.Error(err))
			cancel()
		}
		return err
	})
	errGroup.Go(func() error {
		err := o.listenEventsRoutine(ctx)
		if err != nil {
			o.logger.Error("event listening routine failed", zap.Error(err))
			cancel()
		}
		return err
	})
	return errGroup.Wait()
}

func (e *Eric) SetConfigs() {

}

func (e *Eric) doPayout(ctx context.Context, master common.Address) error {
	balance, err := e.bch.SidechainToken().BalanceOf(ctx, master)
	if err != nil {
		return err
	}

	return e.bch.AutoPayout().DoAutoPayout(ctx, e.key, master)
}
