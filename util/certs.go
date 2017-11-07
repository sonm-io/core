package util

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base32"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/noxiouz/zapctx/ctxlog"
)

const validPeriod = time.Hour * 4

// HitlessCertRotator renews TLS cert periodically
type HitlessCertRotator interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
	Close()
}

type hitlessCertRotator struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex

	cert    *tls.Certificate
	ethPriv *ecdsa.PrivateKey
}

func NewHitlessCertRotator(ctx context.Context, ethPriv *ecdsa.PrivateKey) (HitlessCertRotator, *tls.Config, error) {
	var err error
	rotator := hitlessCertRotator{
		ethPriv: ethPriv,
	}

	rotator.cert, err = rotator.rotateOnce()
	if err != nil {
		return nil, nil, err
	}

	TLSConfig := tls.Config{
		GetCertificate: rotator.GetCertificate,
	}

	rotator.ctx, rotator.cancel = context.WithCancel(ctx)

	go rotator.rotation()
	return &rotator, &TLSConfig, nil
}

func (r *hitlessCertRotator) rotateOnce() (*tls.Certificate, error) {
	certPEM, keyPEM, err := GenerateCert(r.ethPriv)
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func (r *hitlessCertRotator) rotation() {
	t := time.NewTicker(validPeriod / 3)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			cert, err := r.rotateOnce()
			if err == nil {
				r.mu.Lock()
				r.cert = cert
				r.mu.Unlock()
			} else {
				log.G(r.ctx).Error("failed to rotate certificate", zap.Error(err))
			}
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *hitlessCertRotator) Close() {
	r.cancel()
}

// GetCertificate works as tls.Config.GetCertificate callback
func (r *hitlessCertRotator) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	r.mu.Lock()
	cert := r.cert
	r.mu.Unlock()
	return cert, nil
}

// GenerateCert generates new PEM encoded x509cert and privatekey key.
// Generated certificate contains signature of a publick key by eth key
func GenerateCert(ethpriv *ecdsa.PrivateKey) (cert []byte, key []byte, err error) {
	var issuerCommonName = new(bytes.Buffer)
	// x509 Certificate signed with an randomly generated ecdsa key
	// Certificate contains signature of ecdsa publick with ethprivate key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	var ethPubKey = btcec.PublicKey(ethpriv.PublicKey)
	issuerCommonName.Write(ethPubKey.SerializeCompressed())
	issuerCommonName.WriteByte('@')

	serializedPubKey, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return nil, nil, err
	}
	// Issuer must be signed with ethkey
	var prv = btcec.PrivateKey(*ethpriv)
	signature, err := prv.Sign(chainhash.DoubleHashB(serializedPubKey))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign public key: %v", err)
	}
	issuerCommonName.Write(signature.Serialize())

	template := &x509.Certificate{
		SerialNumber: big.NewInt(100),
		Subject: pkix.Name{
			CommonName: base32.StdEncoding.EncodeToString(issuerCommonName.Bytes()),
		},
		NotBefore: time.Now().Add(-time.Hour * 1),
		NotAfter:  time.Now().Add(validPeriod),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, priv.Public(), priv)
	if err != nil {
		return nil, nil, err
	}
	// PEM encoded cert and key to load via tls.X509KeyPair
	cert = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	key = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return cert, key, err
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
