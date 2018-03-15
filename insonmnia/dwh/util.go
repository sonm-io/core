package dwh

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
