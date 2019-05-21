package optimus

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
