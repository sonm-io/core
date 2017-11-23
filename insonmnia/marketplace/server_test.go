package marketplace

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
)

func makeOrder() *pb.Order {
	return &pb.Order{
		Price: 1,
		Slot: &pb.Slot{
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

func TestInMemOrderStorage_GetOrders_NilParams(t *testing.T) {
	s := NewInMemoryStorage()
	_, err := s.GetOrders(nil)
	assert.EqualError(t, err, errSearchParamsIsNil.Error())
}

func TestInMemOrderStorage_GetOrders_NilSlot(t *testing.T) {
	s := NewInMemoryStorage()
	_, err := s.GetOrders(&searchParams{})
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

func TestCompareWithType(t *testing.T) {
	cases := []struct {
		slotT     pb.OrderType
		slot      *pb.Slot
		order     *pb.Order
		mustMatch bool
	}{
		{
			slotT: pb.OrderType_ANY,
			slot: &pb.Slot{
				Resources: &pb.Resources{},
			},

			order: &pb.Order{
				OrderType: pb.OrderType_BID,
				Price:     1,
				Slot: &pb.Slot{
					Resources: &pb.Resources{},
				},
			},
			mustMatch: true,
		},
		{
			slotT: pb.OrderType_ANY,
			slot: &pb.Slot{
				Resources: &pb.Resources{},
			},

			order: &pb.Order{
				OrderType: pb.OrderType_ASK,
				Price:     1,
				Slot: &pb.Slot{
					Resources: &pb.Resources{},
				},
			},
			mustMatch: true,
		},

		{
			slotT: pb.OrderType_ASK,
			slot: &pb.Slot{
				Resources: &pb.Resources{},
			},

			order: &pb.Order{
				OrderType: pb.OrderType_ASK,
				Price:     1,
				Slot: &pb.Slot{
					Resources: &pb.Resources{},
				},
			},
			mustMatch: true,
		},
		{
			slotT: pb.OrderType_ASK,
			slot: &pb.Slot{
				Resources: &pb.Resources{},
			},

			order: &pb.Order{
				OrderType: pb.OrderType_BID,
				Price:     1,
				Slot: &pb.Slot{
					Resources: &pb.Resources{},
				},
			},
			mustMatch: false,
		},
	}

	for i, cc := range cases {
		ord, err := structs.NewOrder(cc.order)
		assert.NoError(t, err)
		sl, err := structs.NewSlot(cc.slot)
		assert.NoError(t, err)

		isMatch := compareOrderAndSlot(sl, ord, cc.slotT)
		assert.Equal(t, cc.mustMatch, isMatch, fmt.Sprintf("%d", i))
	}
}

func TestInMemOrderStorage_GetOrders_Count(t *testing.T) {
	stor := NewInMemoryStorage()
	for i := 0; i < 100; i++ {
		ord, err := structs.NewOrder(&pb.Order{
			Price:     1,
			OrderType: pb.OrderType_BID,
			Slot: &pb.Slot{
				Resources: &pb.Resources{},
			},
		})

		assert.NoError(t, err)

		_, err = stor.CreateOrder(ord)
		assert.NoError(t, err)
	}

	sl, err := structs.NewSlot(&pb.Slot{
		Resources: &pb.Resources{},
	})
	assert.NoError(t, err)

	search := &searchParams{
		slot:      sl,
		count:     3,
		orderType: pb.OrderType_BID,
	}

	found, err := stor.GetOrders(search)
	assert.NoError(t, err)

	assert.Equal(t, int(search.count), len(found))
}

func TestInMemOrderStorage_GetOrders_Count2(t *testing.T) {
	stor := NewInMemoryStorage()
	for i := 0; i < 100; i++ {
		bid, err := structs.NewOrder(&pb.Order{
			Price:     1,
			OrderType: pb.OrderType_BID,
			Slot: &pb.Slot{
				Resources: &pb.Resources{},
			},
		})

		assert.NoError(t, err)

		_, err = stor.CreateOrder(bid)
		assert.NoError(t, err)
	}

	ask, err := structs.NewOrder(&pb.Order{
		Price:     1,
		OrderType: pb.OrderType_ASK,
		Slot: &pb.Slot{
			Resources: &pb.Resources{},
		},
	})
	assert.NoError(t, err)

	_, err = stor.CreateOrder(ask)
	assert.NoError(t, err)

	sl, err := structs.NewSlot(&pb.Slot{
		Resources: &pb.Resources{},
	})
	assert.NoError(t, err)

	search := &searchParams{
		slot:      sl,
		count:     10,
		orderType: pb.OrderType_ASK,
	}

	found, err := stor.GetOrders(search)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(found))
}

func TestInMemOrderStorage_GetOrders_Count3(t *testing.T) {
	stor := NewInMemoryStorage()
	for i := 0; i < 100; i++ {
		var ot pb.OrderType
		if i%2 == 0 {
			ot = pb.OrderType_ASK
		} else {
			ot = pb.OrderType_BID
		}

		bid, err := structs.NewOrder(&pb.Order{
			Price:     1,
			OrderType: ot,
			Slot: &pb.Slot{
				Resources: &pb.Resources{},
			},
		})

		assert.NoError(t, err)

		_, err = stor.CreateOrder(bid)
		assert.NoError(t, err)
	}

	sl, err := structs.NewSlot(&pb.Slot{
		Resources: &pb.Resources{},
	})
	assert.NoError(t, err)

	search := []*searchParams{
		{
			slot:      sl,
			count:     5,
			orderType: pb.OrderType_ANY,
		},
		{
			slot:      sl,
			count:     10,
			orderType: pb.OrderType_ASK,
		},
		{
			slot:      sl,
			count:     50,
			orderType: pb.OrderType_BID,
		},
	}

	for _, ss := range search {
		found, err := stor.GetOrders(ss)
		assert.NoError(t, err)

		assert.Equal(t, int(ss.count), len(found))
	}
}

func TestMarketplace_GetOrders(t *testing.T) {
	mp := NewMarketplace(context.Background(), "")

	req := &pb.GetOrdersRequest{
		Slot:      nil,
		Count:     0,
		OrderType: pb.OrderType_ANY,
	}
	_, err := mp.GetOrders(nil, req)
	assert.EqualError(t, err, errSlotIsNil.Error())
}

func TestInMemOrderStorage_GetOrders_Ordering(t *testing.T) {
	// check if order is sorted
	stor := NewInMemoryStorage()

	for i := 100; i > 0; i-- {
		bid, err := structs.NewOrder(&pb.Order{
			Price:     int64(i + 1),
			OrderType: pb.OrderType_BID,
			Slot: &pb.Slot{
				Resources: &pb.Resources{},
			},
		})
		assert.NoError(t, err)

		_, err = stor.CreateOrder(bid)
		assert.NoError(t, err)
	}

	sl, err := structs.NewSlot(&pb.Slot{
		Resources: &pb.Resources{},
	})
	assert.NoError(t, err)

	search := &searchParams{
		slot:      sl,
		count:     10,
		orderType: pb.OrderType_BID,
	}

	found, err := stor.GetOrders(search)
	assert.NoError(t, err)

	assert.Equal(t, int(search.count), len(found))

	for i := 1; i < len(found); i++ {
		p1 := found[i-1].GetPrice()
		p2 := found[i].GetPrice()
		ok := p1 > p2

		assert.True(t, ok, fmt.Sprintf("iter %d :: %d should be gt %d", i, p1, p2))
	}
}
