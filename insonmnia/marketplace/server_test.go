package marketplace

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
)

func makeOrder() *pb.Order {
	return &pb.Order{
		Price: 1,
		Slot: &pb.Slot{
			StartTime: &pb.Timestamp{Seconds: 1},
			EndTime:   &pb.Timestamp{Seconds: 2},
			Resources: &pb.Resources{},
		},
	}
}

func TestInMemOrderStorage_CreateOrder(t *testing.T) {
	s := NewInMemoryStorage()
	order := makeOrder()
	o, _ := structs.NewOrder(order)

	created, err := s.CreateOrder(o)

	assert.NoError(t, err)
	assert.NotEmpty(t, created.GetID(), "order must have id after creation")
}

func TestNewInMemoryStorage_CreateOrder_Errors(t *testing.T) {
	cases := []struct {
		fn  func() *pb.Order
		err error
	}{
		{
			fn: func() *pb.Order {
				order := makeOrder()
				order.Price = 0
				return order
			},
			err: errPriceIsZero,
		},
		{
			fn: func() *pb.Order {
				order := makeOrder()
				order.Slot.StartTime = nil
				return order
			},
			err: errStartTimeRequired,
		},
		{
			fn: func() *pb.Order {
				order := makeOrder()
				order.Slot.EndTime = nil
				return order
			},
			err: errEndTimeRequired,
		},
		{
			fn: func() *pb.Order {
				order := makeOrder()
				order.Slot.EndTime = &pb.Timestamp{Seconds: 1}
				order.Slot.StartTime = &pb.Timestamp{Seconds: 2}
				return order
			},
			err: errStartTimeAfterEnd,
		},
		{
			fn: func() *pb.Order {
				order := makeOrder()
				order.Slot = nil
				return order
			},
			err: errSlotIsNil,
		},
		{
			fn: func() *pb.Order {
				order := makeOrder()
				order.Slot.Resources = nil
				return order
			},
			err: errResourcesIsNil,
		},
		{
			fn: func() *pb.Order {
				return nil
			},
			err: errOrderIsNil,
		},
	}

	for i, cc := range cases {
		_, err := structs.NewOrder(cc.fn())
		assert.EqualError(t, err, cc.err.Error(), fmt.Sprintf("%d", i))
	}
}

func TestInMemOrderStorage_DeleteOrder(t *testing.T) {
	s := NewInMemoryStorage()
	o, err := structs.NewOrder(makeOrder())
	assert.NoError(t, err)

	order, err := s.CreateOrder(o)
	assert.NoError(t, err)

	err = s.DeleteOrder(order.GetID())
	assert.NoError(t, err)
}

func TestInMemOrderStorage_DeleteOrder_NotExists(t *testing.T) {
	s := NewInMemoryStorage()
	err := s.DeleteOrder("1234")
	assert.EqualError(t, err, errOrderNotFound.Error())
}

func TestInMemOrderStorage_GetOrderByID(t *testing.T) {
	s := NewInMemoryStorage()
	order, err := structs.NewOrder(makeOrder())
	assert.NoError(t, err)

	created, err := s.CreateOrder(order)
	assert.NoError(t, err)
	assert.NotEmpty(t, created.GetID())

	found, err := s.GetOrderByID(created.GetID())
	assert.NoError(t, err)
	assert.Equal(t, created.GetID(), found.GetID())
	assert.Equal(t, created.GetPrice(), found.GetPrice())
}

func TestInMemOrderStorage_GetOrderByID_NotExists(t *testing.T) {
	s := NewInMemoryStorage()
	order, err := s.GetOrderByID("1234")
	assert.Nil(t, order)
	assert.EqualError(t, err, errOrderNotFound.Error())
}

func TestInMemOrderStorage_GetOrders_NilSlot(t *testing.T) {
	s := NewInMemoryStorage()
	_, err := s.GetOrders(nil)
	assert.EqualError(t, err, errSlotIsNil.Error())
}

func TestNewOrder(t *testing.T) {
	cases := []struct {
		ord *pb.Order
		err error
	}{
		{
			ord: nil,
			err: errOrderIsNil,
		},
		{
			ord: &pb.Order{
				Price: 0,
				Slot:  &pb.Slot{},
			},
			err: errPriceIsZero,
		},
		{
			ord: &pb.Order{
				Price: 1,
			},
			err: errSlotIsNil,
		},
		{
			ord: &pb.Order{
				Price: 1,
				Slot:  &pb.Slot{},
			},
			err: errResourcesIsNil,
		},
	}

	for i, cc := range cases {
		_, err := structs.NewOrder(cc.ord)
		assert.EqualError(t, err, cc.err.Error(), fmt.Sprintf("%d", i))
	}
}
