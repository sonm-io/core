package util

import (
	"crypto/tls"
	"crypto/x509"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func TestTLSGenCerts(t *testing.T) {
	priv, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}
	certPEM, keyPEM, err := GenerateCert(priv, time.Second*20)
	if err != nil {
		t.Fatalf("%v", err)
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("%v", err)
	}
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("%v", err)
	}
	_, err = checkCert(x509Cert)
	if err != nil {
		t.Fatal(err)
	}
}
