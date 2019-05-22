package optimus

import (
	"fmt"
)

const (
	planPolicyPrecise = iota
	planPolicyEntireMachine
)

type planPolicy struct {
	Type int
}

func (m *planPolicy) IsPrecise() bool {
	return m.Type == planPolicyPrecise
}

func (m *planPolicy) IsEntireMachine() bool {
	return m.Type == planPolicyEntireMachine
}

func (m *planPolicy) UnmarshalText(text []byte) error {
	ty := string(text)

	switch ty {
	case "precise":
		*m = planPolicy{Type: planPolicyPrecise}
	case "entire_machine":
		*m = planPolicy{Type: planPolicyEntireMachine}
	default:
		return fmt.Errorf("unknown plan policy: %s", ty)
	}

	return nil
}
