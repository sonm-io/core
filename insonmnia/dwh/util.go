package dwh

import "fmt"

const (
	CertificateName           = 1102
	CertificateCountry        = 1303
	MaxBenchmark       uint64 = 1 << 63
)

var (
	attributeToString = map[uint64]string{
		CertificateName:    "Name",
		CertificateCountry: "Country",
	}
)

func getBenchmarkColumn(id uint64) string {
	return fmt.Sprintf("benchmark%d", id)
}
