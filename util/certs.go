package util

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base32"
	"fmt"
	"math/big"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func generateCert(ethpriv *ecdsa.PrivateKey) (*x509.Certificate, error) {
	var issuerCommonName = new(bytes.Buffer)
	// x509 Certificate signed with an randomly generated ecdsa key
	// Certificate contains signature of ecdsa publick with ethprivate key
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, err
	}

	var ethPubKey = btcec.PublicKey(ethpriv.PublicKey)
	issuerCommonName.Write(ethPubKey.SerializeCompressed())
	issuerCommonName.WriteByte('@')

	serializedPubKey, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return nil, err
	}
	// Issuer must be signed with ethkey
	var prv = btcec.PrivateKey(*ethpriv)
	signature, err := prv.Sign(chainhash.DoubleHashB(serializedPubKey))
	if err != nil {
		return nil, fmt.Errorf("failed to sign public key: %v", err)
	}
	issuerCommonName.Write(signature.Serialize())

	template := &x509.Certificate{
		SerialNumber: big.NewInt(100),
		Subject: pkix.Name{
			CommonName: base32.StdEncoding.EncodeToString(issuerCommonName.Bytes()),
		},
		NotBefore: time.Now().Add(-time.Hour * 24 * 7),
		NotAfter:  time.Now().Add(time.Hour * 24),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template,
		priv.Public(), priv)
	if err != nil {
		return nil, err
	}
	cert, err := x509.ParseCertificate(certDER)
	return cert, err
}

func checkCert(cert *x509.Certificate) (string, error) {
	// FORMAT compressedethpubkey@signature
	issuer, err := base32.StdEncoding.DecodeString(cert.Issuer.CommonName)
	if err != nil {
		return "", err
	}

	parts := bytes.Split(issuer, []byte("@"))
	if len(parts) != 2 {
		return "", fmt.Errorf("malformed issuer")
	}

	ethPubKey, err := btcec.ParsePubKey(parts[0], btcec.S256())
	if err != nil {
		return "", err
	}

	serializedPubKey, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return "", err
	}

	signature, err := btcec.ParseSignature(parts[1], btcec.S256())
	if err != nil {
		return "", err
	}

	if !signature.Verify(chainhash.DoubleHashB(serializedPubKey), ethPubKey) {
		return "", fmt.Errorf("invalid signature")
	}

	// Check that a public key of a Certificate is signed with eth publick key
	return ethcrypto.PubkeyToAddress(*ethPubKey.ToECDSA()).String(), nil
}
