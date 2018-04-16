package sonm

import (
	"errors"
	"time"
)

// GetDuration returns order's duration in seconds.
func (m *Order) GetDuration() time.Duration {
	return time.Duration(m.Slot.Duration) * time.Second
}

func (m *BidOrder) Validate() error {
	if len(m.GetTag()) > 32 {
		return errors.New("tag value is too long")
	}

	return nil
}
