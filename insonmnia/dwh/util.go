package dwh

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

func stringSliceToSet(in []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range in {
		out[value] = true
	}

	return out
}
