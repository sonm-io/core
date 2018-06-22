package oracle

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

type Oracle struct {
	key *ecdsa.PrivateKey
	cfg *Config

	logger *zap.Logger

	bch blockchain.API

	actualPrice  *big.Int
	currentPrice *big.Int
}

func NewOracle(ctx context.Context, key *ecdsa.PrivateKey, cfg *Config) (*Oracle, error) {
	logger := ctxlog.GetLogger(ctx)

	var mode string
	if cfg.Oracle.Mode {
		mode = "master"
	} else {
		mode = "slave"
	}
	logger.Info("start USD-SNM Oracle",
		zap.String("mode", mode),
		zap.String("account", crypto.PubkeyToAddress(key.PublicKey).String()),
		zap.String("price update period:", cfg.Oracle.PriceUpdatePeriod.String()),
		zap.String("contract update period", cfg.Oracle.ContractUpdatePeriod.String()),
		zap.Float64("deviation percent", cfg.Oracle.Percent))

	bch, err := blockchain.NewAPI(ctx, blockchain.WithConfig(cfg.Blockchain))
	if err != nil {
		return nil, err
	}

	return &Oracle{
		logger:       logger,
		key:          key,
		cfg:          cfg,
		bch:          bch,
		currentPrice: big.NewInt(0),
		actualPrice:  big.NewInt(0),
	}, nil
}

func (o *Oracle) Serve(ctx context.Context) error {
	priceWatcher := NewPriceWatcher(o.cfg.Oracle.PriceUpdatePeriod)
	dw := priceWatcher.Start(ctx)

	t := util.NewImmediateTicker(o.cfg.Oracle.ContractUpdatePeriod)

	events, err := o.bch.Events().GetEvents(ctx, o.bch.Events().GetMultiSigFilter([]common.Address{o.bch.ContractRegistry().OracleUsdAddress()}, big.NewInt(0)))
	if err != nil {
		return err
	}

	for {
		select {
		case p := <-dw:
			if p.err != nil {
				return p.err
			}
			o.actualPrice = p.price
			o.logger.Debug("loaded new price", zap.String("price", p.price.String()))
		case <-t.C:
			currentPrice, err := o.bch.OracleUSD().GetCurrentPrice(ctx)
			if err != nil {
				return err
			}
			if o.currentPrice.Cmp(currentPrice) != 0 {
				o.currentPrice = currentPrice
				o.logger.Debug("current price changed", zap.String("price", o.currentPrice.String()))
			}
			// master mode
			if o.cfg.Oracle.Mode {
				go o.SetPrice(ctx)
			}
		case event, ok := <-events:
			if !ok {
				return fmt.Errorf("events chanel closed")
			}
			// slave mode
			if !o.cfg.Oracle.Mode {
				o.logger.Debug("new submission", zap.Any("event", event))
				switch value := event.Data.(type) {
				case *blockchain.SubmissionData:
					if o.transactionValid(ctx, value.TransactionId) {
						go o.confirmChanging(ctx, value.TransactionId)
					}
				}
			}
		}
	}
}

func (o *Oracle) SetPrice(ctx context.Context) error {
	if o.actualPrice == nil {
		o.logger.Debug("set price dropped, actual price not downloaded")
		return fmt.Errorf("price is nil")
	}
	o.logger.Info("submitting new price", zap.String("price", o.actualPrice.String()))
	data, err := o.bch.OracleUSD().PackSetCurrentPriceTransactionData(o.actualPrice)
	if err != nil {
		return err
	}
	return o.bch.OracleMultiSig().SubmitTransaction(ctx, o.key, o.bch.ContractRegistry().OracleUsdAddress(), big.NewInt(0), data)
}

// checkPrice check that submitted price does not differ more than 1%
// incapsulate ugly big math to separate function
func (o *Oracle) checkPrice(price *big.Int) bool {
	diff := big.NewFloat(0).SetInt(o.actualPrice)
	diff.Sub(diff, big.NewFloat(0).SetInt(price))
	diff.Abs(diff)

	p := big.NewInt(0).Set(o.actualPrice)
	p.Div(p, big.NewInt(100))
	percent := big.NewFloat(o.cfg.Oracle.Percent)
	percent.Mul(percent, big.NewFloat(0).SetInt(p))

	if diff.Cmp(percent) == -1 {
		return true
	}
	return false
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
