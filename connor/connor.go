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
	poolReportedHashRateURL = "https://api.nanopool.org/v1/eth/reportedhashrates/"
	poolAverageHashRateURL  = "https://api.nanopool.org/v1/eth/avghashrateworkers/"
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
	connor := &Connor{
		key: key,
		cfg: cfg,
	}

	creds, err := newCredentials(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("can't create TLS credentials: %v", err)
	}

	nodeCC, err := xgrpc.NewClient(ctx, cfg.Market.Endpoint, creds)
	if err != nil {
		return nil, fmt.Errorf("can't create node connection: %v", err)
	}

	connor.Market = sonm.NewMarketClient(nodeCC)
	connor.TaskClient = sonm.NewTaskManagementClient(nodeCC)
	connor.DealClient = sonm.NewDealManagementClient(nodeCC)
	connor.TokenClient = sonm.NewTokenManagementClient(nodeCC)
	connor.MasterClient = sonm.NewMasterManagementClient(nodeCC)

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

	c.logger.Info("balance",
		zap.String("live", balanceReply.GetLiveBalance().Unwrap().String()),
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

	reportedPool := watchers.NewPoolWatcher(poolReportedHashRateURL, []string{c.cfg.Pool.PoolAccount})
	avgPool := watchers.NewPoolWatcher(poolAverageHashRateURL, []string{c.cfg.Pool.PoolAccount + "/1"})

	if err := snm.Update(ctx); err != nil {
		return fmt.Errorf("cannot update snm data: %v", err)
	}
	if err := token.Update(ctx); err != nil {
		return fmt.Errorf("cannot update token data: %v", err)
	}

	profitModule := NewProfitableModules(c)
	poolModule := NewPoolModules(c)
	traderModule := NewTraderModules(c, poolModule, profitModule)

	md := errgroup.Group{}
	md.Go(func() error {
		// TODO(sshaman1101): this goroutine looks weird.
		return traderModule.ChargeOrdersOnce(ctx, token, snm, balanceReply)
	})
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done %v", ctx.Err())
		case <-dataUpdate.C:
			md.Go(func() error {
				return snm.Update(ctx)
			})
			md.Go(func() error {
				return token.Update(ctx)
			})
		case <-tradeUpdate.C:
			c.logger.Info("update new deals")

			if err := traderModule.SaveNewActiveDealsIntoDB(ctx); err != nil {
				return fmt.Errorf("cannot save active deals: %v", err)
			}

			_, pricePerSec, err := traderModule.GetPriceForTokenPerSec(token)
			if err != nil {
				return fmt.Errorf("cannot get pricePerSec for token per sec: %v", err)
			}

			actualPrice := traderModule.FloatToBigInt(pricePerSec * c.cfg.Trade.MarginAccounting)
			if actualPrice == big.NewInt(0) {
				return fmt.Errorf("actual price is 0")
			}

			if err := traderModule.DealsTrading(ctx, actualPrice); err != nil {
				return err
			}

			md.Go(func() error {
				return traderModule.OrderTrading(ctx, actualPrice)
			})

		case <-task.C:
			err := poolModule.CheckTaskStatus(ctx)
			if err != nil {
				return err
			}

		case <-poolUpdate.C:
			c.logger.Info("start poolUpdate module hashrate tracking")

			dealsDb, err := traderModule.c.db.GetDealsFromDB()
			if err != nil {
				return fmt.Errorf("cannot get deals from DB: %v", err)
			}

			for _, dealDb := range dealsDb {
				if dealDb.DeployStatus == DeployStatusDeployed {

					dealOnMarket, err := c.DealClient.Status(ctx, sonm.NewBigIntFromInt(dealDb.DealID))
					if err != nil {
						return fmt.Errorf("cannot get deal from market %v, %v", dealDb.DealID, err)
					}
					if dealOnMarket.Deal.Status == sonm.DealStatus_DEAL_CLOSED {
						id, err := strconv.Atoi(dealOnMarket.Deal.Id.String())
						if err != nil {
							return err
						}
						if err := c.db.UpdateBadGayStatusInPoolDB(int64(id), numberOfLives, time.Now()); err != nil {
							return err
						}
						continue
					}
					if err := poolModule.AddWorkerToPoolDB(ctx, dealOnMarket, c.cfg.Pool.PoolAccount); nil != err {
						return fmt.Errorf("cannot add worker to Db: %v", err)
					}
				}
			}
			if err := poolModule.AdvancedPoolHashrateTracking(ctx, reportedPool, avgPool); err != nil {
				return err
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
