package sonm

const (
	NumNetflags = 3
)

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

func UintToNetflags(flags uint64) [NumNetflags]bool {
	var fixedNetflags [3]bool
	for idx := 0; idx < NumNetflags; idx++ {
		fixedNetflags[NumNetflags-1-idx] = flags&(1<<uint64(idx)) != 0
	}

	return fixedNetflags
}

func NetflagsToUint(flags [NumNetflags]bool) uint64 {
	var netflags uint64
	for idx, flag := range flags {
		if flag {
			netflags |= 1 << uint64(NumNetflags-1-idx)
		}
	}

	return netflags
}
