package task_config

import (
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
)

type OrderConfig struct {
	ID         string     `yaml:"id"`
	BuyerID    string     `yaml:"buyer_id"`
	SupplierID string     `yaml:"supplier_id"`
	Price      string     `yaml:"price" required:"true"`
	OrderType  string     `yaml:"order_type" required:"true"`
	Slot       SlotConfig `yaml:"slot" required:"true"`
}

func (c *OrderConfig) IntoOrder() (*structs.Order, error) {
	orderType, err := structs.ParseOrderType(c.OrderType)
	if err != nil {
		return nil, err
	}

	slot, err := c.Slot.IntoSlot()
	if err != nil {
		return nil, err
	}

	return structs.NewOrder(&sonm.Order{
		Id:         c.ID,
		ByuerID:    c.BuyerID,
		SupplierID: c.SupplierID,
		Price:      c.Price,
		OrderType:  orderType,
		Slot:       slot.Unwrap(),
	})
}
