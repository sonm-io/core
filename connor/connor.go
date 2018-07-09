package connor

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/credentials"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/connor/database"
	"github.com/sonm-io/core/connor/watchers"
	"go.uber.org/zap"
)

const (
	coinMarketCapTicker     = "https://api.coinmarketcap.com/v1/ticker/"
	coinMarketCapSonmTicker = coinMarketCapTicker + "sonm/"
	cryptoCompareCoinData   = "https://www.cryptocompare.com/api/data/coinsnapshotfullbyid/?id="
	poolReportedHashRate    = "https://api.nanopool.org/v1/eth/reportedhashrates/"
	poolAverageHashRate     = "https://api.nanopool.org/v1/eth/avghashrateworkers/"
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
		return nil, fmt.Errorf("can't create node connection: %v\r\n", err)
	}

	connor.Market = sonm.NewMarketClient(nodeCC)
	connor.TaskClient = sonm.NewTaskManagementClient(nodeCC)
	connor.DealClient = sonm.NewDealManagementClient(nodeCC)
	connor.TokenClient = sonm.NewTokenManagementClient(nodeCC)
	connor.MasterClient = sonm.NewMasterManagementClient(nodeCC)

	connor.db, err = database.NewDatabaseConnect(connor.cfg.Database.Driver, connor.cfg.Database.DataSource)
	if err != nil {
		return nil, err
	}

	balanceReply, err := connor.TokenClient.Balance(ctx, &sonm.Empty{})
	if err != nil {
		connor.logger.Error("cannot load balanceReply", zap.Error(err))
		return nil, err
	}

	connor.logger = ctxlog.GetLogger(ctx)
	connor.logger.Info("balance",
		zap.String("live", balanceReply.GetLiveBalance().Unwrap().String()),
		zap.String("Side", balanceReply.GetSideBalance().ToPriceString()))
	connor.logger.Info("configuring connor", zap.Any("config", cfg))
	return connor, nil
}

func (c *Connor) Serve(ctx context.Context) error {
	c.logger.Info("Connor started work ...")
	defer c.logger.Info("Connor has been stopped")

	c.ClearStart()

	dataUpdate := time.NewTicker(time.Duration(c.cfg.Tickers.DataUpdate) * time.Second)
	defer dataUpdate.Stop()
	tradeUpdate := time.NewTicker(time.Duration(c.cfg.Tickers.TradeTicker) * time.Second)
	defer tradeUpdate.Stop()
	poolInit := time.NewTicker(time.Duration(c.cfg.Tickers.PoolInit) * time.Second)
	defer poolInit.Stop()

	snm := watchers.NewSNMPriceWatcher(coinMarketCapSonmTicker)
	token := watchers.NewTokenPriceWatcher(coinMarketCapTicker, cryptoCompareCoinData)
	reportedPool := watchers.NewPoolWatcher(poolReportedHashRate, []string{c.cfg.PoolAddress.EthPoolAddr})
	avgPool := watchers.NewPoolWatcher(poolAverageHashRate, []string{c.cfg.PoolAddress.EthPoolAddr + "/1"})

	// straight update of all watchers
	if err := snm.Update(ctx); err != nil {
		return fmt.Errorf("cannot update snm data: %v", err)
	}
	if err := token.Update(ctx); err != nil {
		return fmt.Errorf("cannot update token data: %v", err)
	}
	if err := reportedPool.Update(ctx); err != nil {
		return fmt.Errorf("cannot update reportedPool data: %v", err)
	}
	if err := avgPool.Update(ctx); err != nil {
		return fmt.Errorf("cannot update avgPool data: %v", err)
	}
	profitModule := NewProfitableModules(c)
	poolModule := NewPoolModules(c)
	traderModule := NewTraderModules(c, poolModule, profitModule)

	balanceReply, err := c.TokenClient.Balance(ctx, &sonm.Empty{})
	if err != nil {
		return err
	}

	md := errgroup.Group{}
	md.Go(func() error {
		return traderModule.ChargeOrdersOnce(ctx, c.cfg.UsingToken.Token, token, snm, balanceReply)
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
			md.Go(func() error {
				return reportedPool.Update(ctx)
			})
			md.Go(func() error {
				return avgPool.Update(ctx)
			})
		case <-tradeUpdate.C:
			c.logger.Info("start trade module hashrate tracking")

			err := traderModule.SaveNewActiveDealsIntoDB(ctx)
			if err != nil {
				return fmt.Errorf("cannot save active deals : %v", err)
			}

			if err := traderModule.UpdateDealsIntoDb(ctx); err != nil {
				return err
			}

			_, pricePerSec, err := traderModule.GetPriceForTokenPerSec(token, c.cfg.UsingToken.Token)
			if err != nil {
				return fmt.Errorf("cannot get pricePerSec for token per sec %v", err)
			}

			actualPrice := traderModule.FloatToBigInt(pricePerSec)
			if actualPrice == big.NewInt(0) {
				return fmt.Errorf("actual price is 0")
			}

			c.logger.Info("new actual price hashes per sec", zap.String("price", sonm.NewBigInt(actualPrice).ToPriceString()))

			dealsDb, err := traderModule.c.db.GetDealsFromDB()
			if err != nil {
				return fmt.Errorf("cannot get deals from DB %v\r\n", err)
			}

			for _, dealDb := range dealsDb {
				if dealDb.Status != int64(sonm.DealStatus_DEAL_CLOSED) {
					if dealDb.DeployStatus == int64(DeployStatusNOTDEPLOYED) {

						c.logger.Info("Response to active deals")

						if err := traderModule.ResponseToActiveDeals(ctx, dealDb, c.cfg.Images.Image); err != nil {
							return err
						}
					} else if dealDb.DeployStatus == int64(DeployStatusDEPLOYED) && dealDb.ChangeRequestStatus != int64(sonm.ChangeRequestStatus_REQUEST_CREATED) {
						if err := traderModule.DeployedDealsProfitTrack(ctx, actualPrice, dealDb, c.cfg.Images.Image); err != nil {
							return err
						}
					}
				}
			}

			orders, err := traderModule.c.db.GetOrdersFromDB()
			if err != nil {
				return fmt.Errorf("cannot get orders from DB %v\r\n", err)
			}
			for _, order := range orders {
				if order.ButterflyEffect != int64(OrderStatusCANCELLED) {
					err := traderModule.OrdersProfitTracking(ctx, c.cfg, actualPrice, order)
					if err != nil {
						return fmt.Errorf("cannot start orders profit tracking: %v", err)
					}
				}
			}
		case <-poolInit.C:
			c.logger.Info("start pool module hashrate tracking")

			dealsDb, err := traderModule.c.db.GetDealsFromDB()
			if err != nil {
				return fmt.Errorf("cannot get deals from DB %v\r\n", err)
			}

			for _, dealDb := range dealsDb {
				if dealDb.DeployStatus == int64(DeployStatusDEPLOYED) {

					dealOnMarket, err := c.DealClient.Status(ctx, sonm.NewBigIntFromInt(dealDb.DealID))
					if err != nil {
						return fmt.Errorf("cannot get deal from market %v", dealDb.DealID)
					}

					if err := poolModule.AddWorkerToPoolDB(ctx, dealOnMarket, c.cfg.PoolAddress.EthPoolAddr); err != nil {
						return fmt.Errorf("cannot add worker to Db : %v", err)
					}
				}
			}
			if err := poolModule.AdvancedPoolHashrateTracking(ctx, reportedPool, avgPool); err != nil {
				return err
			}
		}
	}
}

func (c *Connor) ClearStart() error {
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
