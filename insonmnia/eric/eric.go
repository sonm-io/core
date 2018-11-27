package eric

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"math/big"
	"sync"
	"time"
)

type Eric struct {
	key    *ecdsa.PrivateKey
	logger *zap.Logger
	bch    blockchain.API
	cfg    *Config

	payoutSettings map[string]*struct {
		mu             sync.Mutex
		currentBalance *big.Int
		*blockchain.AutoPayoutSetting
	}
}

func NewEric(ctx context.Context, key *ecdsa.PrivateKey, cfg *Config) (*Eric, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(30*time.Second))
	defer cancel()
	bch, err := blockchain.NewAPI(ctx, blockchain.WithConfig(cfg.Blockchain))
	if err != nil {
		return nil, err
	}
	return &Eric{
		cfg:    cfg,
		key:    key,
		bch:    bch,
		logger: ctxlog.GetLogger(ctx),
	}, nil
}

func (e *Eric) Start(ctx context.Context) error {
	// load configs

	lastBlock, err := e.bch.Events().GetLastBlock(ctx)
	if err != nil {
		return err
	}

	settings, err := e.bch.AutoPayout().GetPayoutSettings(ctx, big.NewInt(0).SetUint64(lastBlock))
	if err != nil {
		return err
	}

	for _, s := range settings {
		e.setConfigs(s)
	}

	// listen routine

	ctx, cancel := context.WithCancel(ctx)

	errGroup := errgroup.Group{}
	errGroup.Go(func() error {
		err := e.eventsRoutine(ctx)
		if err != nil {
			e.logger.Error("event watching routine failed", zap.Error(err))
			cancel()
		}
		return err
	})
	return errGroup.Wait()
}

func (e *Eric) eventsRoutine(ctx context.Context) error {
	events, err := e.bch.Events().GetEvents(ctx, e.bch.Events().GetMultiSigFilter(
		[]common.Address{e.bch.ContractRegistry().OracleMultiSig()}, big.NewInt(0)))
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-events:
			if !ok {
				return fmt.Errorf("events chanel closed")
			}
			switch data := event.Data.(type) {
			case *blockchain.AutoPayoutSetting:
				e.logger.Debug("new setting found", zap.Any("event", event))
				e.setConfigs(data)
			case *blockchain.TransferData:
				e.logger.Debug("new transfer found", zap.Any("event", event))
				if _, ok := e.payoutSettings[data.To.String()]; !ok {
					return fmt.Errorf("address not found")
				}
				err := e.doPayout(ctx, data.To)
				return err
			}

		}
	}
}

func (e *Eric) setConfigs(s *blockchain.AutoPayoutSetting) {
	e.payoutSettings[s.Master.String()].mu.Lock()
	e.payoutSettings[s.Master.String()].AutoPayoutSetting = s
	e.payoutSettings[s.Master.String()].mu.Unlock()
}

func (e *Eric) doPayout(ctx context.Context, master common.Address) error {
	balance, err := e.bch.SidechainToken().BalanceOf(ctx, master)
	if err != nil {
		return err
	}

	// current currentBalance < low limit
	if balance.SNM.Cmp(e.payoutSettings[master.String()].LowLimit) == -1 {
		return fmt.Errorf("currentBalance lower than low limit")
	}

	return e.bch.AutoPayout().DoAutoPayout(ctx, e.key, master)
}
