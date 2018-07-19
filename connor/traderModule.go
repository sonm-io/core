package connor

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/sonm-io/core/connor/database"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

const (
	hashes   = 1000000
	fullPath = 100
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

const (
	DeployStatusNotDeployed = 0
	DeployStatusDeployed    = 1
	DeployStatusDestroyed   = 2
)

const (
	OrderStatusCancelled = 3
	OrderStatusReinvoice = 4
)

func (t *TraderModule) SaveNewActiveDealsIntoDB(ctx context.Context) error {
	t.c.logger.Debug("SaveNewActiveDealsIntoDB")

	deals, err := t.c.DealClient.List(ctx, &sonm.Count{Count: 100})
	if err != nil {
		return fmt.Errorf("cannot get deals list: %v", err)
	}

	for _, deal := range deals.Deal {
		if deal.GetStatus() != sonm.DealStatus_DEAL_CLOSED {
			t.c.db.SaveDealIntoDB(&database.DealDB{
				DealID:       deal.Id.Unwrap().Int64(),
				Status:       int64(deal.Status),
				Price:        deal.Price.Unwrap().Int64(),
				AskID:        deal.AskID.Unwrap().Int64(),
				BidID:        deal.BidID.Unwrap().Int64(),
				DeployStatus: DeployStatusNotDeployed,
			})
		}
	}
	return nil
}

// DealsTrading makes a decision depending on the status of the deal.
// Not deployed: deploy new container || deployed: create change request if necessary.
func (t *TraderModule) DealsTrading(ctx context.Context, actualPrice *big.Int) error {
	log := t.c.logger.With(zap.String("module", "deals-trading"))

	log.Debug("started")
	defer log.Debug("finished")

	dealsDb, err := t.c.db.GetDealsFromDB()
	if err != nil {
		return fmt.Errorf("cannot get deals from database: %v", err)
	}

	for _, dealDB := range dealsDb {
		dealDeployStatus, err := t.c.db.GetDeployStatus(dealDB.DealID)
		if err != nil {
			log.Warn("cannot get deploy status of deal", zap.Error(err))
			continue
		}

		if dealDB.Status == int64(sonm.DealStatus_DEAL_ACCEPTED) {
			checkDealStatus, err := t.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(dealDB.DealID))
			if err != nil {
				log.Warn("cannot get deal status", zap.Error(err))
				continue
			}

			if checkDealStatus.Deal.Status == sonm.DealStatus_DEAL_CLOSED {
				log.Info("deal closed on market, task tracking stop")

				if err = t.c.db.UpdateDeployAndDealStatusDB(dealDB.DealID, DeployStatusDestroyed, sonm.DealStatus_DEAL_CLOSED); err != nil {
					log.Warn("cannot update deploy ans deal status", zap.Error(err))
				}

				continue
			}

			switch dealDeployStatus {
			case DeployStatusNotDeployed:
				if err := t.ResponseToActiveDeal(ctx, dealDB); err != nil {
					log.Warn("response to active deal failed", zap.Error(err))
					continue
				}
			case DeployStatusDeployed:
				if dealDB.ChangeRequestStatus != int64(sonm.ChangeRequestStatus_REQUEST_CREATED) {
					if err := t.deployedDealProfitTracking(ctx, actualPrice, dealDB); err != nil {
						log.Warn("deployed deal profit tracking failed", zap.Error(err))
						continue
					}
				}
			case DeployStatusDestroyed:
				log.Debug("deal is destroyed, nothing to do", zap.Any("deal_id", dealDB.DealID))
			}
		}
	}
	return nil
}

// ResponseToActiveDeal deploys new container and reinvoice order from deal.
func (t *TraderModule) ResponseToActiveDeal(ctx context.Context, dealDB *database.DealDB) error {
	dealOnMarket, err := t.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(dealDB.DealID))
	if err != nil {
		return fmt.Errorf("cannot get deal info: %v", err)
	}

	t.c.logger.Info("processing of deploying new container", zap.Any("deal_ID", dealOnMarket.Deal))
	newContainer, err := t.pool.DeployNewContainer(ctx, dealOnMarket.Deal, t.c.cfg.Pool.Image)
	if err != nil {
		t.c.logger.Warn("cannot start task", zap.Error(err))
		return err
	}

	if err := t.c.db.UpdateDeployStatusDealInDB(dealOnMarket.Deal.Id.Unwrap().Int64(), DeployStatusDeployed); err != nil {
		return err
	}

	t.c.logger.Info("new container deployed successfully", zap.Int64("deal", dealDB.DealID), zap.Any("container", newContainer))

	if err := t.ReinvoiceOrderFromDeal(ctx, dealOnMarket.Deal); err != nil {
		return err
	}

	return nil
}

// Compare deal price and new price. If the price went down/up create change request and get response.
func (t *TraderModule) deployedDealProfitTracking(ctx context.Context, actualPrice *big.Int, dealDB *database.DealDB) error {

	changeRequestStatus, err := t.c.db.GetChangeRequestStatus(dealDB.DealID)
	if err != nil {
		return err
	}

	// TODO(sshaman1101): WAT?
	if changeRequestStatus == 1 {
		return nil
	}

	dealOnMarket, err := t.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(dealDB.DealID))
	if err != nil {
		return fmt.Errorf("cannot get deal info: %v", err)
	}

	bidOrder, err := t.c.Market.GetOrderByID(ctx, &sonm.ID{Id: dealOnMarket.Deal.BidID.Unwrap().String()})
	if err != nil {
		return err
	}

	pack := float64(bidOrder.Benchmarks.GPUEthHashrate()) / float64(hashes)
	actualPriceForPack := big.NewInt(0).Mul(actualPrice, big.NewInt(int64(pack)))

	div := big.NewInt(0).Div(big.NewInt(0).Mul(actualPriceForPack, big.NewInt(100)), dealOnMarket.Deal.Price.Unwrap())

	if actualPriceForPack.IsInt64() == false {
		return fmt.Errorf("actual price overflows int64")
	}

	changePricePercent, _ := big.NewFloat(0).SetInt64(div.Int64()).Float64()

	if changePricePercent > fullPath+t.c.cfg.Trade.DealsChangePercent || changePricePercent < fullPath-t.c.cfg.Trade.DealsChangePercent {
		dealChangeRequest, err := t.CreateChangeRequest(ctx, dealOnMarket, actualPriceForPack)
		if err != nil {
			return fmt.Errorf("cannot create change request: %v", err)
		}
		t.c.logger.Info("change percent for deal ==> create deal change request ", zap.String("high_CR", dealChangeRequest.Unwrap().String()),
			zap.String("deal_ID", dealOnMarket.Deal.Id.Unwrap().String()), zap.String("deal_price", sonm.NewBigInt(dealOnMarket.Deal.Price.Unwrap()).ToPriceString()),
			zap.String("actual_price_for_pack", sonm.NewBigInt(actualPriceForPack).ToPriceString()), zap.Float64("change_percent", changePricePercent))

		if err := t.c.db.UpdateChangeRequestStatusDealDB(dealDB.DealID, sonm.ChangeRequestStatus_REQUEST_CREATED, actualPriceForPack.Int64()); err != nil {
			return err
		}
		go t.GetChangeRequest(ctx, dealOnMarket) // TODO: wait for the go-routine to finish.
	}
	return nil
}

// Create new order from deal (inherits benchmark from bid).
func (t *TraderModule) ReinvoiceOrderFromDeal(ctx context.Context, deal *sonm.Deal) error {
	bidOrder, err := t.c.Market.GetOrderByID(ctx, &sonm.ID{Id: deal.GetBidID().String()})
	if err != nil {
		return fmt.Errorf("cannot get order by ID from market: %v", err)
	}
	bench := t.GetBidBenchmarks(bidOrder)

	if err := t.ReinvoiceOrder(ctx, &sonm.Price{PerSecond: deal.GetPrice()}, bench, "Reinvoice(active deal)"); err != nil {
		return fmt.Errorf("cannot reinvoice order: %v", err)
	}
	return nil
}

// Create change request. Use deal change request with new price.
func (t *TraderModule) CreateChangeRequest(ctx context.Context, dealOnMarket *sonm.DealInfoReply, actualPriceForPack *big.Int) (*sonm.BigInt, error) {
	dealChangeRequest, err := t.c.DealClient.CreateChangeRequest(ctx, &sonm.DealChangeRequest{
		Id:          nil,
		DealID:      dealOnMarket.Deal.Id,
		RequestType: sonm.OrderType_BID,
		Duration:    0,
		Price:       sonm.NewBigIntFromInt(actualPriceForPack.Int64()),
		Status:      sonm.ChangeRequestStatus_REQUEST_CREATED,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create change request: %v", err)
	}
	return dealChangeRequest, nil
}

// OrderTrading makes a decision depending on the status of the order.
func (t *TraderModule) OrderTrading(ctx context.Context, actualPrice *big.Int) error {
	orders, err := t.c.db.GetOrdersFromDB()
	if err != nil {
		return fmt.Errorf("cannot get orders from DB: %v", err)
	}

	for _, order := range orders {
		if order.Status != OrderStatusCancelled {
			if err := t.ordersProfitTracking(ctx, actualPrice, order); err != nil {
				return fmt.Errorf("cannot start orders profit tracking: %v", err)
			}
		}
	}
	return nil
}

// ComparisonWithDealHashrate order price and new price. If the price went down/up order will reinvoice.
func (t *TraderModule) ordersProfitTracking(ctx context.Context, actualPrice *big.Int, orderDB *database.OrderDb) error {
	log := t.c.logger.With(zap.String("module", "order-profit-tracking"))

	log.Debug("started")
	defer log.Debug("finished")

	order, err := t.c.Market.GetOrderByID(ctx, &sonm.ID{Id: strconv.Itoa(int(orderDB.OrderID))})
	if err != nil {
		return fmt.Errorf("cannot get order from market %v", err)
	}

	switch order.GetOrderStatus() {
	case sonm.OrderStatus_ORDER_ACTIVE:
		megaHashes := order.GetBenchmarks().GPUEthHashrate() / hashes
		log.Debug("megaHashes", zap.Uint64("value", megaHashes))

		pricePerSecForPack := big.NewInt(0).Mul(actualPrice, big.NewInt(int64(megaHashes)))
		log.Debug("pricePerSecForPack", zap.String("value", pricePerSecForPack.String()))

		if pricePerSecForPack.Cmp(big.NewInt(0)) == 0 {
			return fmt.Errorf("actual price is zero")
		}

		//TODO: remake this part
		divider := big.NewInt(0).Mul(pricePerSecForPack, big.NewInt(100))
		div := big.NewInt(0).Div(divider, order.Price.Unwrap())
		log.Debug("div", zap.String("divider", divider.String()), zap.String("result", div.String()))

		changePricePercent, _ := big.NewFloat(0).SetInt(div).Float64()
		log.Debug("div -> changePricePercent float conversion result", zap.Float64("value", changePricePercent))
		if changePricePercent == 0 {
			return fmt.Errorf("calculated price delta is zero")
		}

		mnogo := changePricePercent > fullPath+t.c.cfg.Trade.OrdersChangePercent
		malo := changePricePercent < fullPath-t.c.cfg.Trade.OrdersChangePercent

		log.Debug("magic comparison", zap.Bool("mnogo", mnogo), zap.Bool("malo", malo),
			zap.Any("order_change_percent", t.c.cfg.Trade.OrdersChangePercent))

		if mnogo || malo {
			t.c.logger.Info("change price ==>  create reinvoice order", zap.String("active_orderID", order.Id.Unwrap().String()),
				zap.String("order_price", sonm.NewBigInt(order.Price.Unwrap()).ToPriceString()),
				zap.String("actual_price_for_pack", sonm.NewBigInt(pricePerSecForPack).ToPriceString()),
				zap.Float64("change_percent", changePricePercent))

			price := &sonm.Price{PerSecond: sonm.NewBigInt(pricePerSecForPack)}
			bench := t.GetBidBenchmarks(order)
			tag := fmt.Sprintf("reinvoice[new price][deal=%d]", orderDB.OrderID)

			if err := t.ReinvoiceOrder(ctx, price, bench, tag); err != nil {
				log.Warn("cannot create order", zap.Error(err))
				return err
			}

			if _, err := t.c.Market.CancelOrder(ctx, &sonm.ID{Id: fmt.Sprintf("%d", orderDB.OrderID)}); err != nil {
				log.Warn("cannot cancel order", zap.Error(err))
				return err
			}
		}
	case sonm.OrderStatus_ORDER_INACTIVE:
		log.Debug("order is not active", zap.String("ID", order.Id.Unwrap().String()))
		if err := t.c.db.UpdateOrderInDB(orderDB.OrderID, OrderStatusCancelled); err != nil {
			log.Warn("cannot update order status", zap.Error(err))
		}
	}
	return nil
}

func (t *TraderModule) ReinvoiceOrder(ctx context.Context, price *sonm.Price, bench map[string]uint64, tag string) error {
	order, err := t.c.Market.CreateOrder(ctx, &sonm.BidOrder{
		Duration: &sonm.Duration{Nanoseconds: 0},
		Price:    price,
		Tag:      tag,
		Identity: t.c.cfg.Trade.IdentityForBid,
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
		// todo: why lucky?
		t.c.logger.Warn("cannot create lucky order", zap.Error(err))
		return err
	}

	if err := t.c.db.SaveOrderIntoDB(&database.OrderDb{
		OrderID:    order.Id.Unwrap().Int64(),
		Price:      order.Price.Unwrap().Int64(),
		Hashrate:   order.Benchmarks.GPUEthHashrate(),
		StartTime:  time.Now(),
		Status:     OrderStatusReinvoice,
		ActualStep: 0,
	}); err != nil {
		return fmt.Errorf("cannot save reinvoice order %s to DB: %v", order.GetId().Unwrap().String(), err)
	}

	t.c.logger.Info("order has been reinvoiced", zap.Any("order", order), zap.String("tag", tag))
	return nil
}

func (t *TraderModule) GetChangeRequest(ctx context.Context, dealChangeRequest *sonm.DealInfoReply) error {
	time.Sleep(time.Duration(900 * time.Second))

	requestsList, err := t.c.DealClient.ChangeRequestsList(ctx, dealChangeRequest.Deal.Id)
	if err != nil {
		return fmt.Errorf("cannot get change request status: %v", err)
	}

	for _, cr := range requestsList.Requests {
		if cr.DealID == dealChangeRequest.Deal.Id {
			if cr.Status == sonm.ChangeRequestStatus_REQUEST_ACCEPTED {

				if err := t.c.db.UpdateChangeRequestStatusDealDB(dealChangeRequest.Deal.Id.Unwrap().Int64(), sonm.ChangeRequestStatus_REQUEST_ACCEPTED, dealChangeRequest.Deal.Price.Unwrap().Int64()); err != nil {
					return err
				}
				t.c.logger.Info("worker accepted change request", zap.String("deal", dealChangeRequest.Deal.Id.Unwrap().String()))
				return nil

			} else {
				_, err := t.c.DealClient.Finish(ctx, &sonm.DealFinishRequest{
					Id: dealChangeRequest.Deal.Id,
				})
				if err != nil {
					return fmt.Errorf("cannot finish deal: %v", err)
				}

				if err := t.c.db.GetChangeRequestStatusDealDB(dealChangeRequest.Deal.Id.Unwrap().Int64(), sonm.ChangeRequestStatus_REQUEST_REJECTED); err != nil {
					return err
				}
				t.c.logger.Info("worker didn't accept change request", zap.String("deal", dealChangeRequest.Deal.Id.Unwrap().String()))
			}
		}
	}
	return nil
}

func (t *TraderModule) GetBidBenchmarks(bidOrder *sonm.Order) map[string]uint64 {
	getBench := bidOrder.GetBenchmarks()
	return map[string]uint64{
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
}

func (t *TraderModule) CancelAllOrders(ctx context.Context) error {
	orders, err := t.c.db.GetOrdersFromDB()
	if err != nil {
		return fmt.Errorf("cannot get orders from DB: %v", err)
	}

	for _, o := range orders {
		if o.Status == OrderStatusReinvoice || o.Status == int64(sonm.OrderStatus_ORDER_ACTIVE) {
			_, err := t.c.Market.CancelOrder(ctx, &sonm.ID{
				Id: strconv.Itoa(int(o.OrderID)),
			})
			if err != nil {
				return err
			}
			t.c.logger.Info("order id cancelled", zap.Int64("order", o.OrderID))
			t.c.db.UpdateOrderInDB(o.OrderID, OrderStatusCancelled)
		}
	}
	return nil
}
