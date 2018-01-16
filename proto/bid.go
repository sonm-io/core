package sonm

import "time"

// GetDuration returns order's duration in seconds.
func (m *Order) GetDuration() time.Duration {
	return time.Duration(m.Slot.Duration) * time.Second
}
