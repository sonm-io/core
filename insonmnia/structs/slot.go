package structs

import (
	"errors"

	pb "github.com/sonm-io/core/proto"
)

var (
	errSlotIsNil         = errors.New("Order slot cannot be nil")
	errResourcesIsNil    = errors.New("Slot resources cannot be nil")
	errStartTimeAfterEnd = errors.New("Start time is after end time")
	errStartTimeRequired = errors.New("Start time is required")
	errEndTimeRequired   = errors.New("End time is required")
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

	if s.GetStartTime() == nil {
		return errStartTimeRequired
	}

	if s.GetEndTime() == nil {
		return errEndTimeRequired
	}

	if s.GetStartTime().GetSeconds() >= s.GetEndTime().GetSeconds() {
		return errStartTimeAfterEnd
	}

	return nil
}

func (s *Slot) compareSupplierRating(two *Slot) bool {
	return two.inner.GetSupplierRating() >= s.inner.GetSupplierRating()
}

func (s *Slot) compareTime(two *Slot) bool {
	startOK := s.inner.GetStartTime().GetSeconds() >= two.inner.GetStartTime().GetSeconds()
	endOK := s.inner.GetEndTime().GetSeconds() <= two.inner.GetEndTime().GetSeconds()

	return startOK && endOK
}

func (s *Slot) compareCpuCores(two *Slot) bool {
	return two.inner.GetResources().GetCpuCores() >= s.inner.GetResources().GetCpuCores()
}

func (s *Slot) compareRamBytes(two *Slot) bool {
	return two.inner.GetResources().GetRamBytes() >= s.inner.GetResources().GetRamBytes()
}

func (s *Slot) compareGpuCount(two *Slot) bool {
	return two.inner.GetResources().GetGpuCount() >= s.inner.GetResources().GetGpuCount()
}

func (s *Slot) compareStorage(two *Slot) bool {
	return two.inner.GetResources().GetStorage() >= s.inner.GetResources().GetStorage()
}

func (s *Slot) compareNetTrafficIn(two *Slot) bool {
	return two.inner.GetResources().GetNetTrafficIn() >= s.inner.GetResources().GetNetTrafficIn()
}

func (s *Slot) compareNetTrafficOut(two *Slot) bool {
	return two.inner.GetResources().GetNetTrafficOut() >= s.inner.GetResources().GetNetTrafficOut()
}

func (s *Slot) compareNetworkType(two *Slot) bool {
	return two.inner.GetResources().GetNetworkType() >= s.inner.GetResources().GetNetworkType()
}

func (s *Slot) Compare(two *Slot) bool {
	return s.compareSupplierRating(two) &&
		s.compareTime(two) &&
		s.compareCpuCores(two) &&
		s.compareRamBytes(two) &&
		s.compareGpuCount(two) &&
		s.compareStorage(two) &&
		s.compareNetTrafficIn(two) &&
		s.compareNetTrafficOut(two) &&
		s.compareNetworkType(two)
}
