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

type DeployStatus int32

const (
	DeployStatusNotDeployed DeployStatus = 0
	DeployStatusDeployed    DeployStatus = 1
	DeployStatusDestroyed   DeployStatus = 2
)

type OrderStatus int32

const (
	OrderStatusCancelled OrderStatus = 3
	OrderStatusReinvoice OrderStatus = 4
)

func (t *TraderModule) SaveNewActiveDealsIntoDB(ctx context.Context) error {
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
				DeployStatus: int64(DeployStatusNotDeployed),
			})
		}
	}
	return nil
}

// Takes a decision depending on the status of the deal. Not deployed : deploy new container, deployed : create change request if necessary.
func (t *TraderModule) DealsTrading(ctx context.Context, actualPrice *big.Int) error {
	dealsDb, err := t.c.db.GetDealsFromDB()
	if err != nil {
		return fmt.Errorf("cannot get deals from database: %v", err)
	}

	for _, dealDB := range dealsDb {
		dealDeployStatus, err := t.c.db.GetDeployStatus(dealDB.DealID)
		if err != nil {
			return fmt.Errorf("cannot get deploy status from deal %v", err)
		}

		if dealDB.Status == int64(sonm.DealStatus_DEAL_ACCEPTED) {
			checkDealStatus, err := t.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(dealDB.DealID))
			if err != nil {
				return fmt.Errorf("cannot get deal status %v", err)
			}

			if checkDealStatus.Deal.Status == sonm.DealStatus_DEAL_CLOSED {
				if err = t.c.db.UpdateDeployAndDealStatusDB(dealDB.DealID, int64(DeployStatusDestroyed), sonm.DealStatus_DEAL_CLOSED); err != nil {
					return fmt.Errorf("cannot update deploy ans deal status %v", err)
				}
				t.c.logger.Info("deal closed on market, task tracking stop")
				continue
			}

			switch dealDeployStatus {
			case int64(DeployStatusNotDeployed):
				if err := t.ResponseToActiveDeal(ctx, dealDB); err != nil {
					return err
				}
			case int64(DeployStatusDeployed):
				if dealDB.ChangeRequestStatus != int64(sonm.ChangeRequestStatus_REQUEST_CREATED) {
					if err := t.deployedDealProfitTracking(ctx, actualPrice, dealDB); err != nil {
						return err
					}
				}
			case int64(DeployStatusDestroyed):
				continue
			}
		}
	}
	return nil
}

// Deploy new container and reinvoice order from deal.
func (t *TraderModule) ResponseToActiveDeal(ctx context.Context, dealDB *database.DealDB) error {
	dealOnMarket, err := t.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(dealDB.DealID))
	if err != nil {
		return fmt.Errorf("cannot get deal info: %v", err)
	}

	t.c.logger.Info("processing of deploying new container", zap.Any("deal_ID", dealOnMarket.Deal))

	newContainer, err := t.pool.DeployNewContainer(ctx, t.c.cfg, dealOnMarket.Deal, t.c.cfg.Pool.Image)
	if err != nil {
		t.c.logger.Warn("cannot start task", zap.Error(err))

		if err := t.ReinvoiceOrderFromDeal(ctx, dealOnMarket.Deal); err != nil {
			return err
		}
		return nil
	}

	if err := t.c.db.UpdateDeployStatusDealInDB(dealOnMarket.Deal.Id.Unwrap().Int64(), int64(DeployStatusDeployed)); err != nil {
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
	if err != nil {
		return fmt.Errorf("cannot get change percent from deal: %v", err)
	}

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

	bench, err := t.GetBidBenchmarks(bidOrder)
	if err != nil {
		return fmt.Errorf("cannot get benchmarks from bid order: %v", err)
	}

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

// Takes a decision depending on the status of the order.
func (t *TraderModule) OrderTrading(ctx context.Context, actualPrice *big.Int) error {
	orders, err := t.c.db.GetOrdersFromDB()
	if err != nil {
		return fmt.Errorf("cannot get orders from DB: %v", err)
	}

	for _, order := range orders {
		if order.Status != int64(OrderStatusCancelled) {
			if err := t.OrdersProfitTracking(ctx, actualPrice, order); err != nil {
				return fmt.Errorf("cannot start orders profit tracking: %v", err)
			}
		}
	}
	return nil
}

// ComparisonWithDealHashrate order price and new price. If the price went down/up order will reinvoice.
func (t *TraderModule) OrdersProfitTracking(ctx context.Context, actualPrice *big.Int, orderDB *database.OrderDb) error {
	order, err := t.c.Market.GetOrderByID(ctx, &sonm.ID{Id: strconv.Itoa(int(orderDB.OrderID))})
	if err != nil {
		return fmt.Errorf("cannot get order from market %v", err)
	}

	switch order.GetOrderStatus() {
	case sonm.OrderStatus_ORDER_ACTIVE:
		pack := int64(order.GetBenchmarks().GPUEthHashrate()) / hashes

		pricePerSecForPack := big.NewInt(0).Mul(actualPrice, big.NewInt(pack))
		if pricePerSecForPack == big.NewInt(0) {
			return fmt.Errorf("actual price = 0")
		}

		//TODO: remake this part
		div := big.NewInt(0).Div(big.NewInt(0).Mul(pricePerSecForPack, big.NewInt(100)), order.Price.Unwrap())
		if err != nil {
			return fmt.Errorf("cannot get change percent from deal: %v", err)
		}

		changePricePercent, _ := big.NewFloat(0).SetInt64(div.Int64()).Float64()
		if changePricePercent == 0 {
			return err
		}
		bench, err := t.GetBidBenchmarks(order)
		if err != nil {
			return fmt.Errorf("cannot get benchmarks from Order : %v, %v", order.Id.Unwrap().Int64(), err)
		}

		if changePricePercent > fullPath+t.c.cfg.Trade.OrdersChangePercent || changePricePercent < fullPath-t.c.cfg.Trade.OrdersChangePercent {
			t.c.logger.Info("change price ==>  create reinvoice order", zap.String("active_orderID", order.Id.Unwrap().String()),
				zap.String("order_price", sonm.NewBigInt(order.Price.Unwrap()).ToPriceString()),
				zap.String("actual_price_for_pack", sonm.NewBigInt(pricePerSecForPack).ToPriceString()),
				zap.Float64("change_percent", changePricePercent))

			if err := t.ReinvoiceOrder(ctx, &sonm.Price{PerSecond: sonm.NewBigInt(pricePerSecForPack)}, bench, "Reinvoice(update price): "+strconv.Itoa(int(orderDB.OrderID))); err != nil {
				return err
			}
			_, err = t.c.Market.CancelOrder(ctx, &sonm.ID{Id: strconv.Itoa(int(orderDB.OrderID))})
			if err != nil {
				return err
			}
		}
	case sonm.OrderStatus_ORDER_INACTIVE:
		t.c.logger.Info("order is not active", zap.String("ID", order.Id.Unwrap().String()))
		t.c.db.UpdateOrderInDB(orderDB.OrderID, int64(OrderStatusCancelled))
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
		t.c.logger.Error("cannot create lucky order", zap.Error(err))
		return err
	}

	if err := t.c.db.SaveOrderIntoDB(&database.OrderDb{
		OrderID:    order.Id.Unwrap().Int64(),
		Price:      order.Price.Unwrap().Int64(),
		Hashrate:   order.Benchmarks.GPUEthHashrate(),
		StartTime:  time.Now(),
		Status:     int64(OrderStatusReinvoice),
		ActualStep: 0,
	}); err != nil {
		return fmt.Errorf("cannot save reinvoice order %s to DB: %v", order.GetId().Unwrap().String(), err)
	}

	t.c.logger.Info("reinvoice order", zap.Any("order", order), zap.String("tag", tag))
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

func (t *TraderModule) CancelAllOrders(ctx context.Context) error {
	orders, err := t.c.db.GetOrdersFromDB()
	if err != nil {
		return fmt.Errorf("cannot get orders from DB: %v", err)
	}

	for _, o := range orders {
		if o.Status == int64(OrderStatusReinvoice) || o.Status == int64(sonm.OrderStatus_ORDER_ACTIVE) {
			_, err := t.c.Market.CancelOrder(ctx, &sonm.ID{
				Id: strconv.Itoa(int(o.OrderID)),
			})
			if err != nil {
				return err
			}
			t.c.logger.Info("order id cancelled", zap.Int64("order", o.OrderID))
			t.c.db.UpdateOrderInDB(o.OrderID, int64(OrderStatusCancelled))
		}
	}
	return nil
}
