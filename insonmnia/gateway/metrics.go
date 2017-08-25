package gateway

// Metrics describes a virtual service metrics.
type Metrics struct {
	Connections uint64
	InPackets   uint64
	OutPackets  uint64
	InBytes     uint64
	OutBytes    uint64

	ConnectionsPerSecond uint64
	InPacketsPerSecond   uint64
	OutPacketsPerSecond  uint64
	InBytesPerSecond     uint64
	OutBytesPerSecond    uint64
}

func (m *Metrics) Add(metrics *Metrics) {
	m.Connections += metrics.Connections
	m.InPackets += metrics.InPackets
	m.OutPackets += metrics.OutPackets
	m.InBytes += metrics.InBytes
	m.OutBytes += metrics.OutBytes

	m.ConnectionsPerSecond += metrics.ConnectionsPerSecond
	m.InPacketsPerSecond += metrics.InPacketsPerSecond
	m.OutPacketsPerSecond += metrics.OutPacketsPerSecond
	m.InBytesPerSecond += metrics.InBytesPerSecond
	m.OutBytesPerSecond += metrics.OutBytesPerSecond
}
