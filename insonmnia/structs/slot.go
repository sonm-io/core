package structs

import (
	"errors"

	pb "github.com/sonm-io/core/proto"
)

var (
	errSlotIsNil         = errors.New("order slot cannot be nil")
	errResourcesIsNil    = errors.New("slot resources cannot be nil")
	errStartTimeAfterEnd = errors.New("start time is after end time")
	errStartTimeRequired = errors.New("start time is required")
	errEndTimeRequired   = errors.New("end time is required")
)

type Slot struct {
	inner *pb.Slot
}

func NewSlot(s *pb.Slot) (*Slot, error) {
	if err := validateSlot(s); err != nil {
		return nil, err
	} else {
		return &Slot{inner: s}, nil
	}
}

func (s *Slot) GetResources() *Resources {
	resources, err := NewResources(s.inner.GetResources())
	if err != nil {
		panic("validation has failed")
	}

	return resources
}

func (s *Slot) Unwrap() *pb.Slot {
	return s.inner
}

func validateSlot(s *pb.Slot) error {
	if s == nil {
		return errSlotIsNil
	}

	if err := ValidateResources(s.GetResources()); err != nil {
		return err
	}

	return nil
}

func (s *Slot) compareSupplierRating(two *Slot) bool {
	return two.inner.GetSupplierRating() >= s.inner.GetSupplierRating()
}

func (s *Slot) compareCpuCoresBid(two *Slot) bool {
	return two.inner.GetResources().GetCpuCores() >= s.inner.GetResources().GetCpuCores()
}

func (s *Slot) compareRamBytesBid(two *Slot) bool {
	return two.inner.GetResources().GetRamBytes() >= s.inner.GetResources().GetRamBytes()
}

func (s *Slot) compareGpuCountBid(two *Slot) bool {
	return two.inner.GetResources().GetGpuCount() >= s.inner.GetResources().GetGpuCount()
}

func (s *Slot) compareStorageBid(two *Slot) bool {
	return two.inner.GetResources().GetStorage() >= s.inner.GetResources().GetStorage()
}

func (s *Slot) compareNetTrafficInBid(two *Slot) bool {
	return two.inner.GetResources().GetNetTrafficIn() >= s.inner.GetResources().GetNetTrafficIn()
}

func (s *Slot) compareNetTrafficOutBid(two *Slot) bool {
	return two.inner.GetResources().GetNetTrafficOut() >= s.inner.GetResources().GetNetTrafficOut()
}

func (s *Slot) compareNetworkTypeBid(two *Slot) bool {
	return two.inner.GetResources().GetNetworkType() >= s.inner.GetResources().GetNetworkType()
}

func (s *Slot) compareCpuCoresAsk(two *Slot) bool {
	return two.inner.GetResources().GetCpuCores() <= s.inner.GetResources().GetCpuCores()
}

func (s *Slot) compareRamBytesAsk(two *Slot) bool {
	return two.inner.GetResources().GetRamBytes() <= s.inner.GetResources().GetRamBytes()
}

func (s *Slot) compareGpuCountAsk(two *Slot) bool {
	return two.inner.GetResources().GetGpuCount() == s.inner.GetResources().GetGpuCount()
}

func (s *Slot) compareStorageAsk(two *Slot) bool {
	return two.inner.GetResources().GetStorage() <= s.inner.GetResources().GetStorage()
}

func (s *Slot) compareNetTrafficInAsk(two *Slot) bool {
	return two.inner.GetResources().GetNetTrafficIn() <= s.inner.GetResources().GetNetTrafficIn()
}

func (s *Slot) compareNetTrafficOutAsk(two *Slot) bool {
	return two.inner.GetResources().GetNetTrafficOut() <= s.inner.GetResources().GetNetTrafficOut()
}

func (s *Slot) compareNetworkTypeAsk(two *Slot) bool {
	return two.inner.GetResources().GetNetworkType() <= s.inner.GetResources().GetNetworkType()
}

func (s *Slot) Compare(two *Slot, orderType pb.OrderType) bool {
	// comparison of rating is performing
	// at the same way for different types of orders
	rt := s.compareSupplierRating(two)

	// TODO: Seems equal.
	if orderType == pb.OrderType_BID {
		return rt &&
			s.compareCpuCoresBid(two) &&
			s.compareRamBytesBid(two) &&
			s.compareGpuCountBid(two) &&
			s.compareStorageBid(two) &&
			s.compareNetTrafficInBid(two) &&
			s.compareNetTrafficOutBid(two) &&
			s.compareNetworkTypeBid(two)
	} else if orderType == pb.OrderType_ASK {
		return rt &&
			s.compareCpuCoresAsk(two) &&
			s.compareRamBytesAsk(two) &&
			s.compareGpuCountAsk(two) &&
			s.compareStorageAsk(two) &&
			s.compareNetTrafficInAsk(two) &&
			s.compareNetTrafficOutAsk(two) &&
			s.compareNetworkTypeAsk(two)
	}

	return false
}

func (s *Slot) Eq(other *Slot) bool {
	return s.compareCpuCoresBid(other) &&
		s.compareRamBytesBid(other) &&
		s.compareGpuCountBid(other) &&
		s.compareStorageBid(other) &&
		s.compareNetTrafficInBid(other) &&
		s.compareNetTrafficOutBid(other) &&
		s.compareNetworkTypeBid(other)
}
