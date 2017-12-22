package structs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/sonm-io/core/proto"
)

func TestSlotCompare_compareCpuCores(t *testing.T) {
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

func TestSlotCompare_compareRamBytes(t *testing.T) {
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

func TestSlotCompare_compareGpuCount(t *testing.T) {
	cases := []struct {
		gpu1      pb.GPUCount
		gpu2      pb.GPUCount
		mustMatch bool
	}{
		{
			gpu1:      pb.GPUCount_NO_GPU,
			gpu2:      pb.GPUCount_NO_GPU,
			mustMatch: true,
		},
		{
			gpu1:      pb.GPUCount_NO_GPU,
			gpu2:      pb.GPUCount_SINGLE_GPU,
			mustMatch: true,
		},
		{
			gpu1:      pb.GPUCount_NO_GPU,
			gpu2:      pb.GPUCount_MULTIPLE_GPU,
			mustMatch: true,
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

func TestSlotCompare__compareStorage(t *testing.T) {
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

func TestSlotCompare_compareNetTrafficIn(t *testing.T) {
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

func TestSlotCompare_compareNetTrafficOut(t *testing.T) {
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

func TestSlotCompare_compareNetTrafficType(t *testing.T) {
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
			slot: &pb.Slot{Duration: uint64(MinSlotDuration.Seconds())},
			err:  errResourcesIsNil,
		},
		{
			slot: &pb.Slot{Duration: 0, Resources: &pb.Resources{}},
			err:  ErrDurationIsTooShort,
		},
	}

	for i, cc := range cases {
		_, err := NewSlot(cc.slot)
		assert.EqualError(t, err, cc.err.Error(), fmt.Sprintf("%d", i))
	}

}

func TestSlot_Compare(t *testing.T) {
	cases := []struct {
		orderType pb.OrderType
		one       *Slot
		two       *Slot
		mustMatch bool
	}{
		{
			one:       &Slot{inner: nil},
			two:       &Slot{inner: nil},
			mustMatch: true,
		},
		// compareCpuCores
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{CpuCores: 2}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{CpuCores: 1}}},
			mustMatch: true,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{CpuCores: 1}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{CpuCores: 2}}},
			mustMatch: false,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{CpuCores: 1}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{CpuCores: 1}}},
			mustMatch: true,
		},

		// compareRamBytes
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{RamBytes: 1}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{RamBytes: 2}}},
			mustMatch: false,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{RamBytes: 2}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{RamBytes: 1}}},
			mustMatch: true,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{RamBytes: 2}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{RamBytes: 2}}},
			mustMatch: true,
		},

		// compareGpuCountBid
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{GpuCount: pb.GPUCount_NO_GPU}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{GpuCount: pb.GPUCount_MULTIPLE_GPU}}},
			mustMatch: false,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{GpuCount: pb.GPUCount_SINGLE_GPU}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{GpuCount: pb.GPUCount_SINGLE_GPU}}},
			mustMatch: true,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{GpuCount: pb.GPUCount_MULTIPLE_GPU}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{GpuCount: pb.GPUCount_SINGLE_GPU}}},
			mustMatch: true,
		},

		// compareStorage
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{Storage: 2}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{Storage: 1}}},
			mustMatch: true,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{Storage: 1}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{Storage: 2}}},
			mustMatch: false,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{Storage: 1}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{Storage: 1}}},
			mustMatch: true,
		},

		// compareNetTrafficIn
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetTrafficIn: 2}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetTrafficIn: 1}}},
			mustMatch: true,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetTrafficIn: 1}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetTrafficIn: 2}}},
			mustMatch: false,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetTrafficIn: 2}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetTrafficIn: 2}}},
			mustMatch: true,
		},

		// compareNetTrafficOut
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetTrafficOut: 2}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetTrafficOut: 1}}},
			mustMatch: true,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetTrafficOut: 1}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetTrafficOut: 2}}},
			mustMatch: false,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetTrafficOut: 2}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetTrafficOut: 1}}},
			mustMatch: true,
		},

		// compare network type
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_NO_NETWORK}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_NO_NETWORK}}},
			mustMatch: true,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_NO_NETWORK}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_OUTBOUND}}},
			mustMatch: false,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_NO_NETWORK}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_INCOMING}}},
			mustMatch: false,
		},

		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_OUTBOUND}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_NO_NETWORK}}},
			mustMatch: true,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_OUTBOUND}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_OUTBOUND}}},
			mustMatch: true,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_OUTBOUND}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_INCOMING}}},
			mustMatch: false,
		},

		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_INCOMING}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_NO_NETWORK}}},
			mustMatch: true,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_INCOMING}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_OUTBOUND}}},
			mustMatch: true,
		},
		{
			two:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_INCOMING}}},
			one:       &Slot{inner: &pb.Slot{Resources: &pb.Resources{NetworkType: pb.NetworkType_INCOMING}}},
			mustMatch: true,
		},
	}

	for i, cc := range cases {
		isMatch := cc.one.Compare(cc.two)
		assert.Equal(t, cc.mustMatch, isMatch, fmt.Sprintf("%d", i))
	}
}
