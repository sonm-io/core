package sonm

import "strings"

func NewBenchmarks(benchmarks []uint64) (*Benchmarks, error) {
	b := &Benchmarks{
		Values: make([]uint64, len(benchmarks)),
	}
	copy(b.Values, benchmarks)
	if err := b.Validate(); err != nil {
		return nil, err
	}
	return b, nil
}

func (m *Benchmarks) ToArray() []uint64 {
	return m.Values
}

func (m *Benchmarks) Contains(other *Benchmarks) bool {
	return m.CPUSysbenchMulti() >= other.CPUSysbenchMulti() &&
		m.CPUSysbenchOne() >= other.CPUSysbenchOne() &&
		m.CPUCores() >= other.CPUCores() &&
		m.RAMSize() >= other.RAMSize() &&
		m.StorageSize() >= other.StorageSize() &&
		m.NetTrafficIn() >= other.NetTrafficIn() &&
		m.NetTrafficOut() >= other.NetTrafficOut() &&
		m.GPUCount() >= other.GPUCount() &&
		m.GPUMem() >= other.GPUMem() &&
		m.GPUEthHashrate() >= other.GPUEthHashrate() &&
		m.GPUCashHashrate() >= other.GPUCashHashrate() &&
		m.GPURedshift() >= other.GPURedshift()
}

// GetAttributeName converts profile cert attr number to
// human readable name.
func (m *Certificate) GetAttributeName() string {
	attrs := map[uint64]string{
		1201: "KYC2",
		1301: "KYC3",
		1401: "KYC4",
		1302: "Logo",
		1102: "Name",
		1202: "Website",
		2201: "Phone",
		1303: "Country",
		2202: "E-mail",
		2203: "Social networks",
		1304: "Is corporation",
		1103: "Description",
		1104: "KYC URL",
		1105: "KYC icon",
		1106: "KYC Price",
	}

	return attrs[m.GetAttribute()]
}

// GetAttributeNameNormalized returns GetAttributeName with spaces replaced by underscores.
func (m *Certificate) GetAttributeNameNormalized() string {
	return strings.Replace(m.GetAttributeName(), " ", "_", -1)
}
