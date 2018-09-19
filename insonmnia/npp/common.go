package npp

type ConnSource int

const (
	sourceError ConnSource = iota
	SourceDirectConnection
	SourceNPPConnection
	SourceRelayedConnection
)

func (m ConnSource) String() string {
	switch m {
	case SourceDirectConnection:
		return "direct"
	case SourceNPPConnection:
		return "NPP"
	case SourceRelayedConnection:
		return "relay"
	default:
		return "unknown source"
	}
}
