package structs

import (
	"errors"
	"math/big"
	"time"

	"github.com/cnf/structhash"
	pb "github.com/sonm-io/core/proto"
)

// Order represents a safe order wrapper.
//
// This is used to decompose the validation out of the protocol. All
// methods must return the valid sub-structures.
type Order struct {
	*pb.Order
}

func (o *Order) Unwrap() *pb.Order {
	return o.Order
}

func NewOrder(o *pb.Order) (*Order, error) {
	if err := validateOrder(o); err != nil {
		return nil, err
	} else {
		return &Order{o}, nil
	}
}

func validateOrder(o *pb.Order) error {
	if o == nil {
		return errors.New("order cannot be nil")
	}

	if o.PricePerSecond == nil {
		return errors.New("price cannot be nil")
	}

	if o.PricePerSecond.Unwrap().Sign() <= 0 {
		return errors.New("price/sec must be positive")
	}

	if err := validateSlot(o.Slot); err != nil {
		return err
	}

	return nil
}

func (o *Order) GetID() string {
	return o.Order.GetId()
}

func (o *Order) SetID(ID string) {
	o.Order.Id = ID
}

func (o *Order) GetSlot() *Slot {
	slot, err := NewSlot(o.Order.GetSlot())
	if err != nil {
		panic("validation has failed")
	}
	return slot
}

func (o *Order) IsBid() bool {
	return o.GetOrderType() == pb.OrderType_BID
}

func (o *Order) GetDuration() time.Duration {
	return time.Duration(o.Slot.Duration) * time.Second
}

func (o *Order) GetTotalPrice() *big.Int {
	return CalculateTotalPrice(o.Unwrap())
}

func CalculateTotalPrice(order *pb.Order) *big.Int {
	pricePerSecond := order.PricePerSecond.Unwrap()
	durationInSeconds := big.NewInt(int64(order.GetSlot().GetDuration()))
	price := big.NewInt(0).Mul(pricePerSecond, durationInSeconds)

	return price
}

// CalculateSpecHash hashes handler's order and convert hash to big.Int
func CalculateSpecHash(order *pb.Order) string {
	s := structhash.Md5(order.GetSlot().GetResources(), 1)
	result := big.NewInt(0).SetBytes(s).String()
	return result
}
