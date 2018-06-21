package connor

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type EmergencyModule struct {
	c *Connor
}

func NewEmergencyModule(c *Connor) *EmergencyModule {
	return &EmergencyModule{
		c: c,
	}
}

func (t *EmergencyModule) CancelAllActiveOrders(ctx context.Context) error {
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

func (t *EmergencyModule) CloseAllActiveDeals(ctx context.Context) error {
	return nil
}
