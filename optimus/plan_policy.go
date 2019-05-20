package optimus

import (
	"fmt"
)

const (
	planPolicyPrecise planPolicy = iota + 1
	planPolicyEntireMachine
)

type planPolicy int

// NewPlanPolicy constructs a new plan policy from the given string.
func newPlanPolicy(ty string) (planPolicy, error) {
	switch ty {
	case "precise":
		return planPolicyPrecise, nil
	case "entire_machine":
		return planPolicyEntireMachine, nil
	default:
		return planPolicyPrecise, fmt.Errorf("unknown plan type: %s", ty)
	}
}

func (m *planPolicy) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ty string
	if err := unmarshal(&ty); err != nil {
		return err
	}

	policy, err := newPlanPolicy(ty)
	if err != nil {
		return err
	}

	fmt.Printf("type: %v", ty)

	*m = policy
	return nil
}
