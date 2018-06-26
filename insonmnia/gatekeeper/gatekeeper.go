package gatekeeper

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

type Gatekeeper struct {
	key *ecdsa.PrivateKey
	cfg *Config

	logger *zap.Logger

	bch blockchain.API

	in  blockchain.SimpleGatekeeperAPI
	out blockchain.SimpleGatekeeperAPI

	freezingTime time.Duration

	spentToday *big.Int
	lastDay    *time.Time

	mu sync.Mutex
}

func NewGatekeeper(ctx context.Context, key *ecdsa.PrivateKey, cfg *Config) (*Gatekeeper, error) {
	logger := ctxlog.GetLogger(ctx)

	logger.Info("start gatekeeper instance",
		zap.String("direction", cfg.Gatekeeper.Direction),
		zap.Duration("delay", cfg.Gatekeeper.Delay),
		zap.String("key", crypto.PubkeyToAddress(key.PublicKey).String()))

	bch, err := blockchain.NewAPI(blockchain.WithConfig(cfg.Blockchain))
	if err != nil {
		return nil, err
	}

	var in blockchain.SimpleGatekeeperAPI
	var out blockchain.SimpleGatekeeperAPI

	if cfg.Gatekeeper.Direction == "masterchain" {
		in = bch.SidechainGate()
		out = bch.MasterchainGate()
	} else if cfg.Gatekeeper.Direction == "sidechain" {
		in = bch.MasterchainGate()
		out = bch.SidechainGate()
	}

	return &Gatekeeper{
		key:    key,
		cfg:    cfg,
		in:     in,
		out:    out,
		logger: logger,
		bch:    bch,
	}, nil
}

func (g *Gatekeeper) Serve(ctx context.Context) error {
	t := util.NewImmediateTicker(g.cfg.Gatekeeper.Period)

	// straight load freezing time
	g.loadFreezeTime(ctx)

	freezingTimeLoader := util.NewImmediateTicker(g.cfg.Gatekeeper.ReloadFreezingPeriod)

	for {
		select {
		case <-t.C:
			g.processTransaction(ctx)
		case <-freezingTimeLoader.C:
			go g.loadFreezeTime(ctx)
		}
	}
}

func (g *Gatekeeper) processTransaction(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.logger.Debug("start transaction processing")

	inTxs, outTxs, err := g.loadTransactions(ctx)
	if err != nil {
		return err
	}

	// find payin transactions doesn't exists in payout
	for k, inTx := range inTxs {
		_, ok := outTxs[k]
		if !ok {
			g.logger.Debug("found new unpaid transaction",
				zap.String("from", inTx.From.String()),
				zap.String("value", inTx.Value.String()),
				zap.String("tx number", inTx.Number.String()),
				zap.String("block number", inTx.BlockNumber.String()))

			go g.processUnpaidTransaction(ctx, inTx)
		}
	}

	g.findScummyTransactions(ctx, inTxs, outTxs)
	g.logger.Debug("finish transaction processing")
	return nil
}

func (g *Gatekeeper) loadTransactions(ctx context.Context) (map[string]*blockchain.GateTx, map[string]*blockchain.GateTx, error) {
	var err error
	inTxs, err := g.in.GetPayinTransactions(ctx)
	if err != nil {
		return nil, nil, err
	}
	outTxs, err := g.out.GetPayoutTransactions(ctx)
	if err != nil {
		return nil, nil, err
	}

	g.logger.Info("loaded transactions",
		zap.Int("amount of payin transactions", len(inTxs)),
		zap.Int("amount of payout transactions", len(outTxs)))

	return inTxs, outTxs, nil
}

// checkDelay verify that tx out of delay
func (g *Gatekeeper) checkDelay(ctx context.Context, tx *blockchain.GateTx) bool {
	payinTimestamp, err := g.bch.Events().GetBlockTimestamp(ctx, tx.BlockNumber)
	if err != nil {
		return false
	}
	// cast delay, time of payin tx and now time to uint64
	payinTime := int64(time.Unix(int64(payinTimestamp), 0).Second())
	delay := int64(g.cfg.Gatekeeper.Delay.Seconds())
	nowTime := int64(time.Now().Second())

	g.logger.Debug("delay check", zap.Int64("delay", payinTime+delay), zap.Int64("nowTime", nowTime))
	if nowTime >= payinTime+delay {
		return false
	}
	return true
}

// checkUnpaid verify that tx paid already
func (g *Gatekeeper) isNotPaid(ctx context.Context, tx *blockchain.GateTx) bool {
	// verify that transaction not payout now
	txState, err := g.out.GetTransactionState(ctx, tx.From, tx.Value, tx.Number)
	if err != nil {
		g.logger.Debug("err while getting tx data", zap.Error(err))
		return false
	}

	g.logger.Debug("transaction state",
		zap.Any("commitTS", txState.CommitTS),
		zap.Any("keeper", txState.Keeper),
		zap.Any("paid", txState.Paid))
	return !txState.Paid
}

func (g *Gatekeeper) processUnpaidTransaction(ctx context.Context, tx *blockchain.GateTx) error {
	if !g.checkDelay(ctx, tx) {
		g.logger.Debug("not cover delay check")
		return fmt.Errorf("not cover delay check")
	}

	if !g.isNotPaid(ctx, tx) {
		g.logger.Debug("transaction already paid")
		return fmt.Errorf("transaction already paid")
	}

	return g.Payout(ctx, tx)
}

func (g *Gatekeeper) Payout(ctx context.Context, tx *blockchain.GateTx) error {
	g.logger.Debug("fix transaction",
		zap.String("from", tx.From.String()),
		zap.String("value", tx.Value.String()),
		zap.String("tx number", tx.Number.String()))

	err := g.out.CommitPayout(ctx, g.key, tx.From, tx.Value, tx.Number)
	if err != nil {
		g.logger.Debug("error while commit", zap.Error(err))
		return err
	}

	g.logger.Debug("transaction committed")

	// sleeping for freezing time
	time.Sleep(g.freezingTime)

	err = g.out.Payout(ctx, g.key, tx.From, tx.Value, tx.Number)
	if err != nil {
		g.logger.Debug("error while payout", zap.Error(err))
		return err
	}
	g.logger.Debug("transaction payout")

	return nil
}

func (g *Gatekeeper) findScummyTransactions(ctx context.Context, inTxs map[string]*blockchain.GateTx, outTxs map[string]*blockchain.GateTx) {
	log.Debug("starting verify transactions")
	for k, tx := range outTxs {
		_, ok := inTxs[k]
		if !ok {
			go g.freezeScummy(ctx, tx)
		}
	}
}

func (g *Gatekeeper) freezeScummy(ctx context.Context, tx *blockchain.GateTx) error {
	g.logger.Debug("found new scum transaction",
		zap.String("from", tx.From.String()),
		zap.String("value", tx.Value.String()),
		zap.String("tx number", tx.Number.String()),
		zap.String("block number", tx.BlockNumber.String()))

	txState, err := g.out.GetTransactionState(ctx, tx.From, tx.Value, tx.Number)
	if err != nil {
		return err
	}

	keeper, err := g.out.GetKeeper(ctx, txState.Keeper)
	if err != nil {
		return err
	}

	if keeper.Frozen {
		return fmt.Errorf("keeper already frozen")
	}

	return g.out.FreezeKeeper(ctx, g.key, txState.Keeper)
}

// loadFreezeTime watch current freezing time in contract
func (g *Gatekeeper) loadFreezeTime(ctx context.Context) error {
	var freezingTime *big.Int
	var err error

	freezingTime, err = g.out.GetFreezingTime(ctx)
	if err != nil {
		return err
	}

	if !freezingTime.IsInt64() {
		return fmt.Errorf("freezing time is abused")
	}

	// multiply to second because blockchain operate with time in seconds
	ft := time.Duration(freezingTime.Int64()) * time.Second
	if g.freezingTime == ft {
		return nil
	}
	g.freezingTime = ft

	g.logger.Debug("changing freezing time", zap.Duration("freezing time", g.freezingTime))

	return nil
}
