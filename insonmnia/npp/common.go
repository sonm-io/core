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
		return "npp"
	case sourceRelayedConnection:
		return "relay"
	case sourceNPPQUICConnection:
		return "npp/quic"
	default:
		return "unknown source"
	}
}
