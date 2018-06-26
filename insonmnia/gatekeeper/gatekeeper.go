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
	"golang.org/x/sync/errgroup"
)

type Gatekeeper struct {
	key          *ecdsa.PrivateKey
	cfg          *Config
	logger       *zap.Logger
	bch          blockchain.API
	in           blockchain.SimpleGatekeeperAPI
	out          blockchain.SimpleGatekeeperAPI
	freezingTime time.Duration
	spentToday   *big.Int
	lastDay      *time.Time
	mu           sync.Mutex
}

func NewGatekeeper(ctx context.Context, key *ecdsa.PrivateKey, cfg *Config) (*Gatekeeper, error) {
	logger := ctxlog.GetLogger(ctx)

	logger.Info("start gatekeeper instance",
		zap.String("direction", cfg.Gatekeeper.Direction),
		zap.Duration("delay", cfg.Gatekeeper.Delay),
		zap.String("key", crypto.PubkeyToAddress(key.PublicKey).String()))

	bch, err := blockchain.NewAPI(ctx, blockchain.WithConfig(cfg.Blockchain))
	if err != nil {
		return nil, err
	}

	var in blockchain.SimpleGatekeeperAPI
	var out blockchain.SimpleGatekeeperAPI

	if cfg.Gatekeeper.Direction == masterchainDirection {
		in = bch.SidechainGate()
		out = bch.MasterchainGate()
	} else if cfg.Gatekeeper.Direction == sidechainDirection {
		in = bch.MasterchainGate()
		out = bch.SidechainGate()
	}

	keeper, err := out.GetKeeper(ctx, crypto.PubkeyToAddress(key.PublicKey))
	if err != nil {
		return nil, err
	}

	if keeper.DayLimit.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("used key is not keeper")
	}

	if keeper.Frozen {
		return nil, fmt.Errorf("keeper with given key is frozen")
	}

	logger.Info("start gatekeeper instance",
		zap.String("direction", cfg.Gatekeeper.Direction),
		zap.Duration("delay", cfg.Gatekeeper.Delay),
		zap.String("key", crypto.PubkeyToAddress(key.PublicKey).String()),
		zap.String("day limit", keeper.DayLimit.String()),
		zap.String("spent today", keeper.SpentToday.String()))

	return &Gatekeeper{
		key:    key,
		cfg:    cfg,
		in:     in,
		out:    out,
		logger: logger,
		bch:    bch,
	}, nil
}

func (g *Gatekeeper) freezingTimeRoutine(ctx context.Context) error {
	// straight load freezing time
	err := g.loadFreezeTime(ctx)
	if err != nil {
		return err
	}

	t := util.NewImmediateTicker(g.cfg.Gatekeeper.ReloadFreezingPeriod)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			err = g.loadFreezeTime(ctx)
			if err != nil {
				g.logger.Warn("failed to reload freezing time", zap.Error(err))
			}
		}
	}
}

func (g *Gatekeeper) payoutRoutine(ctx context.Context) error {
	t := util.NewImmediateTicker(g.cfg.Gatekeeper.Period)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			err := g.processTransaction(ctx)
			if err != nil {
				g.logger.Warn("failed to process transactions", zap.Error(err))
			}
		}
	}
}

func (g *Gatekeeper) scummyFinderRoutine(ctx context.Context) error {
	t := util.NewImmediateTicker(g.cfg.Gatekeeper.Period)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			inTxs, outTxs, err := g.loadTransactions(ctx)
			if err != nil {
				g.logger.Warn("failed to load transactions", zap.Error(err))
			}
			g.findScummyTransactions(ctx, inTxs, outTxs)
		}
	}
}

func (g *Gatekeeper) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	errGroup := errgroup.Group{}
	errGroup.Go(func() error {
		err := g.freezingTimeRoutine(ctx)
		if err != nil {
			g.logger.Error("freezing time reload routine failed", zap.Error(err))
			cancel()
		}
		return err
	})
	errGroup.Go(func() error {
		err := g.payoutRoutine(ctx)
		if err != nil {
			g.logger.Error("payout routine failed", zap.Error(err))
			cancel()
		}
		return err
	})
	errGroup.Go(func() error {
		err := g.scummyFinderRoutine(ctx)
		if err != nil {
			g.logger.Error("scummy finder routine failed", zap.Error(err))
			cancel()
		}
		return err
	})

	return errGroup.Wait()
}

func (g *Gatekeeper) processTransaction(ctx context.Context) error {
	g.logger.Debug("start transaction processing")

	inTxs, outTxs, err := g.loadTransactions(ctx)
	if err != nil {
		return err
	}

	errG := errgroup.Group{}

	// find payin transactions doesn't exists in payout
	for k, inTx := range inTxs {
		_, ok := outTxs[k]
		if !ok {
			g.logger.Debug("found new unpaid transaction",
				zap.String("from", inTx.From.String()),
				zap.String("value", inTx.Value.String()),
				zap.String("tx number", inTx.Number.String()),
				zap.Uint64("block number", inTx.BlockNumber))

			if err := g.processUnpaidTransaction(ctx, inTx); err != nil {
				g.logger.Warn("failed to process unpaid transaction", zap.Error(err))
			}
			return nil
		}
	}
	errG.Wait()
	g.logger.Debug("finish transaction processing")
	return nil
}

func (g *Gatekeeper) loadTransactions(ctx context.Context) (map[string]*blockchain.GateTx, map[string]*blockchain.GateTx, error) {
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
	payinTime, err := g.in.GetGateTransactionTime(ctx, tx)
	if err != nil {
		return false
	}

	// g.logger.Debug("delay check",
	// 	zap.Time("time with delay", payinTime.Add(g.cfg.Gatekeeper.Delay)),
	// 	zap.Time("nowTime", time.Now().UTC()))

	return payinTime.Add(g.cfg.Gatekeeper.Delay).Before(time.Now().UTC())
}

// checkUnpaid verify that tx paid already
func (g *Gatekeeper) isNotPaid(ctx context.Context, tx *blockchain.GateTx) bool {
	txState, err := g.out.GetTransactionState(ctx, tx.From, tx.Value, tx.Number)
	if err != nil {
		g.logger.Debug("err while getting tx data", zap.Error(err))
		return false
	}
	g.logger.Debug("tx state in isNotPaid",
		zap.Bool("paid", txState.Paid))
	return !txState.Paid
}

// verify that transaction is committed already
func (g *Gatekeeper) isCommitted(ctx context.Context, txState *blockchain.GateTxState) bool {
	return txState.CommitTS.Unix() != 0
}

// verify that transaction committed by this gate instance
func (g *Gatekeeper) transactionCommittedByMe(ctx context.Context, txState *blockchain.GateTxState) bool {
	return txState.Keeper.String() == crypto.PubkeyToAddress(g.key.PublicKey).String()
}

// verify that transaction committed by this gate instance
func (g *Gatekeeper) transactionOnQuarantine(ctx context.Context, txState *blockchain.GateTxState) bool {
	return txState.CommitTS.Add(g.freezingTime).Before(time.Now().UTC())
}

func (g *Gatekeeper) underLimit(ctx context.Context, tx *blockchain.GateTx) bool {
	keeper, err := g.out.GetKeeper(ctx, crypto.PubkeyToAddress(g.key.PublicKey))
	if err != nil {
		return false
	}

	spentToday := keeper.SpentToday

	if keeper.LastDay.Day() < time.Now().UTC().Day() {
		spentToday = big.NewInt(0)
	}

	// keeper.spent_today + tx.value > dayLimit
	if spentToday.Add(spentToday, tx.Value).Cmp(keeper.DayLimit) == 1 {
		log.Debug("tx over keper limit",
			zap.String("spent today", spentToday.String()),
			zap.String("tx value", tx.Value.String()),
			zap.String("keeper day limit", keeper.DayLimit.String()),
			zap.String("spentToday + value", spentToday.Add(spentToday, tx.Value).String()))
		return false
	}

	return true
}

func (g *Gatekeeper) processUnpaidTransaction(ctx context.Context, tx *blockchain.GateTx) error {
	if !g.underLimit(ctx, tx) {
		g.logger.Debug("tx over keeper limit")
		return fmt.Errorf("tx over keeper limit")
	}

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
	txState, err := g.out.GetTransactionState(ctx, tx.From, tx.Value, tx.Number)
	if err != nil {
		return err
	}

	if !g.isCommitted(ctx, txState) {
		_, err := g.out.Payout(ctx, g.key, tx.From, tx.Value, tx.Number)
		if err != nil {
			g.logger.Debug("error while commit", zap.Error(err))
			return err
		}
		g.logger.Info("transaction committed",
			zap.String("from", tx.From.String()),
			zap.String("value", tx.Value.String()),
			zap.String("tx number", tx.Number.String()))

		g.logger.Debug("transaction going to quarantine")
		return nil
	}

	if !g.transactionCommittedByMe(ctx, txState) {
		g.logger.Debug("transaction commited by other gate")
		return fmt.Errorf("transaction commited by other gate")
	}

	if !g.transactionOnQuarantine(ctx, txState) {
		g.logger.Debug("transaction on quarantine now",
			zap.Time("commitTS", txState.CommitTS))
		return fmt.Errorf("transaction on quarantine now")
	}

	g.logger.Info("payout transaction",
		zap.String("keeper", txState.Keeper.String()),
		zap.String("commitTs", txState.CommitTS.String()),
		zap.Bool("paid", txState.Paid),
		zap.String("from", tx.From.String()),
		zap.String("value", tx.Value.String()),
		zap.String("tx number", tx.Number.String()))

	_, err = g.out.Payout(ctx, g.key, tx.From, tx.Value, tx.Number)
	if err != nil {
		g.logger.Debug("error while payout", zap.Error(err))
		return err
	}
	g.logger.Debug("transaction payouted",
		zap.String("from", tx.From.String()),
		zap.String("value", tx.Value.String()),
		zap.String("tx number", tx.Number.String()))

	return nil
}

func (g *Gatekeeper) findScummyTransactions(ctx context.Context, inTxs map[string]*blockchain.GateTx, outTxs map[string]*blockchain.GateTx) {
	for k, tx := range outTxs {
		_, ok := inTxs[k]
		if !ok {
			err := g.processScummyTx(ctx, tx)
			if err != nil {
				log.Warn("failed to process scummy transaction", zap.Error(err))
			}
		}
	}
}

func (g *Gatekeeper) processScummyTx(ctx context.Context, tx *blockchain.GateTx) error {
	txState, err := g.out.GetTransactionState(ctx, tx.From, tx.Value, tx.Number)
	if err != nil {
		return err
	}

	keeper, err := g.out.GetKeeper(ctx, txState.Keeper)
	if err != nil {
		return err
	}

	g.logger.Info("found new scum transaction",
		zap.String("from", tx.From.String()),
		zap.String("value", tx.Value.String()),
		zap.String("tx number", tx.Number.String()),
		zap.Uint64("block number", tx.BlockNumber),
		zap.Time("commit timestamp", txState.CommitTS),
		zap.Bool("paid", txState.Paid),
		zap.String("keeper address", keeper.Address.String()),
		zap.String("keeper day limit", keeper.DayLimit.String()),
		zap.String("keeper spent today", keeper.SpentToday.String()),
		zap.String("keeper last day ", keeper.LastDay.String()),
		zap.Bool("keeper frozen", keeper.Frozen))

	if keeper.Frozen {
		return fmt.Errorf("keeper already frozen")
	}

	return g.out.FreezeKeeper(ctx, g.key, txState.Keeper)
}

// loadFreezeTime watch current freezing time in contract
func (g *Gatekeeper) loadFreezeTime(ctx context.Context) error {
	freezingTime, err := g.out.GetFreezingTime(ctx)
	if err != nil {
		return err
	}
	g.logger.Debug("changing freezing time", zap.Duration("freezing time", g.freezingTime))

	g.mu.Lock()
	defer g.mu.Unlock()

	g.freezingTime = freezingTime
	return nil
}
