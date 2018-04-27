package sonm

import "strings"

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

func (r *DealsRequest) ConsumerIDNormalized() string {
	return strings.ToLower(r.ConsumerID)
}

func (r *DealsRequest) SupplierIDNormalized() string {
	return strings.ToLower(r.SupplierID)
}

func (r *DealsRequest) MasterIDNormalized() string {
	return strings.ToLower(r.MasterID)
}

func (r *OrdersRequest) AuthorIDNormalized() string {
	return strings.ToLower(r.AuthorID)
}

func (r *OrdersRequest) CounterpartyIDNormalized() string {
	return strings.ToLower(r.CounterpartyID)
}

func (o *Order) AuthorIDNormalized() string {
	return strings.ToLower(o.AuthorID)
}

func (o *Order) CounterpartyIDNormalized() string {
	return strings.ToLower(o.CounterpartyID)
}

func (d *Deal) ConsumerIDNormalized() string {
	return strings.ToLower(d.ConsumerID)
}

func (d *Deal) SupplierIDNormalized() string {
	return strings.ToLower(d.SupplierID)
}

func (d *Deal) MasterIDNormalized() string {
	return strings.ToLower(d.MasterID)
}

func (v *Validator) IDNormalized() string {
	return strings.ToLower(v.Id)
}

func (c *Certificate) OwnerIDNormalized() string {
	return strings.ToLower(c.OwnerID)
}

func (c *Certificate) ValidatorIDNormalized() string {
	return strings.ToLower(c.ValidatorID)
}
