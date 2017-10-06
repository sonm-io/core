package structs

import (
	"fmt"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"

	pb "github.com/sonm-io/core/proto"
)

func TestNewInMemoryStorage_GetOrders_compareTime(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Hour)

	cases := []struct {
		slotStartTime int64
		slotEndTime   int64
		ordStartTime  int64
		ordEndTime    int64
		isMatch       bool
		message       string
	}{
		{
			slotStartTime: start.Unix(),
			slotEndTime:   end.Unix(),

			ordStartTime: start.Add(-1 * time.Hour).Unix(),
			ordEndTime:   end.Add(time.Hour).Unix(),

			isMatch: true,
			message: "Both time is match",
		},
		{
			slotStartTime: start.Unix(),
			slotEndTime:   end.Unix(),

			ordStartTime: start.Add(10 * time.Minute).Unix(),
			ordEndTime:   end.Add(-10 * time.Minute).Unix(),

			isMatch: false,
			message: "Both StartTime and EndTime is not match",
		},
		{
			slotStartTime: start.Unix(),
			slotEndTime:   end.Unix(),

			ordStartTime: start.Add(-10 * time.Minute).Unix(),
			ordEndTime:   end.Add(-10 * time.Minute).Unix(),

			isMatch: false,
			message: "StartTime is not match",
		},
		{
			slotStartTime: start.Unix(),
			slotEndTime:   end.Unix(),

			ordStartTime: start.Add(10 * time.Minute).Unix(),
			ordEndTime:   end.Add(10 * time.Minute).Unix(),

			isMatch: false,
			message: "End time is not match",
		},
	}

	for i, cc := range cases {
		s1 := &Slot{
			inner: &pb.Slot{
				StartTime: &pb.Timestamp{Seconds: cc.slotStartTime},
				EndTime:   &pb.Timestamp{Seconds: cc.slotEndTime},
			},
		}
		s2 := &Slot{
			inner: &pb.Slot{
				StartTime: &pb.Timestamp{Seconds: cc.ordStartTime},
				EndTime:   &pb.Timestamp{Seconds: cc.ordEndTime},
			},
		}

		ok := s1.compareTime(s2)
		assert.Equal(t, cc.isMatch, ok, fmt.Sprintf("%d :: %s", i, cc.message))
	}
}

func TestNewInMemoryStorage_GetOrders_compareSupRating(t *testing.T) {
	cases := []struct {
		r1        int64
		r2        int64
		mustMatch bool
		message   string
	}{
		{
			r1:        1,
			r2:        1,
			mustMatch: true,
		},
		{
			r1:        1,
			r2:        2,
			mustMatch: true,
		},
		{
			r1:        2,
			r2:        1,
			mustMatch: false,
		},
	}

	for i, cc := range cases {
		s1 := &Slot{
			inner: &pb.Slot{
				SupplierRating: cc.r1,
			},
		}
		s2 := &Slot{
			inner: &pb.Slot{
				SupplierRating: cc.r2,
			},
		}

		isMatch := s1.compareSupplierRating(s2)
		assert.Equal(t, cc.mustMatch, isMatch, fmt.Sprintf("%d", i))
	}
}

func TestNewInMemoryStorage_GetOrders_compareCpuCores(t *testing.T) {
	cases := []struct {
		c1        uint64
		c2        uint64
		mustMatch bool
	}{
		{
			c1:        1,
			c2:        1,
			mustMatch: true,
		},
		{
			c1:        1,
			c2:        2,
			mustMatch: true,
		},
		{
			c1:        2,
			c2:        1,
			mustMatch: false,
		},
	}

	for i, cc := range cases {
		s1 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					CpuCores: cc.c1,
				},
			},
		}
		s2 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					CpuCores: cc.c2,
				},
			},
		}

		isMatch := s1.compareCpuCores(s2)
		assert.Equal(t, cc.mustMatch, isMatch, fmt.Sprintf("%d", i))
	}
}

func TestNewInMemoryStorage_GetOrders_compareRamBytes(t *testing.T) {
	cases := []struct {
		ram1      uint64
		ram2      uint64
		mustMatch bool
	}{
		{
			ram1:      1,
			ram2:      1,
			mustMatch: true,
		},
		{
			ram1:      1,
			ram2:      2,
			mustMatch: true,
		},
		{
			ram1:      2,
			ram2:      1,
			mustMatch: false,
		},
	}

	for i, cc := range cases {
		s1 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					RamBytes: cc.ram1,
				},
			},
		}
		s2 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					RamBytes: cc.ram2,
				},
			},
		}

		isMatch := s1.compareRamBytes(s2)
		assert.Equal(t, cc.mustMatch, isMatch, fmt.Sprintf("%d", i))
	}
}

func TestNewInMemoryStorage_GetOrders_compareGpuCount(t *testing.T) {
	cases := []struct {
		gpu1      uint64
		gpu2      uint64
		mustMatch bool
	}{
		{
			gpu1:      1,
			gpu2:      1,
			mustMatch: true,
		},
		{
			gpu1:      1,
			gpu2:      2,
			mustMatch: true,
		},
		{
			gpu1:      2,
			gpu2:      1,
			mustMatch: false,
		},
	}

	for i, cc := range cases {
		s1 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					GpuCount: cc.gpu1,
				},
			},
		}
		s2 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					GpuCount: cc.gpu2,
				},
			},
		}

		isMatch := s1.compareGpuCount(s2)
		assert.Equal(t, cc.mustMatch, isMatch, fmt.Sprintf("%d", i))
	}
}

func TestNewInMemoryStorage_GetOrders_compareStorage(t *testing.T) {
	cases := []struct {
		stor1     uint64
		stor2     uint64
		mustMatch bool
	}{
		{
			stor1:     1,
			stor2:     1,
			mustMatch: true,
		},
		{
			stor1:     1,
			stor2:     2,
			mustMatch: true,
		},
		{
			stor1:     2,
			stor2:     1,
			mustMatch: false,
		},
	}

	for i, cc := range cases {
		s1 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					Storage: cc.stor1,
				},
			},
		}
		s2 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					Storage: cc.stor2,
				},
			},
		}

		isMatch := s1.compareStorage(s2)
		assert.Equal(t, cc.mustMatch, isMatch, fmt.Sprintf("%d", i))
	}
}

func TestNewInMemoryStorage_GetOrders_compareNetTrafficIn(t *testing.T) {
	cases := []struct {
		t1        uint64
		t2        uint64
		mustMatch bool
	}{
		{
			t1:        1,
			t2:        1,
			mustMatch: true,
		},
		{
			t1:        1,
			t2:        2,
			mustMatch: true,
		},
		{
			t1:        2,
			t2:        1,
			mustMatch: false,
		},
	}

	for i, cc := range cases {
		s1 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					NetTrafficIn: cc.t1,
				},
			},
		}
		s2 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					NetTrafficIn: cc.t2,
				},
			},
		}

		isMatch := s1.compareNetTrafficIn(s2)
		assert.Equal(t, cc.mustMatch, isMatch, fmt.Sprintf("%d", i))
	}
}

func TestNewInMemoryStorage_GetOrders_compareNetTrafficOut(t *testing.T) {
	cases := []struct {
		t1        uint64
		t2        uint64
		mustMatch bool
	}{
		{
			t1:        1,
			t2:        1,
			mustMatch: true,
		},
		{
			t1:        1,
			t2:        2,
			mustMatch: true,
		},
		{
			t1:        2,
			t2:        1,
			mustMatch: false,
		},
	}

	for i, cc := range cases {
		s1 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					NetTrafficOut: cc.t1,
				},
			},
		}
		s2 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					NetTrafficOut: cc.t2,
				},
			},
		}

		isMatch := s1.compareNetTrafficOut(s2)
		assert.Equal(t, cc.mustMatch, isMatch, fmt.Sprintf("%d", i))
	}
}

func TestNewInMemoryStorage_GetOrders_compareNetTrafficType(t *testing.T) {
	cases := []struct {
		n1        pb.NetworkType
		n2        pb.NetworkType
		mustMatch bool
	}{
		{
			n1:        pb.NetworkType_NO_NETWORK,
			n2:        pb.NetworkType_NO_NETWORK,
			mustMatch: true,
		},

		{
			n1:        pb.NetworkType_NO_NETWORK,
			n2:        pb.NetworkType_OUTBOUND,
			mustMatch: true,
		},
		{
			n1:        pb.NetworkType_NO_NETWORK,
			n2:        pb.NetworkType_INCOMING,
			mustMatch: true,
		},
		{
			n1:        pb.NetworkType_OUTBOUND,
			n2:        pb.NetworkType_NO_NETWORK,
			mustMatch: false,
		},
		{
			n1:        pb.NetworkType_OUTBOUND,
			n2:        pb.NetworkType_OUTBOUND,
			mustMatch: true,
		},
		{
			n1:        pb.NetworkType_OUTBOUND,
			n2:        pb.NetworkType_INCOMING,
			mustMatch: true,
		},
		{
			n1:        pb.NetworkType_INCOMING,
			n2:        pb.NetworkType_NO_NETWORK,
			mustMatch: false,
		},
		{
			n1:        pb.NetworkType_INCOMING,
			n2:        pb.NetworkType_OUTBOUND,
			mustMatch: false,
		},
		{
			n1:        pb.NetworkType_INCOMING,
			n2:        pb.NetworkType_INCOMING,
			mustMatch: true,
		},
	}

	for i, cc := range cases {
		s1 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					NetworkType: cc.n1,
				},
			},
		}
		s2 := &Slot{
			inner: &pb.Slot{
				Resources: &pb.Resources{
					NetworkType: cc.n2,
				},
			},
		}

		isMatch := s1.compareNetworkType(s2)
		assert.Equal(t, cc.mustMatch, isMatch, fmt.Sprintf("%d", i))
	}
}


func TestNewSlot(t *testing.T) {
	cases := []struct {
		slot *pb.Slot
		err  error
	}{
		{
			slot: nil,
			err:  errSlotIsNil,
		},
		{
			slot: &pb.Slot{},
			err:  errResourcesIsNil,
		},
		{
			slot: &pb.Slot{
				StartTime: &pb.Timestamp{Seconds: 1},
				Resources: &pb.Resources{},
			},
			err: errEndTimeRequired,
		},
		{
			slot: &pb.Slot{
				EndTime:   &pb.Timestamp{Seconds: 1},
				Resources: &pb.Resources{},
			},
			err: errStartTimeRequired,
		},
		{
			slot: &pb.Slot{
				StartTime: &pb.Timestamp{Seconds: 2},
				EndTime:   &pb.Timestamp{Seconds: 1},
				Resources: &pb.Resources{},
			},
			err: errStartTimeAfterEnd,
		},
	}

	for i, cc := range cases {
		_, err := NewSlot(cc.slot)
		assert.EqualError(t, err, cc.err.Error(), fmt.Sprintf("%d", i))
	}

}
