package npp

type connSource int

const (
	sourceError connSource = iota
	sourceDirectConnection
	sourceNPPConnection
	sourceRelayedConnection
	sourceNPPQUICConnection
)

func (m connSource) String() string {
	switch m {
	case sourceDirectConnection:
		return "direct"
	case sourceNPPConnection:
		return "NPP"
	case sourceRelayedConnection:
		return "relay"
	case sourceNPPQUICConnection:
		return "NPP/QUIC"
	default:
		return "unknown source"
	}
}
