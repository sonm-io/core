package connor

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/connor/database"
	"github.com/sonm-io/core/connor/watchers"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/credentials"
)

const (
	poolReportedHashrateURL    = "https://api.nanopool.org/v1/eth/reportedhashrates/"
	poolAverageHashrateURL     = "https://api.nanopool.org/v1/eth/avghashrateworkers/"
	poolZECAverageHashrateURL  = "https://api.nanopool.org/v1/zec/avghashrateworkers/"
	poolZECReportedHashrateURL = "https://api.nanopool.org/v1/zec/reportedhashrate/"
)

type Connor struct {
	key          *ecdsa.PrivateKey
	Market       sonm.MarketClient
	TaskClient   sonm.TaskManagementClient
	DealClient   sonm.DealManagementClient
	TokenClient  sonm.TokenManagementClient
	MasterClient sonm.MasterManagementClient

	cfg    *Config
	db     *database.Database
	logger *zap.Logger
}

func NewConnor(ctx context.Context, key *ecdsa.PrivateKey, cfg *Config) (*Connor, error) {
	creds, err := newCredentials(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("can't create TLS credentials: %v", err)
	}

	nodeCC, err := xgrpc.NewClient(ctx, cfg.Market.Endpoint, creds)
	if err != nil {
		return nil, fmt.Errorf("can't create node connection: %v", err)
	}

	db, err := database.NewDatabaseConnect(cfg.Database.Driver, cfg.Database.DataSource)
	if err != nil {
		return nil, fmt.Errorf("cannot create database connection: %v", err)
	}

	connor := &Connor{
		key:          key,
		cfg:          cfg,
		Market:       sonm.NewMarketClient(nodeCC),
		TaskClient:   sonm.NewTaskManagementClient(nodeCC),
		DealClient:   sonm.NewDealManagementClient(nodeCC),
		TokenClient:  sonm.NewTokenManagementClient(nodeCC),
		MasterClient: sonm.NewMasterManagementClient(nodeCC),
		logger:       ctxlog.GetLogger(ctx),
		db:           db,
	}

	connor.logger = ctxlog.GetLogger(ctx)
	connor.db, err = database.NewDatabaseConnect(connor.cfg.Database.Driver, connor.cfg.Database.DataSource)
	if err != nil {
		return nil, err
	}
	return connor, nil
}

func (c *Connor) Serve(ctx context.Context) error {
	c.logger.Info("сonnor started work ...")
	defer c.logger.Info("сonnor has been stopped")

	c.clearStart()

	balanceReply, err := c.TokenClient.Balance(ctx, &sonm.Empty{})
	if err != nil {
		c.logger.Error("cannot load balance reply", zap.Error(err))
		return err
	}

	c.logger.Info("balance", zap.String("live", balanceReply.GetLiveBalance().ToPriceString()),
		zap.String("side", balanceReply.GetSideBalance().ToPriceString()))
	c.logger.Info("configuring connor", zap.Any("config", c.cfg))

	dataUpdate := util.NewImmediateTicker(c.cfg.Tickers.DataUpdate)
	defer dataUpdate.Stop()
	tradeUpdate := util.NewImmediateTicker(c.cfg.Tickers.TradeTicker)
	defer tradeUpdate.Stop()
	poolUpdate := time.NewTicker(c.cfg.Tickers.PoolInit)
	defer poolUpdate.Stop()
	task := time.NewTicker(c.cfg.Tickers.TaskCheck)
	defer task.Stop()

	snm := watchers.NewSNMPriceWatcher()
	token := watchers.NewTokenPriceWatcher()

	// add optional configure for each token (soon XMR)
	var reportedPool watchers.PoolWatcher
	var avgPool watchers.PoolWatcher

	switch c.cfg.MiningToken {
	case "ETH":
		reportedPool = watchers.NewPoolWatcher(poolReportedHashrateURL, []string{c.cfg.Pool.PoolAccount})
		avgPool = watchers.NewPoolWatcher(poolAverageHashrateURL, []string{c.cfg.Pool.PoolAccount + "/1"})
	case "ZEC":
		reportedPool = watchers.NewPoolWatcher(poolZECReportedHashrateURL, []string{c.cfg.Pool.PoolAccount})
		avgPool = watchers.NewPoolWatcher(poolZECAverageHashrateURL, []string{c.cfg.Pool.PoolAccount + "/1"})
	}

	if err := snm.Update(ctx); err != nil {
		return fmt.Errorf("cannot update snm data: %v", err)
	}
	if err := token.Update(ctx); err != nil {
		return fmt.Errorf("cannot update token data: %v", err)
	}

	profitModule := NewProfitableModules(c)
	poolModule := NewPoolModules(c)
	traderModule := NewTraderModules(c, poolModule, profitModule)

	md, ctx := errgroup.WithContext(ctx)
	//md.Go(func() error {
	//	// TODO(sshaman1101): this goroutine looks weird.
	//	return traderModule.ChargeOrdersOnce(ctx, token, snm, balanceReply)
	//})

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done %v", ctx.Err())
		case <-dataUpdate.C:
			go snm.Update(ctx)
			go token.Update(ctx)
		case <-tradeUpdate.C:
			tradeLog := c.logger.With(zap.String("subsystem", "trade"))
			if err := traderModule.SaveNewActiveDealsIntoDB(ctx); err != nil {
				tradeLog.Warn("cannot save active deals", zap.Error(err))
				continue
			}

			_, pricePerSec, err := traderModule.GetPriceForTokenPerSec(token)
			if err != nil {
				tradeLog.Warn("cannot get pricePerSec for token", zap.Error(err))
				continue
			}

			actualPrice := traderModule.FloatToBigInt(pricePerSec * c.cfg.Trade.MarginAccounting)
			if actualPrice.Cmp(big.NewInt(0)) == 0 {
				return fmt.Errorf("actual price is zero")
			}
			c.logger.Info("actual price", zap.String("price", actualPrice.String()))

			if err := traderModule.DealsTrading(ctx, actualPrice); err != nil {
				tradeLog.Warn("DealsTrading failed", zap.Error(err))
				continue
			}

			md.Go(func() error {
				if err := traderModule.OrderTrading(ctx, actualPrice); err != nil {
					tradeLog.Warn("OrderTrading failed", zap.Error(err))
				}
				return nil
			})

		case <-task.C:
			if err := poolModule.CheckTaskStatus(ctx); err != nil {
				c.logger.Warn("poolModule.CheckTaskStatus failed", zap.Error(err))
			}

		case <-poolUpdate.C:
			poolLog := c.logger.With(zap.String("subsystem", "pool"))

			deals, err := traderModule.c.db.GetDealsFromDB()
			if err != nil {
				poolLog.Warn("cannot get deals from DB", zap.Error(err))
				continue
			}

			for _, deal := range deals {
				if deal.DeployStatus == DeployStatusDeployed {

					marketDeal, err := c.DealClient.Status(ctx, sonm.NewBigIntFromInt(deal.DealID))
					if err != nil {
						poolLog.Warn("cannot get deal from market", zap.Int64("deal_ID", deal.DealID), zap.Error(err))
						continue
					}

					if marketDeal.Deal.Status == sonm.DealStatus_DEAL_CLOSED {
						id, err := strconv.Atoi(marketDeal.Deal.Id.String())
						if err != nil {
							return err
						}
						if err := c.db.UpdateBadGayStatusInPoolDB(int64(id), numberOfLives, time.Now()); err != nil {
							return err
						}
						continue
					}
					if err := poolModule.AddWorkerToPoolDB(ctx, marketDeal, c.cfg.Pool.PoolAccount); nil != err {
						return fmt.Errorf("cannot add worker to Db: %v", err)
					}
				}
			}

			if err := poolModule.AdvancedPoolHashrateTracking(ctx, reportedPool, avgPool); err != nil {
				c.logger.Warn("AdvancedPoolHashrateTracking failed", zap.Error(err))
				continue
			}
		}
	}
}

func (c *Connor) clearStart() error {
	if err := c.db.CreateAllTables(); err != nil {
		return err
	}
	return nil
}

func newCredentials(ctx context.Context, key *ecdsa.PrivateKey) (credentials.TransportCredentials, error) {
	_, TLSConfig, err := util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		return nil, err
	}
	return util.NewTLS(TLSConfig), nil
}
