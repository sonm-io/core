package gateway

// Metrics describes a virtual service metrics.
type Metrics struct {
	Connections uint64
	InBytes     uint64
	OutBytes    uint64
}

func (m *Metrics) Add(metrics *Metrics) {
	m.Connections += metrics.Connections
	m.InBytes += metrics.InBytes
	m.OutBytes += metrics.OutBytes
}
