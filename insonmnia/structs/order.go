package structs

import (
	"errors"

	pb "github.com/sonm-io/core/proto"
)

var (
	errOrderIsNil  = errors.New("Order cannot be nil")
	errPriceIsZero = errors.New("Order price cannot be less or equal than zero")
)

// Order represents a safe order wrapper.
//
// This is used for decomposition the validation out of the protocol. All
// methods must return the valid sub-structures.
type Order struct {
	inner *pb.Order
}

// ByPrice implements sort.Interface
// allows to sort Orders by Price filed
type ByPrice []*Order

func (a ByPrice) Len() int           { return len(a) }
func (a ByPrice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPrice) Less(i, j int) bool { return a[i].GetPrice() > a[j].GetPrice() }

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

func (o *Order) SetID(ID string) {
	o.inner.Id = ID
}

func (o *Order) GetPrice() int64 {
	return o.inner.GetPrice()
}

func (o *Order) GetSlot() *Slot {
	slot, err := NewSlot(o.inner.GetSlot())
	if err != nil {
		panic("validation has failed")
	}
	return slot
}

func (o *Order) IsBid() bool {
	return o.inner.GetOrderType() == pb.OrderType_BID
}

func (o *Order) GetType() pb.OrderType {
	return o.inner.GetOrderType()
}
