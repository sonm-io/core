package task_config

import (
	"errors"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
)

type OrderConfig struct {
	ID         string     `yaml:"id"`
	BuyerID    string     `yaml:"buyer_id"`
	SupplierID string     `yaml:"supplier_id"`
	Price      int64      `yaml:"price" required:"true"`
	OrderType  string     `yaml:"order_type" required:"true"`
	Slot       SlotConfig `yaml:"slot" required:"true"`
}

func (c *OrderConfig) IntoOrder() (*structs.Order, error) {
	orderType, err := ParseOrderType(c.OrderType)
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

func ParseOrderType(ty string) (sonm.OrderType, error) {
	switch ty {
	case "ANY":
		return sonm.OrderType_ANY, nil
	case "BID":
		return sonm.OrderType_BID, nil
	case "ASK":
		return sonm.OrderType_ASK, nil
	default:
		return sonm.OrderType_ANY, errors.New("unknown order type")
	}
}
