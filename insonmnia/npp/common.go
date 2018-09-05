package npp

type connSource int

const (
	sourceError connSource = iota
	sourceDirectConnection
	sourceNPPConnection
	sourceRelayedConnection
)

func (m connSource) String() string {
	switch m {
	case sourceDirectConnection:
		return "direct"
	case sourceNPPConnection:
		return "NPP"
	case sourceRelayedConnection:
		return "relay"
	default:
		return "unknown source"
	}
}
