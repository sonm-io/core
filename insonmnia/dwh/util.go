package dwh

import (
	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
)

const (
	CertificateName           = 1102
	CertificateCountry        = 1303
	MaxBenchmark       uint64 = 9223372036854775808
)

var (
	attributeToString = map[uint64]string{
		CertificateName:    "Name",
		CertificateCountry: "Country",
	}
)

func CheckBenchmarks(benches *pb.Benchmarks) error {
	for idx, bench := range benches.Values {
		if bench >= MaxBenchmark {
			return errors.Errorf("benchmark %d is greater that %d", idx, MaxBenchmark)
		}
	}

	return nil
}
