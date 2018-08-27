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
	"io/ioutil"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/noxiouz/zapctx/ctxlog"
)

const defaultValidPeriod = time.Hour * 4

// HitlessCertRotator renews TLS cert periodically
type HitlessCertRotator interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
	GetClientCertificate(*tls.CertificateRequestInfo) (*tls.Certificate, error)
	Close()
}

type hitlessCertRotator struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex

	cert    *tls.Certificate
	ethPriv *ecdsa.PrivateKey

	certValidPeriod time.Duration
}

func NewHitlessCertRotator(ctx context.Context, ethPriv *ecdsa.PrivateKey) (HitlessCertRotator, *tls.Config, error) {
	return newHitlessCertRotator(ctx, ethPriv, defaultValidPeriod)
}

func newHitlessCertRotator(ctx context.Context, ethPriv *ecdsa.PrivateKey, certValidPeriod time.Duration) (HitlessCertRotator, *tls.Config, error) {
	var err error
	rotator := hitlessCertRotator{
		ethPriv:         ethPriv,
		certValidPeriod: certValidPeriod,
	}

	rotator.cert, err = rotator.rotateOnce()
	if err != nil {
		return nil, nil, err
	}

	TLSConfig := tls.Config{
		GetCertificate:       rotator.GetCertificate,
		GetClientCertificate: rotator.GetClientCertificate,
		// NOTE: if we do not set this, the gRPC client will check the hostname in provided
		// certificate. Probably we should consider solution with OverrideServerName from
		// TransportCredentials as we have no CA, we trust only in ethereum key pair, so
		// there should be no MITM-attack (subject to investigate).
		InsecureSkipVerify: true,
		ClientAuth:         tls.RequireAnyClientCert,
	}

	rotator.ctx, rotator.cancel = context.WithCancel(ctx)

	go rotator.rotation()
	return &rotator, &TLSConfig, nil
}

func (r *hitlessCertRotator) rotateOnce() (*tls.Certificate, error) {
	certPEM, keyPEM, err := GenerateCert(r.ethPriv, r.certValidPeriod)
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
	rotationPeriod := r.certValidPeriod / 3
	log.G(r.ctx).Debug("start certificate rotation loop", zap.Duration("every", rotationPeriod))
	t := time.NewTicker(rotationPeriod)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			log.G(r.ctx).Debug("rotate certificate")
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

func (r *hitlessCertRotator) getCert() *tls.Certificate {
	r.mu.Lock()
	cert := r.cert
	r.mu.Unlock()
	return cert
}

// GetCertificate works as tls.Config.GetCertificate callback
func (r *hitlessCertRotator) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return r.getCert(), nil
}

// GetClientCertificate works as tls.Config.GetClientCertificate callback
func (r *hitlessCertRotator) GetClientCertificate(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
	return r.getCert(), nil
}

// GenerateCert generates new PEM encoded x509cert and privatekey key.
// Generated certificate contains signature of a publick key by eth key
func GenerateCert(ethpriv *ecdsa.PrivateKey, validPeriod time.Duration) (cert []byte, key []byte, err error) {
	var issuerCommonName = new(bytes.Buffer)
	// x509 Certificate signed with an randomly generated RSA key
	// Certificate contains signature of ecdsa publick with ethprivate key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	ethPubKey := btcec.PublicKey(ethpriv.PublicKey)
	base32SerializedPubETHKey := base32.StdEncoding.EncodeToString(ethPubKey.SerializeCompressed())
	issuerCommonName.WriteString(base32SerializedPubETHKey)
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
	issuerCommonName.WriteString(base32.StdEncoding.EncodeToString(signature.Serialize()))

	//dnsName := ethcrypto.PubkeyToAddress(ethpriv.PublicKey)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(100),
		Subject: pkix.Name{
			CommonName: "127.0.0.1",
		},
		NotBefore:   time.Now().Add(-time.Hour * 1),
		NotAfter:    time.Now().Add(validPeriod),
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},

		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
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
	if time.Now().After(cert.NotAfter) {
		return "", fmt.Errorf("certificate has expired")
	}
	if time.Now().Before(cert.NotBefore) {
		return "", fmt.Errorf("certificate is not active yet")
	}
	// FORMAT:
	// base32CompressedPubKey@base32Signature
	parts := strings.Split(cert.Issuer.CommonName, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("malformed issuer")
	}

	compressedETHPubKey, err := ioutil.ReadAll(base32.NewDecoder(base32.StdEncoding, strings.NewReader(parts[0])))
	if err != nil {
		return "", err
	}
	ethPubKey, err := btcec.ParsePubKey(compressedETHPubKey, btcec.S256())
	if err != nil {
		return "", err
	}
	serializedPubKey, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return "", err
	}

	signatureBytes, err := ioutil.ReadAll(base32.NewDecoder(base32.StdEncoding, strings.NewReader(parts[1])))
	if err != nil {
		return "", err
	}
	signature, err := btcec.ParseSignature(signatureBytes, btcec.S256())
	if err != nil {
		return "", err
	}

	if !signature.Verify(chainhash.DoubleHashB(serializedPubKey), ethPubKey) {
		return "", fmt.Errorf("invalid signature")
	}

	// Check that a public key of a Certificate is signed with eth public key.
	return ethcrypto.PubkeyToAddress(*ethPubKey.ToECDSA()).String(), nil
}
