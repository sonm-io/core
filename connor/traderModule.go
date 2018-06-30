package connor

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/params"
	"github.com/sonm-io/core/connor/database"
	"github.com/sonm-io/core/connor/watchers"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

const (
	hashes       = 1000000
	daysPerMonth = 30
	secsPerDay   = 86400

	hashingPower     = 1
	costPerkWh       = 0.0
	powerConsumption = 0.0
)

type TraderModule struct {
	c      *Connor
	pool   *PoolModule
	profit *ProfitableModule
}

func NewTraderModules(c *Connor, pool *PoolModule, profit *ProfitableModule) *TraderModule {
	return &TraderModule{
		c:      c,
		pool:   pool,
		profit: profit,
	}
}

type DeployStatus int32

const (
	DeployStatusDEPLOYED    DeployStatus = 3
	DeployStatusNOTDEPLOYED DeployStatus = 4
	DeployStatusDESTROYED   DeployStatus = 5
)

type OrderStatus int32

const (
	OrderStatusCANCELLED OrderStatus = 3
	OrderStatusREINVOICE OrderStatus = 4
)

func (t *TraderModule) getTokenConfiguration(symbol string, cfg *Config) (float64, float64, float64) {
	switch symbol {
	case "ETH":
		return cfg.ChargeIntervalETH.Start, cfg.ChargeIntervalETH.Destination, cfg.Distances.StepForETH
	case "ZEC":
		return cfg.ChargeIntervalZEC.Start, cfg.ChargeIntervalZEC.Destination, cfg.Distances.StepForZEC
	case "XMR":
		return cfg.ChargeIntervalXMR.Start, cfg.ChargeIntervalXMR.Destination, cfg.Distances.StepForXMR
	}
	return 0, 0, 0
}

func (t *TraderModule) ChargeOrdersOnce(ctx context.Context, symbol string, token watchers.TokenWatcher, snm watchers.PriceWatcher, balanceReply *sonm.BalanceReply) error {
	if err := t.c.db.CreateOrderDB(); err != nil {
		return err
	}
	start, destination, step := t.getTokenConfiguration(symbol, t.c.cfg)

	count, err := t.c.db.GetCountFromDB()
	if err != nil {
		return fmt.Errorf("cannot get count from database %v", err)
	}

	if count == 0 {
		t.c.logger.Info("Save TEST order cause DB is empty!")
		if err := t.c.db.SaveOrderIntoDB(&database.OrderDb{
			OrderID:         0,
			Price:           0,
			Hashrate:        0,
			StartTime:       time.Time{},
			ButterflyEffect: 4,
			ActualStep:      start,
		}); err != nil {
			return fmt.Errorf("Cannot save order into DB %v\r\n", err)
		}
	}

	pricePerMonthUSD, pricePerSecMh, err := t.GetPriceForTokenPerSec(token, symbol)
	if err != nil {
		t.c.logger.Error("cannot get profit for tokens", zap.Error(err))
		return err
	}

	limitChargeInSNM := t.profit.LimitChargeSNM(balanceReply.GetSideBalance().Unwrap(), t.c.cfg.Sensitivity.PartCharge)
	limitChargeInSNMClone := big.NewInt(0).Set(limitChargeInSNM)
	limitChargeInUSD := t.profit.ConvertingToUSDBalance(limitChargeInSNMClone, snm.GetPrice())

	//limiter := 100.0 //сумма цен всех выставленных ордеров

	mhashForToken, err := t.c.db.GetLastActualStepFromDb()
	if err != nil {
		t.c.logger.Error("cannot get last actual step from DB", zap.Error(err))
		return err
	}

	pricePackMhInUSDPerMonth := mhashForToken * (pricePerMonthUSD * t.c.cfg.Sensitivity.MarginAccounting)
	sumOrdersPerMonth := limitChargeInUSD / pricePackMhInUSDPerMonth
	if sumOrdersPerMonth == 0 {
		return err
	}

	if limitChargeInSNM.Cmp(big.NewInt(0)) <= -1 {
		t.c.logger.Error("balance SNM is not enough for create orders!", zap.Error(err))
		return err
	}

	t.c.logger.Info("Start charge orders", zap.String("symbol ", symbol),
		zap.String("limit for charge SNM :", sonm.NewBigInt(limitChargeInSNM).ToPriceString()),
		zap.Float64("limit for charge USD :", limitChargeInUSD),
		zap.Float64("pack price per month USD", pricePackMhInUSDPerMonth),
		zap.Int64("sum orders per month", int64(sumOrdersPerMonth)),
		zap.Float64("step", step),
		zap.Int64("start", int64(start)),
		zap.Int64("destination", int64(destination)))

	for i := 0; i < int(sumOrdersPerMonth); i++ {
		if mhashForToken >= destination {
			t.c.logger.Info("charge is finished cause reached the limit", zap.Float64("charge destination", t.c.cfg.ChargeIntervalETH.Destination))
			break
		}
		pricePerSecPack := t.FloatToBigInt(mhashForToken * pricePerSecMh)
		zap.String("price", sonm.NewBigInt(pricePerSecPack).ToPriceString())
		mhashForToken, err = t.ChargeOrders(ctx, symbol, pricePerSecPack, step, mhashForToken)
		if err != nil {
			return fmt.Errorf("cannot charging market! %v\r\n", err)
		}
	}
	return nil
}

// Prepare price and Map depends on token symbol. Create orders to the market, until the budget is over.
func (t *TraderModule) ChargeOrders(ctx context.Context, symbol string, priceForHashPerSec *big.Int, step float64, buyMghash float64) (float64, error) {
	requiredHashRate := uint64(buyMghash * hashes)
	benchmarks, err := t.getBenchmarksForSymbol(symbol, uint64(requiredHashRate))
	if err != nil {
		return 0, err
	}
	buyMghash, err = t.CreateOrderOnMarketStep(ctx, step, benchmarks, buyMghash, priceForHashPerSec)
	if err != nil {
		return 0, err
	}
	return buyMghash, nil
}

// Create order on market depends on token.
func (t *TraderModule) CreateOrderOnMarketStep(ctx context.Context, step float64, benchmarks map[string]uint64, buyMgHash float64, price *big.Int) (float64, error) {
	actOrder, err := t.c.Market.CreateOrder(ctx, &sonm.BidOrder{
		Tag:      "Connor bot",
		Duration: &sonm.Duration{},
		Price: &sonm.Price{
			PerSecond: sonm.NewBigInt(price),
		},
		Identity: t.c.cfg.OtherParameters.IdentityForBid,
		Resources: &sonm.BidResources{
			Benchmarks: benchmarks,
			Network: &sonm.BidNetwork{
				Overlay:  false,
				Outbound: true,
				Incoming: false,
			},
		},
	})
	if err != nil {
		t.c.logger.Warn("cannot create bidOrder:", zap.String(err.Error(), "w"))
		return 0, err
	}
	if actOrder.GetId() != nil && actOrder.GetPrice() != nil {
		if err := t.c.db.SaveOrderIntoDB(&database.OrderDb{
			OrderID:         actOrder.GetId().Unwrap().Int64(),
			Price:           actOrder.GetPrice().Unwrap().Int64(),
			Hashrate:        actOrder.GetBenchmarks().GPUEthHashrate(),
			StartTime:       time.Now(),
			ButterflyEffect: int64(actOrder.GetOrderStatus()),
			ActualStep:      buyMgHash,
		}); err != nil {
			return 0, fmt.Errorf("cannot save order to database: %v", err)
		}
		t.c.logger.Info("order created",
			zap.Int64("id", actOrder.GetId().Unwrap().Int64()),
			zap.String("price", sonm.NewBigInt(actOrder.GetPrice().Unwrap()).ToPriceString()),
			zap.Uint64("hashrate", actOrder.GetBenchmarks().GPUEthHashrate()),
			zap.Float64("currently pack (Mh/s)", buyMgHash))
		reBuyHash := buyMgHash + step
		return reBuyHash, nil
	}
	return buyMgHash, nil
}

func (t *TraderModule) GetProfitForTokenBySymbol(tokens []*TokenMainData, symbol string) (float64, error) {
	for _, t := range tokens {
		if t.Symbol == symbol {
			return t.ProfitPerMonthUsd, nil
		}
	}
	return 0, fmt.Errorf("cannot get price from token! ")
}

func (t *TraderModule) GetPriceForTokenPerSec(token watchers.TokenWatcher, symbol string) (float64, float64, error) {
	tokens, err := t.profit.CollectTokensMiningProfit(token)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot calculate token prices: %v\r\n", err)
	}
	pricePerMonthUSD, err := t.GetProfitForTokenBySymbol(tokens, symbol)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get profit for tokens: %v\r\n", err)
	}
	pricePerSec := pricePerMonthUSD / (secsPerDay * daysPerMonth)
	return pricePerMonthUSD, pricePerSec, nil
}

func (t *TraderModule) ReinvoiceOrder(ctx context.Context, cfg *Config, price *sonm.Price, bench map[string]uint64, tag string) error {
	order, err := t.c.Market.CreateOrder(ctx, &sonm.BidOrder{
		Duration: &sonm.Duration{Nanoseconds: 0},
		Price:    price,
		Tag:      tag,
		Identity: cfg.OtherParameters.IdentityForBid,
		Resources: &sonm.BidResources{
			Network: &sonm.BidNetwork{
				Overlay:  false,
				Outbound: true,
				Incoming: false,
			},
			Benchmarks: bench,
		},
	})
	if err != nil {
		t.c.logger.Error("cannot created Lucky Order", zap.Error(err))
		return err
	}
	if err := t.c.db.SaveOrderIntoDB(&database.OrderDb{
		OrderID:         order.GetId().Unwrap().Int64(),
		Price:           order.GetPrice().Unwrap().Int64(),
		Hashrate:        order.GetBenchmarks().GPUEthHashrate(),
		StartTime:       time.Now(),
		ButterflyEffect: int64(OrderStatusREINVOICE),
		ActualStep:      0,
	}); err != nil {
		return fmt.Errorf("cannot save reinvoice order %s to DB: %v \r\n", order.GetId().Unwrap().String(), err)
	}
	t.c.logger.Info("reinvoice order",
		zap.String("new order", order.Id.Unwrap().String()),
		zap.String("tag", tag),
		zap.String("price", sonm.NewBigInt(order.Price.Unwrap()).ToPriceString()),
		zap.Uint64("hashrate", order.GetBenchmarks().GPUEthHashrate()))
	return nil
}

func (t *TraderModule) cmpChangeOfPrice(change float64, def float64) (int32, error) {
	if change >= 100+def {
		return 1, nil
	} else if change < 100-def {
		return -1, nil
	}
	return 0, nil
}

func (t *TraderModule) OrdersProfitTracking(ctx context.Context, cfg *Config, actualPrice *big.Int, orderDb *database.OrderDb) error {
	order, err := t.c.Market.GetOrderByID(ctx, &sonm.ID{Id: strconv.Itoa(int(orderDb.OrderID))})
	if err != nil {
		t.c.logger.Error("cannot get order from market", zap.Int64("order Id", orderDb.OrderID), zap.Error(err))
		return err
	}

	switch order.GetOrderStatus() {
	case sonm.OrderStatus_ORDER_ACTIVE:
		pack := int64(order.GetBenchmarks().GPUEthHashrate()) / hashes
		pricePerSecForPack := big.NewInt(0).Mul(actualPrice, big.NewInt(pack))
		if pricePerSecForPack == big.NewInt(0) {
			t.c.logger.Info("actual price %v", zap.String("price", sonm.NewBigInt(actualPrice).ToPriceString()))
			return nil
		}

		div := big.NewInt(0).Div(big.NewInt(0).Mul(pricePerSecForPack, big.NewInt(100)), order.Price.Unwrap())
		if err != nil {
			return fmt.Errorf("cannot get change percent from deal: %v", err)
		}
		changePricePercent, _ := big.NewFloat(0).SetInt64(div.Int64()).Float64()
		if changePricePercent == 0 {
			return nil
		}
		sensitivity := t.c.cfg.Sensitivity.OrdersChangePercent

		if changePricePercent > 100+sensitivity {
			t.c.logger.Info("High price ==>  create reinvoice order",
				zap.String("active orderID", order.Id.Unwrap().String()),
				zap.String("order price", sonm.NewBigInt(order.Price.Unwrap()).ToPriceString()),
				zap.String("actual price for pack", sonm.NewBigInt(pricePerSecForPack).ToPriceString()),
				zap.String("actual price per second", sonm.NewBigInt(actualPrice).ToPriceString()),
				zap.Float64("change percent", changePricePercent),
				zap.Float64("sensitivity", 100+sensitivity))
			bench, err := t.GetBidBenchmarks(order)
			if err != nil {
				return fmt.Errorf("Cannot get benchmarks from Order : %v\r\n", order.Id.Unwrap().Int64())
			}
			tag := strconv.Itoa(int(orderDb.OrderID))
			if err := t.ReinvoiceOrder(ctx, cfg, &sonm.Price{PerSecond: sonm.NewBigInt(pricePerSecForPack)}, bench, "Reinvoice(update price): "+tag); err != nil {
				return err
			}
			_, err = t.c.Market.CancelOrder(ctx, &sonm.ID{Id: strconv.Itoa(int(orderDb.OrderID))})
			if err != nil {
				return err
			}
		} else if changePricePercent < 100-sensitivity {
			t.c.logger.Info("Low price ==> create reinvoice order",
				zap.String("active order id", order.Id.Unwrap().String()),
				zap.String("order price", sonm.NewBigInt(order.Price.Unwrap()).ToPriceString()),
				zap.String("actual price for pack", sonm.NewBigInt(pricePerSecForPack).ToPriceString()),
				zap.String("actual price per second", sonm.NewBigInt(actualPrice).ToPriceString()),
				zap.Float64("change percent", changePricePercent),
				zap.Float64("sensitivity", 100-sensitivity))
			bench, err := t.GetBidBenchmarks(order)
			if err != nil {
				return fmt.Errorf("Cannot get benchmarks from Order : %v\r\n", order.Id.Unwrap().Int64())
			}
			tag := strconv.Itoa(int(orderDb.OrderID))
			_, err = t.c.Market.CancelOrder(ctx, &sonm.ID{Id: strconv.Itoa(int(orderDb.OrderID))})
			if err != nil {
				return err
			}
			if err := t.ReinvoiceOrder(ctx, cfg, &sonm.Price{PerSecond: sonm.NewBigInt(pricePerSecForPack)}, bench, "Reinvoice(update price): "+tag); err != nil {
				return err
			}
		}
	case sonm.OrderStatus_ORDER_INACTIVE:
		t.c.logger.Info("order is not active", zap.String("Id", order.Id.Unwrap().String()))
		t.c.db.UpdateOrderInDB(orderDb.OrderID, int64(OrderStatusCANCELLED))
	}
	return nil
}

//Deploy new container + reinvoice order
func (t *TraderModule) ResponseToActiveDeals(ctx context.Context, dealDb *database.DealDb, image string) error {
	dealOnMarket, err := t.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(dealDb.DealID))
	if err != nil {
		return fmt.Errorf("cannot get deal info %v\r\n", err)
	}

	if dealOnMarket.Deal.Status == sonm.DealStatus_DEAL_CLOSED {
		t.c.db.UpdateDealStatusDb(dealDb.DealID, sonm.DealStatus_DEAL_CLOSED)
		return nil
	}

	t.c.logger.Info("Deploying new container",
		zap.String("deal id", dealOnMarket.Deal.Id.String()))
	task, err := t.pool.DeployNewContainer(ctx, t.c.cfg, dealOnMarket.Deal, image)
	if err != nil {
		t.c.db.UpdateDeployStatusDealInDB(dealOnMarket.Deal.Id.Unwrap().Int64(), int64(DeployStatusNOTDEPLOYED))
		t.c.logger.Error("cannot deploy new container from task %s\r\n", zap.Error(err))
		return err
	}

	if err := t.c.db.UpdateDeployStatusDealInDB(dealOnMarket.Deal.Id.Unwrap().Int64(), int64(DeployStatusDEPLOYED)); err != nil {
		return err
	}

	t.c.logger.Info("New deployed task", zap.String("task", task.GetId()), zap.String("deal", dealOnMarket.Deal.GetId().String()))
	bidOrder, err := t.c.Market.GetOrderByID(ctx, &sonm.ID{Id: dealOnMarket.Deal.GetBidID().String()})
	if err != nil {
		return fmt.Errorf("cannot get order by Id: %v\r\n", err)
	}
	bench, err := t.GetBidBenchmarks(bidOrder)
	if err != nil {
		return fmt.Errorf("cannot get benchmarks from bid Order : %v\r\n", bidOrder.Id.Unwrap().Int64())
	}
	if err := t.ReinvoiceOrder(ctx, t.c.cfg, &sonm.Price{PerSecond: dealOnMarket.Deal.GetPrice()}, bench, "Reinvoice(active deal)"); err != nil {
		return fmt.Errorf("cannot reinvoice order %v", err)
	}
	return nil
}

//TODO: not closed not deployed deals
//Compare deal price and new price. If the price went down create change request and get response.
func (t *TraderModule) DeployedDealsProfitTrack(ctx context.Context, actualPrice *big.Int, dealDb *database.DealDb, image string) error {
	dealOnMarket, err := t.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(dealDb.DealID))
	if err != nil {
		return fmt.Errorf("cannot get deal info %v\r\n", err)
	}
	bidOrder, err := t.c.Market.GetOrderByID(ctx, &sonm.ID{Id: dealOnMarket.Deal.BidID.Unwrap().String()})
	if err != nil {
		return err
	}

	pack := float64(bidOrder.Benchmarks.GPUEthHashrate()) / float64(hashes)
	actualPriceForPack := big.NewInt(0).Mul(actualPrice, big.NewInt(int64(pack)))

	dealPrice := dealOnMarket.Deal.Price.Unwrap()
	div := big.NewInt(0).Div(big.NewInt(0).Mul(actualPriceForPack, big.NewInt(100)), dealPrice)
	if err != nil {
		return fmt.Errorf("cannot get change percent from deal: %v", err)
	}
	changePricePercent, _ := big.NewFloat(0).SetInt64(div.Int64()).Float64()
	sensitivity := t.c.cfg.Sensitivity.DealsChangePercent

	if actualPriceForPack.IsInt64() == false {
		return fmt.Errorf("actual price overflows int64")
	}
	changeRequestStatus, err := t.c.db.GetChangeRequestStatus(108);
	if err != nil {
		return err
	}
	if changeRequestStatus == 1 {
		return nil
	}

	if changePricePercent > 100+sensitivity {
		dealChangeRequest, err := t.CreateChangeRequest(ctx, dealOnMarket, actualPriceForPack)
		if err != nil {
			return fmt.Errorf("cannot create change request %v\r\n", err)
		}
		t.c.logger.Info("High percent ==> create deal change request ",
			zap.String("high change request", dealChangeRequest.Unwrap().String()),
			zap.String("active deal id", dealOnMarket.Deal.Id.Unwrap().String()),
			zap.String("deal price", sonm.NewBigInt(dealPrice).ToPriceString()),
			zap.String("actual price for pack", sonm.NewBigInt(actualPriceForPack).ToPriceString()),
			zap.String("actual price per second", sonm.NewBigInt(actualPrice).ToPriceString()),
			zap.Float64("change percent", changePricePercent),
			zap.Float64("sensitivity", 100+sensitivity))
		log.Printf("new price for deal %v", actualPriceForPack.Int64())
		if err := t.c.db.UpdateChangeRequestStatusDealDB(dealDb.DealID, sonm.ChangeRequestStatus_REQUEST_CREATED, actualPriceForPack.Int64()); err != nil {
			return err
		}
	} else if changePricePercent < 100-sensitivity {
		dealChangeRequest, err := t.CreateChangeRequest(ctx, dealOnMarket, actualPriceForPack)
		if err != nil {
			return fmt.Errorf("cannot create change request %v\r\n", err)
		}
		t.c.logger.Info("Low price ==> create deal change request ",
			zap.String("low change request", dealChangeRequest.Unwrap().String()),
			zap.String("active deal id", dealOnMarket.Deal.Id.Unwrap().String()),
			zap.String("deal price", sonm.NewBigInt(dealPrice).ToPriceString()),
			zap.String("actual price for pack", sonm.NewBigInt(actualPriceForPack).ToPriceString()),
			zap.String("actual price per second", sonm.NewBigInt(actualPrice).ToPriceString()),
			zap.Float64("change percent", changePricePercent),
			zap.Float64("sensitivity", 100-sensitivity))
		if err := t.c.db.UpdateChangeRequestStatusDealDB(dealDb.DealID, sonm.ChangeRequestStatus_REQUEST_CREATED, actualPriceForPack.Int64()); err != nil {
			return err
		}
		go t.GetChangeRequest(ctx, dealOnMarket) // TODO: wait for the go-routine to finish.
	}
	return nil
}

func (t *TraderModule) CreateChangeRequest(ctx context.Context, dealOnMarket *sonm.DealInfoReply, actualPriceForPack *big.Int) (*sonm.BigInt, error) {
	dealChangeRequest, err := t.c.DealClient.CreateChangeRequest(ctx, &sonm.DealChangeRequest{
		Id:          nil,
		DealID:      dealOnMarket.Deal.Id,
		RequestType: sonm.OrderType_BID,
		Duration:    0,
		Price:       sonm.NewBigIntFromInt(actualPriceForPack.Int64()),
		Status:      sonm.ChangeRequestStatus_REQUEST_CREATED,
		CreatedTS:   nil,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create change request %v\r\n", err)
	}
	return dealChangeRequest, nil
}

func (t *TraderModule) GetChangeRequest(ctx context.Context, dealChangeRequest *sonm.DealInfoReply) error {
	//time.Sleep(time.Duration(t.c.cfg.Sensitivity.WaitingTimeCRSec))
	fmt.Printf("wait responce by change request id %v", dealChangeRequest.Deal.Id.Unwrap().Int64())

	time.Sleep(time.Duration(900 * time.Second))

	requestsList, err := t.c.DealClient.ChangeRequestsList(ctx, dealChangeRequest.Deal.Id)
	if err != nil {
		return err
	}
	for _, cr := range requestsList.Requests {
		if cr.DealID == dealChangeRequest.Deal.Id {
			if cr.Status == sonm.ChangeRequestStatus_REQUEST_ACCEPTED {
				continue
			} else {
				t.c.DealClient.Finish(ctx, &sonm.DealFinishRequest{
					Id: dealChangeRequest.Deal.Id,
				})
				if err := t.c.db.ReturnChangeRequestStatusDealDB(dealChangeRequest.Deal.Id.Unwrap().Int64(), sonm.ChangeRequestStatus_REQUEST_UNKNOWN); err != nil {
					return err
				}
				t.c.logger.Info("worker didn't accepted change request",
					zap.String("deal", dealChangeRequest.Deal.Id.Unwrap().String()))
			}
		}
	}
	return nil

}

func (t *TraderModule) SaveNewActiveDealsIntoDB(ctx context.Context) error {
	getDeals, err := t.c.DealClient.List(ctx, &sonm.Count{Count: 100})
	if err != nil {
		return fmt.Errorf("Cannot get Deals list %v\r\n", err)
	}
	deals := getDeals.Deal
	if len(deals) < 0 {
		t.c.logger.Info("No active deals")
		return nil
	}

	for _, deal := range deals {

		// todo: странная логика сохранения - сохранять не закрытые сделки
		if deal.GetStatus() != sonm.DealStatus_DEAL_CLOSED {
			t.c.db.SaveDealIntoDB(&database.DealDb{
				DealID:       deal.GetId().Unwrap().Int64(),
				Status:       int64(deal.GetStatus()),
				Price:        deal.GetPrice().Unwrap().Int64(),
				AskID:        deal.GetAskID().Unwrap().Int64(),
				BidID:        deal.GetBidID().Unwrap().Int64(),
				DeployStatus: int64(DeployStatusNOTDEPLOYED),
			})
		}
	}

	return nil
}
func (t *TraderModule) UpdateDealsIntoDb(ctx context.Context) error {
	dealsDb, err := t.c.db.GetDealsFromDB()
	if err != nil {
		return fmt.Errorf("cannot update deals into database %v ", err)
	}
	for _, d := range dealsDb {
		getDeal, err := t.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(d.DealID))
		if err != nil {
			return err
		}
		t.c.db.UpdateDealStatusDb(d.DealID, getDeal.Deal.Status)
	}
	//TODO: update price
	return nil
}

func (t *TraderModule) GetDeployedDeals() ([]int64, error) {
	dealsDB, err := t.c.db.GetDealsFromDB()
	if err != nil {
		return nil, fmt.Errorf("cannot create benchmarkes for symbol \"%s\"", err)
	}
	deployedDeals := make([]int64, 0)
	for _, d := range dealsDB {
		if d.DeployStatus == int64(DeployStatusDEPLOYED) {
			deal := d.DealID
			deployedDeals = append(deployedDeals, deal)
		}
	}
	return deployedDeals, nil
}
func (t *TraderModule) GetBidBenchmarks(bidOrder *sonm.Order) (map[string]uint64, error) {
	getBench := bidOrder.GetBenchmarks()
	bMap := map[string]uint64{
		"ram-size":            getBench.RAMSize(),
		"cpu-cores":           getBench.CPUCores(),
		"cpu-sysbench-single": getBench.CPUSysbenchOne(),
		"cpu-sysbench-multi":  getBench.CPUSysbenchMulti(),
		"net-download":        getBench.NetTrafficIn(),
		"net-upload":          getBench.NetTrafficOut(),
		"gpu-count":           getBench.GPUCount(),
		"gpu-mem":             getBench.GPUMem(),
		"gpu-eth-hashrate":    getBench.GPUEthHashrate(),
	}
	return bMap, nil
}
func (t *TraderModule) FloatToBigInt(val float64) *big.Int {
	realV := big.NewFloat(val)
	price := realV.Mul(realV, big.NewFloat(params.Ether))
	v, _ := price.Int(nil)
	return v
}

func (t *TraderModule) newBaseBenchmarks() map[string]uint64 {
	return map[string]uint64{
		"ram-size":            t.c.cfg.Benchmark.RamSize,
		"cpu-cores":           t.c.cfg.Benchmark.CpuCores,
		"cpu-sysbench-single": t.c.cfg.Benchmark.CpuSysbenchSingle,
		"cpu-sysbench-multi":  t.c.cfg.Benchmark.CpuSysbenchMulti,
		"net-download":        t.c.cfg.Benchmark.NetDownload,
		"net-upload":          t.c.cfg.Benchmark.NetUpload,
		"gpu-count":           t.c.cfg.Benchmark.GpuCount,
		"gpu-mem":             t.c.cfg.Benchmark.GpuMem,
	}
}
func (t *TraderModule) newBenchmarksWithGPU(ethHashRate uint64) map[string]uint64 {
	b := t.newBaseBenchmarks()
	b["gpu-eth-hashrate"] = ethHashRate
	return b
}
func (t *TraderModule) newBenchmarksWithoutGPU() map[string]uint64 {
	return t.newBaseBenchmarks()
}
func (t *TraderModule) getBenchmarksForSymbol(symbol string, ethHashRate uint64) (map[string]uint64, error) {
	switch symbol {
	case "ETH":
		return t.newBenchmarksWithGPU(ethHashRate), nil
	case "ZEC":
		return t.newBenchmarksWithoutGPU(), nil
	case "XMR":
		return t.newBenchmarksWithGPU(ethHashRate), nil
	default:
		return nil, fmt.Errorf("cannot create benchmakes for symbol \"%s\"", symbol)
	}
}

func (t *TraderModule) CancelAllOrders(ctx context.Context) error {
	orders, err := t.c.db.GetOrdersFromDB()
	if err != nil {
		return fmt.Errorf("cannot get orders from DB %v\r\n", err)
	}
	for _, o := range orders {
		if o.ActualStep > 160 {
			t.c.Market.CancelOrder(ctx, &sonm.ID{
				Id: strconv.Itoa(int(o.OrderID))})
		}
		t.c.logger.Info("cancelled", zap.Int64("id", o.OrderID))
	}
	return nil
}
