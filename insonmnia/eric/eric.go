package eric

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	"go.uber.org/zap"
	"math/big"
	"time"
)

type Eric struct {
	key    *ecdsa.PrivateKey
	logger *zap.Logger
	bch    blockchain.API
	cfg    *Config

	payoutSettings map[string]*blockchain.AutoPayoutSetting
	payoutChan     chan *blockchain.TransferData
}

func NewEric(ctx context.Context, cfg *Config) (*Eric, error) {
	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load Ethereum keys: %s", err)
	}

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
	lastBlock, err := e.prestart(ctx)
	if err != nil {
		e.logger.Error("prestart failed", zap.Error(err))
		return err
	}

	err = e.start(ctx, lastBlock)
	if err != nil {
		e.logger.Error("event watching failed", zap.Error(err))
		return err
	}
	return nil
}

func (e *Eric) prestart(ctx context.Context) (uint64, error) {
	// load configs
	e.logger.Info("starting eric")

	lastBlock, err := e.bch.Events().GetLastBlock(ctx)
	if err != nil {
		return 0, err
	}

	// TODO: make blockConfirmations configurable and rewrite this part
	lastBlock = lastBlock - 5

	settings, err := e.bch.AutoPayout().GetPayoutSettings(ctx, big.NewInt(0).SetUint64(lastBlock))
	if err != nil {
		return 0, err
	}

	e.logger.Debug("config loaded",
		zap.Int("configs amount", len(settings)),
		zap.Uint64("last processed block", lastBlock))

	for _, s := range settings {
		e.setConfigs(s)
	}

	return lastBlock, nil
}

func (e *Eric) start(ctx context.Context, fromBlock uint64) error {
	e.logger.Debug("start event routine")
	events, err := e.bch.Events().GetEvents(ctx, e.bch.Events().GetAutoPayoutFilter(big.NewInt(0).SetUint64(fromBlock)))
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
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
				e.processTransferData(data)
			}
		case data := <-e.payoutChan:
			err := e.doPayout(ctx, data.To)
			if err != nil {
				return err
			}
			return nil
		}
	}
}

func (e *Eric) setConfigs(s *blockchain.AutoPayoutSetting) {
	e.payoutSettings[s.Master.String()] = s
}

func (e *Eric) processTransferData(data *blockchain.TransferData) {
	if _, ok := e.payoutSettings[data.To.String()]; !ok {
		e.logger.Debug("address not set autopayout", zap.String("address", data.To.String()))
		return
	}
	e.payoutChan <- data
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

	e.logger.Info("start new payoutChan",
		zap.String("address", master.String()),
		zap.String("lowLimit", e.payoutSettings[master.String()].LowLimit.String()),
		zap.String("balance", balance.SNM.String()))
	return e.bch.AutoPayout().DoAutoPayout(ctx, e.key, master)
}
