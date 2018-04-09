package structs

import (
	"encoding/json"
	"errors"

	"time"

	pb "github.com/sonm-io/core/proto"
)

var (
	errSlotIsNil          = errors.New("order slot cannot be nil")
	errResourcesIsNil     = errors.New("slot resources cannot be nil")
	ErrDurationIsTooShort = errors.New("duration is too short")
)

const (
	MinSlotDuration = 10 * time.Minute
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

func (s *Slot) MarshalJSON() ([]byte, error) {
	if s == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(s.inner)
}

func (s *Slot) UnmarshalJSON(data []byte) error {
	unmarshalled := pb.Slot{}
	err := json.Unmarshal(data, &unmarshalled)
	if err != nil {
		return err
	}
	s.inner = &unmarshalled
	return nil
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

	if time.Duration(s.Duration)*time.Second < MinSlotDuration {
		return ErrDurationIsTooShort
	}

	if err := ValidateResources(s.GetResources()); err != nil {
		return err
	}

	return nil
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

// Compare compares two slots, returns true if "s" slot is fits into an "another" slot
//
// Deprecated: no longer need to compare orders on client-side
func (s *Slot) Compare(another *Slot) bool {
	return s.compareCpuCores(another) &&
		s.compareRamBytes(another) &&
		s.compareGpuCount(another) &&
		s.compareStorage(another) &&
		s.compareNetTrafficIn(another) &&
		s.compareNetTrafficOut(another) &&
		s.compareNetworkType(another)
}
