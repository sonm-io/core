package dwh

import (
	"sync"
)

const (
	CertificateName    = 1102
	CertificateCountry = 1303
)

var (
	attributeToString = map[uint64]string{
		CertificateName:    "Name",
		CertificateCountry: "Country",
	}
)

type filtersPool struct {
	p *sync.Pool
}

func newFiltersPool() filtersPool {
	return filtersPool{
		p: &sync.Pool{
			New: func() interface{} {
				return []*filter{}
			},
		}}
}

func (p filtersPool) Get() []*filter {
	filters := p.p.Get().([]*filter)
	return filters[:0]
}

func (p filtersPool) Put(filters []*filter) {
	p.p.Put(filters)
}
