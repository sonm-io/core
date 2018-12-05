package oracle

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Oracle struct {
	key          *ecdsa.PrivateKey
	cfg          *Config
	logger       *zap.Logger
	bch          blockchain.API
	actualPrice  *big.Int
	currentPrice *big.Int
	mu           sync.Mutex
}

func NewOracle(ctx context.Context, cfg *Config) (*Oracle, error) {
	bch, err := blockchain.NewAPI(ctx, blockchain.WithConfig(cfg.Blockchain))
	if err != nil {
		return nil, err
	}

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load Ethereum keys: %s", err)
	}

	return &Oracle{
		logger:       ctxlog.GetLogger(ctx),
		key:          key,
		cfg:          cfg,
		bch:          bch,
		currentPrice: big.NewInt(0),
		actualPrice:  big.NewInt(0),
	}, nil
}

func (o *Oracle) watchPriceRoutine(ctx context.Context) error {
	priceWatcher := NewPriceWatcher(o.cfg.Oracle.PriceUpdatePeriod)
	dw := priceWatcher.Start(ctx)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case p := <-dw:
			if p.err != nil {
				o.logger.Warn("failed to load price", zap.Error(p.err))
				continue
			}
			o.mu.Lock()
			o.actualPrice = p.price
			o.mu.Unlock()
			o.logger.Debug("loaded new price", zap.String("price", p.price.String()))
		}
	}
}

func (o *Oracle) submitPriceRoutine(ctx context.Context) error {
	if !o.cfg.Oracle.IsMaster {
		return nil
	}
	t := util.NewImmediateTicker(o.cfg.Oracle.ContractUpdatePeriod)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			if err := o.SetPrice(ctx); err != nil {
				o.logger.Warn("failed to submit new price", zap.Error(err))
			}
		}
	}
}

func (o *Oracle) listenEventsRoutine(ctx context.Context) error {
	if o.cfg.Oracle.IsMaster {
		return nil
	}

	var lastBlock uint64 = 0
	if o.cfg.Oracle.FromNow {
		var err error
		lastBlock, err = o.bch.Events().GetLastBlock(ctx)
		if err != nil {
			return err
		}
	}

	events, err := o.bch.Events().GetEvents(ctx, o.bch.Events().GetMultiSigFilter(
		[]common.Address{o.bch.ContractRegistry().OracleMultiSig()}, big.NewInt(0).SetUint64(lastBlock)))
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
			switch value := event.Data.(type) {
			case *blockchain.SubmissionData:
				o.logger.Debug("new submission found", zap.Any("event", event))
				if o.transactionValid(ctx, value.TransactionId) {
					o.confirmChanging(ctx, value.TransactionId)
				} else {
					o.logger.Info("failed to confirm transaction, transaction invalid ")
				}
			}
		}
	}
}

func (o *Oracle) Serve(ctx context.Context) error {
	o.logger.Info("creating USD-SNM Oracle",
		zap.Bool("is_master", o.cfg.Oracle.IsMaster),
		zap.String("account", crypto.PubkeyToAddress(o.key.PublicKey).String()),
		zap.String("price update period:", o.cfg.Oracle.PriceUpdatePeriod.String()),
		zap.String("contract update period", o.cfg.Oracle.ContractUpdatePeriod.String()),
		zap.Float64("deviation percent", o.cfg.Oracle.Percent),
		zap.Bool("from now", o.cfg.Oracle.FromNow))

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

func (o *Oracle) getPriceForSubmit() (*big.Int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.actualPrice == nil {
		return nil, fmt.Errorf("actual price is not downloaded")
	}

	if o.actualPrice.Cmp(big.NewInt(1e15)) < 0 {
		return nil, fmt.Errorf("oracle mustn't automaticly set price lower than 1e15")
	}

	if o.actualPrice.Cmp(big.NewInt(0).Mul(big.NewInt(params.Ether), big.NewInt(1e3))) > 0 {
		return nil, fmt.Errorf("oracle mustn't automaticly set price greater than 1e21")
	}
	return big.NewInt(0).Set(o.actualPrice), nil
}

func (o *Oracle) SetPrice(ctx context.Context) error {
	price, err := o.getPriceForSubmit()
	if err != nil {
		return fmt.Errorf("failed to get price for submission: %s", err)
	}
	o.logger.Info("submitting new price", zap.String("price", price.String()))
	data, err := o.bch.OracleUSD().PackSetCurrentPriceTransactionData(price)
	if err != nil {
		return err
	}
	return o.bch.OracleMultiSig().SubmitTransaction(ctx, o.key, o.bch.ContractRegistry().OracleUsdAddress(), big.NewInt(0), data)
}

// checkPrice check that submitted price does not differ more than 1%
// incapsulate ugly big math to separate function
func (o *Oracle) checkPrice(price *big.Int) bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	diff := big.NewInt(0).Set(o.actualPrice)
	diff.Sub(diff, price)
	diff.Abs(diff)

	p := big.NewInt(0).Set(o.actualPrice)
	p.Div(p, big.NewInt(100))
	percent := big.NewFloat(o.cfg.Oracle.Percent)
	percent.Mul(percent, big.NewFloat(0).SetInt(p))

	return big.NewFloat(0).SetInt(diff).Cmp(percent) == -1
}

// transactionValid check transaction parameters
func (o *Oracle) transactionValid(ctx context.Context, txID *big.Int) bool {
	transaction, err := o.bch.OracleMultiSig().GetTransaction(ctx, txID)
	if err != nil {
		return false
	}
	if transaction.Executed {
		o.logger.Debug("transaction already executed", zap.String("transactionID", txID.String()))
		return false
	}
	if transaction.To.String() != o.bch.ContractRegistry().OracleUsdAddress().String() {
		o.logger.Debug("transaction aimed to strange address",
			zap.String("transactionID", txID.String()),
			zap.String("transactionTo", transaction.To.String()))
		return false
	}
	price, err := o.bch.OracleUSD().UnpackSetCurrentPriceTransactionData(transaction.Data)
	if err != nil {
		return false
	}
	if !o.checkPrice(price) {
		o.logger.Debug("invalid price in transaction",
			zap.String("transactionID", txID.String()),
			zap.String("price", price.String()))
		return false
	}
	return true
}

func (o *Oracle) confirmChanging(ctx context.Context, txID *big.Int) error {
	o.logger.Debug("confirm transaction", zap.Any("transactionID", txID.String()))
	err := o.bch.OracleMultiSig().ConfirmTransaction(ctx, o.key, txID)
	if err != nil {
		o.logger.Debug("confirm failed", zap.Any("transactionID", txID.String()))
		return err
	}
	o.logger.Debug("confirm successful", zap.Any("transactionID", txID.String()))
	return nil
}
