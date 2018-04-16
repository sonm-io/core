package sonm

import (
	"errors"
)

func (m *BidOrder) Validate() error {
	if len(m.GetTag()) > 32 {
		return errors.New("tag value is too long")
	}

	return nil
}
