package util

import (
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func TestTLSGenCerts(t *testing.T) {
	priv, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}
	cert, _, err := GenerateCert(priv)
	if err != nil {
		t.Fatalf("%v", err)
	}
	_, err = checkCert(cert)
	if err != nil {
		t.Fatal(err)
	}
}
