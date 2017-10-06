package structs

import (
	"errors"

	pb "github.com/sonm-io/core/proto"
)

var (
	errOrderIsNil  = errors.New("Order cannot be nil")
	errPriceIsZero = errors.New("Order price cannot be less or equal than zero")
)

type Order struct {
	inner *pb.Order
}

func (o *Order) Unwrap() *pb.Order {
	return o.inner
}

func NewOrder(o *pb.Order) (*Order, error) {
	if err := validateOrder(o); err != nil {
		return nil, err
	} else {
		return &Order{inner: o}, nil
	}
}

func validateOrder(o *pb.Order) error {
	if o == nil {
		return errOrderIsNil
	}

	if o.Price <= 0 {
		return errPriceIsZero
	}

	if err := validateSlot(o.Slot); err != nil {
		return err
	}

	return nil
}

func (o *Order) GetID() string {
	return o.inner.GetId()
}

func (o *Order) GetPrice() int64 {
	return o.inner.GetPrice()
}

func (o *Order) IsBid() bool {
	return o.inner.GetOrderType() == pb.OrderType_BID
}
