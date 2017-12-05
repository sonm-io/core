package util

import (
	"context"
	"crypto/x509"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestHitlessRotator(t *testing.T) {
	oldRotValidPeriod := validPeriod
	validPeriod = time.Second * 15
	defer func() {
		validPeriod = oldRotValidPeriod
	}()

	require := require.New(t)
	priv, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}
	ctx := context.Background()
	r, cfg, err := NewHitlessCertRotator(ctx, priv)
	require.NoError(err)
	defer r.Close()

	deadline := time.Now().Add(validPeriod * 2)
	for time.Now().Before(deadline) {
		tCfg, _ := cfg.GetCertificate(nil)
		x509Cert, err := x509.ParseCertificate(tCfg.Certificate[0])
		require.NoError(err)
		_, err = checkCert(x509Cert)
		require.NoError(err)

		tCfgCl, _ := cfg.GetClientCertificate(nil)
		x509CertCl, err := x509.ParseCertificate(tCfgCl.Certificate[0])
		require.NoError(err)
		_, err = checkCert(x509CertCl)
		require.NoError(err)

		time.Sleep(time.Second)
	}
}
