package eric

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/blockchain"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"sync"
)

type Eric struct {
	key    *ecdsa.PrivateKey
	logger *zap.Logger
	bch    blockchain.API
	cfg    *Config

	payoutSettings map[string]*struct {
		mu sync.Mutex
		*blockchain.AutoPayoutSetting
	}
}

func NewEric() (*Eric, error) {
	return &Eric{}, nil
}

func (e *Eric) Start(ctx context.Context) error {
	// load configs

	lastBlock, err := e.bch.Events().GetLastBlock(ctx)
	if err != nil {
		return err
	}

	settings, err := e.bch.AutoPayout().GetPayoutSettings(ctx)
	if err != nil {
		return err
	}

	for _, s := range settings {

	}

	// listen routine

	ctx, cancel := context.WithCancel(ctx)

	errGroup := errgroup.Group{}
	errGroup.Go(func() error {
		err := e.eventsRoutine(ctx)
		if err != nil {
			e.logger.Error("price watching routine failed", zap.Error(err))
			cancel()
		}
		return err
	})
	return errGroup.Wait()
}

func (e *Eric) eventsRoutine(ctx context.Context) error {

}

func (e *Eric) SetConfigs(s *blockchain.AutoPayoutSetting) {
	e.payoutSettings[s.Master.String()].AutoPayoutSetting = s
}

func (e *Eric) doPayout(ctx context.Context, master common.Address) error {
	balance, err := e.bch.SidechainToken().BalanceOf(ctx, master)
	if err != nil {
		return err
	}

	// TODO
	if balance.SNM.Cmp(e.payoutSettings[master.String()].LowLimit) < 0{
		return fmt.Errorf("balance lower than low limit")
	}

	return e.bch.AutoPayout().DoAutoPayout(ctx, e.key, master)
}
