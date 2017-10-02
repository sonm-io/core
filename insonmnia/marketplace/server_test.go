package marketplace

import (
	"testing"
	"time"

	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func TestInMemOrderStorage_CreateOrder(t *testing.T) {
	s := NewInMemoryStorage()
	order := &pb.Order{Price: 1, Slot: &pb.Slot{Resources: &pb.Resources{}}}
	order, err := s.CreateOrder(order)

	assert.NoError(t, err)
	assert.NotEmpty(t, order.Id, "order must have id after creation")
}

func TestInMemOrderStorage_CreateOrder_NoPrice(t *testing.T) {
	s := NewInMemoryStorage()
	order := &pb.Order{Slot: &pb.Slot{Resources: &pb.Resources{}}}
	order, err := s.CreateOrder(order)

	assert.EqualError(t, err, errPriceIsZero.Error())
}

func TestInMemOrderStorage_CreateOrder_Nil(t *testing.T) {
	s := NewInMemoryStorage()
	_, err := s.CreateOrder(nil)
	assert.EqualError(t, err, errOrderIsNil.Error())
}

func TestInMemOrderStorage_CreateOrder_NilSlot(t *testing.T) {
	s := NewInMemoryStorage()
	_, err := s.CreateOrder(&pb.Order{Price: 1})
	assert.EqualError(t, err, errSlotIsNil.Error())
}

func TestInMemOrderStorage_CreateOrder_NilResources(t *testing.T) {
	s := NewInMemoryStorage()
	_, err := s.CreateOrder(&pb.Order{Slot: &pb.Slot{}})
	assert.EqualError(t, err, errResourcesIsNil.Error())
}

func TestInMemOrderStorage_DeleteOrder(t *testing.T) {
	s := NewInMemoryStorage()
	order := &pb.Order{Price: 1, Slot: &pb.Slot{Resources: &pb.Resources{}}}
	order, err := s.CreateOrder(order)
	assert.NoError(t, err)

	err = s.DeleteOrder(order.Id)
	assert.NoError(t, err)
}

func TestInMemOrderStorage_DeleteOrder_NotExists(t *testing.T) {
	s := NewInMemoryStorage()
	err := s.DeleteOrder("1234")
	assert.EqualError(t, err, errOrderNotFound.Error())
}

func TestInMemOrderStorage_GetOrderByID(t *testing.T) {
	s := NewInMemoryStorage()
	order, err := s.CreateOrder(&pb.Order{Price: 1, Slot: &pb.Slot{Resources: &pb.Resources{}}})
	assert.NoError(t, err)
	assert.NotEmpty(t, order.Id)

	found, err := s.GetOrderByID(order.Id)
	assert.NoError(t, err)
	assert.Equal(t, order.Id, found.Id)
	assert.Equal(t, order.Price, found.Price)
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

func TestInMemOrderStorage_GetOrders_NilResources(t *testing.T) {
	s := NewInMemoryStorage()
	_, err := s.GetOrders(&pb.Slot{})
	assert.EqualError(t, err, errResourcesIsNil.Error())
}

func TestNewInMemoryStorage_GetOrders_compareTime(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Hour)

	slot := &pb.Slot{
		StartTime: &pb.Timestamp{Seconds: start.Unix()},
		EndTime:   &pb.Timestamp{Seconds: end.Unix()},
	}

	order := &pb.Order{
		Slot: &pb.Slot{
			StartTime: &pb.Timestamp{Seconds: start.Add(-1 * time.Hour).Unix()},
			EndTime:   &pb.Timestamp{Seconds: end.Add(time.Hour).Unix()},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareTime(slot, order)
	assert.True(t, ok, "Both tome is match")
}

func TestNewInMemoryStorage_GetOrders_compareTime2(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Hour)

	slot := &pb.Slot{
		StartTime: &pb.Timestamp{Seconds: start.Unix()},
		EndTime:   &pb.Timestamp{Seconds: end.Unix()},
	}

	order := &pb.Order{
		Slot: &pb.Slot{
			StartTime: &pb.Timestamp{Seconds: start.Add(10 * time.Minute).Unix()},
			EndTime:   &pb.Timestamp{Seconds: end.Add(-10 * time.Minute).Unix()},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareTime(slot, order)
	assert.False(t, ok, "Both StartTime and EndTime is not match")
}

func TestNewInMemoryStorage_GetOrders_compareTime3(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Hour)

	slot := &pb.Slot{
		StartTime: &pb.Timestamp{Seconds: start.Unix()},
		EndTime:   &pb.Timestamp{Seconds: end.Unix()},
	}

	order := &pb.Order{
		Slot: &pb.Slot{
			StartTime: &pb.Timestamp{Seconds: start.Add(-10 * time.Minute).Unix()},
			EndTime:   &pb.Timestamp{Seconds: end.Add(-10 * time.Minute).Unix()},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareTime(slot, order)
	assert.False(t, ok, "StartTime is not match")
}

func TestNewInMemoryStorage_GetOrders_compareTime4(t *testing.T) {

	start := time.Now()
	end := start.Add(time.Hour)

	slot := &pb.Slot{
		StartTime: &pb.Timestamp{Seconds: start.Unix()},
		EndTime:   &pb.Timestamp{Seconds: end.Unix()},
	}

	order := &pb.Order{
		Slot: &pb.Slot{
			StartTime: &pb.Timestamp{Seconds: start.Add(10 * time.Minute).Unix()},
			EndTime:   &pb.Timestamp{Seconds: end.Add(10 * time.Minute).Unix()},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareTime(slot, order)
	assert.False(t, ok, "EndTime is not match")
}

func TestNewInMemoryStorage_GetOrders_compareSupRating(t *testing.T) {
	slot := &pb.Slot{
		SupplierRating: 1,
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			SupplierRating: 1,
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareSupplierRating(slot, order)
	assert.True(t, ok, "Required rating is match order rating (eq)")
}

func TestNewInMemoryStorage_GetOrders_compareSupRating2(t *testing.T) {
	slot := &pb.Slot{
		SupplierRating: 1,
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			SupplierRating: 2,
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareSupplierRating(slot, order)
	assert.True(t, ok, "Required rating is match order rating (gt)")
}

func TestNewInMemoryStorage_GetOrders_compareSupRating3(t *testing.T) {
	slot := &pb.Slot{
		SupplierRating: 2,
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			SupplierRating: 1,
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareSupplierRating(slot, order)
	assert.False(t, ok, "Required rating is NOT match order rating (less)")
}

func TestNewInMemoryStorage_GetOrders_compareCpuCoures(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			CpuCores: 1,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				CpuCores: 1,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareCpuCores(slot, order)
	assert.True(t, ok, "Required cpuCores is match order cpuCores (eq)")
}

func TestNewInMemoryStorage_GetOrders_compareCpuCoures2(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			CpuCores: 1,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				CpuCores: 2,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareCpuCores(slot, order)
	assert.True(t, ok, "Required cpuCores is match order cpuCores (gt)")
}

func TestNewInMemoryStorage_GetOrders_compareCpuCoures3(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			CpuCores: 4,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				CpuCores: 2,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareCpuCores(slot, order)
	assert.False(t, ok, "Required cpuCores is NOT match order cpuCores (less)")
}

func TestNewInMemoryStorage_GetOrders_compareRamBytes(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			RamBytes: 1,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				RamBytes: 2,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareRamBytes(slot, order)
	assert.True(t, ok, "Required ramBytes is match order ramBytes (gt)")
}

func TestNewInMemoryStorage_GetOrders_compareRamBytes2(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			RamBytes: 1,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				RamBytes: 1,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareRamBytes(slot, order)
	assert.True(t, ok, "Required ramBytes is match order ramBytes (eq)")
}

func TestNewInMemoryStorage_GetOrders_compareRamBytes3(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			RamBytes: 2,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				RamBytes: 1,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareRamBytes(slot, order)
	assert.False(t, ok, "Required ramBytes is NOT match order ramBytes (less)")
}

func TestNewInMemoryStorage_GetOrders_compareGpuCount(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			GpuCount: 1,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				GpuCount: 1,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareGpuCount(slot, order)
	assert.True(t, ok, "Required gpuCount is match order gpuCount (eq)")
}

func TestNewInMemoryStorage_GetOrders_compareGpuCount2(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			GpuCount: 1,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				GpuCount: 2,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareGpuCount(slot, order)
	assert.True(t, ok, "Required gpuCount is match order gpuCount (gt)")
}

func TestNewInMemoryStorage_GetOrders_compareGpuCount3(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			GpuCount: 2,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				GpuCount: 1,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareGpuCount(slot, order)
	assert.False(t, ok, "Required gpuCount is NOT match order gpuCount (less)")
}

func TestNewInMemoryStorage_GetOrders_compareStorage(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			Storage: 1,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				Storage: 1,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareStorage(slot, order)
	assert.True(t, ok, "Required storage is match order storage (eq)")
}

func TestNewInMemoryStorage_GetOrders_compareStorage2(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			Storage: 1,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				Storage: 2,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareStorage(slot, order)
	assert.True(t, ok, "Required storage is match order storage (gt)")
}

func TestNewInMemoryStorage_GetOrders_compareStorage3(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			Storage: 2,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				Storage: 1,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareStorage(slot, order)
	assert.False(t, ok, "Required storage is NOT match order storage (less)")
}

func TestNewInMemoryStorage_GetOrders_compareNetTrafficIn(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			NetTrafficIn: 1,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				NetTrafficIn: 1,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareNetTrafficIn(slot, order)
	assert.True(t, ok, "Required NetTrafficIn is match order NetTrafficIn (eq)")
}

func TestNewInMemoryStorage_GetOrders_compareNetTrafficIn2(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			NetTrafficIn: 1,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				NetTrafficIn: 2,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareNetTrafficIn(slot, order)
	assert.True(t, ok, "Required NetTrafficIn is match order NetTrafficIn (gt)")
}

func TestNewInMemoryStorage_GetOrders_compareNetTrafficIn3(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			NetTrafficIn: 2,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				NetTrafficIn: 1,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareNetTrafficIn(slot, order)
	assert.False(t, ok, "Required NetTrafficIn is NOT match order NetTrafficIn (less)")
}

func TestNewInMemoryStorage_GetOrders_compareNetTrafficOut(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			NetTrafficOut: 1,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				NetTrafficOut: 1,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareNetTrafficOut(slot, order)
	assert.True(t, ok, "Required NetTrafficOut is match order NetTrafficOut (eq)")
}

func TestNewInMemoryStorage_GetOrders_compareNetTrafficOut2(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			NetTrafficOut: 1,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				NetTrafficOut: 2,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareNetTrafficOut(slot, order)
	assert.True(t, ok, "Required NetTrafficOut is match order NetTrafficOut (gt)")
}

func TestNewInMemoryStorage_GetOrders_compareNetTrafficOut3(t *testing.T) {
	slot := &pb.Slot{
		Resources: &pb.Resources{
			NetTrafficOut: 2,
		},
	}
	order := &pb.Order{
		Slot: &pb.Slot{
			Resources: &pb.Resources{
				NetTrafficOut: 1,
			},
		},
	}

	s := NewInMemoryStorage().(*inMemOrderStorage)
	ok := s.compareNetTrafficOut(slot, order)
	assert.False(t, ok, "Required NetTrafficOut is NOT match order NetTrafficOut (less)")
}

func TestNewInMemoryStorage_GetOrders_compareNetTrafficType(t *testing.T) {
	cases := []struct {
		slot    pb.NetworkType
		order   pb.NetworkType
		isMatch bool
	}{
		{
			slot:    pb.NetworkType_NO_NETWORK,
			order:   pb.NetworkType_NO_NETWORK,
			isMatch: true,
		},

		{
			slot:    pb.NetworkType_NO_NETWORK,
			order:   pb.NetworkType_OUTBOUND,
			isMatch: true,
		},
		{
			slot:    pb.NetworkType_NO_NETWORK,
			order:   pb.NetworkType_INCOMING,
			isMatch: true,
		},
		{
			slot:    pb.NetworkType_OUTBOUND,
			order:   pb.NetworkType_NO_NETWORK,
			isMatch: false,
		},
		{
			slot:    pb.NetworkType_OUTBOUND,
			order:   pb.NetworkType_OUTBOUND,
			isMatch: true,
		},
		{
			slot:    pb.NetworkType_OUTBOUND,
			order:   pb.NetworkType_INCOMING,
			isMatch: true,
		},
		{
			slot:    pb.NetworkType_INCOMING,
			order:   pb.NetworkType_NO_NETWORK,
			isMatch: false,
		},
		{
			slot:    pb.NetworkType_INCOMING,
			order:   pb.NetworkType_OUTBOUND,
			isMatch: false,
		},
		{
			slot:    pb.NetworkType_INCOMING,
			order:   pb.NetworkType_INCOMING,
			isMatch: true,
		},
	}
	s := NewInMemoryStorage().(*inMemOrderStorage)

	for _, cc := range cases {
		ok := s.compareNetType(&pb.Slot{
			Resources: &pb.Resources{
				NetworkType: cc.slot,
			},
		}, &pb.Order{
			Slot: &pb.Slot{
				Resources: &pb.Resources{
					NetworkType: cc.order,
				},
			},
		})
		assert.Equal(t, cc.isMatch, ok)
	}
}
