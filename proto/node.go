package sonm

import "fmt"

func (m *TokenTransferRequest) Validate() error {
	if m.GetTo().IsZero() {
		return fmt.Errorf("destination address must not be zero")
	}
	if m.GetAmount().IsZero() {
		return fmt.Errorf("SNM amount must not be zero")
	}

	return nil
}
