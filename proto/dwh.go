package sonm

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
