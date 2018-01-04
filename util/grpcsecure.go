package util

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/tls"
	"encoding/base32"
	"fmt"
	"io/ioutil"
	"net"

	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/auth"
	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
)

type tlsVerifier struct {
	credentials.TransportCredentials
}

func (tc tlsVerifier) Clone() credentials.TransportCredentials {
	return tlsVerifier{TransportCredentials: tc.TransportCredentials.Clone()}
}

func (tc tlsVerifier) ClientHandshake(ctx context.Context, authority string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, authInfo, err := tc.TransportCredentials.ClientHandshake(ctx, authority, conn)
	if err != nil {
		return conn, authInfo, err
	}

	authInfo, err = verifyCertificate(authInfo)
	if err != nil {
		return nil, nil, err
	}

	return conn, authInfo, nil
}

func (tc tlsVerifier) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, authInfo, err := tc.TransportCredentials.ServerHandshake(conn)
	if err != nil {
		return nil, nil, err
	}

	authInfo, err = verifyCertificate(authInfo)
	if err != nil {
		return nil, nil, err
	}

	return conn, authInfo, nil
}

func verifyCertificate(authInfo credentials.AuthInfo) (credentials.AuthInfo, error) {
	switch authInfo := authInfo.(type) {
	case credentials.TLSInfo:
		if len(authInfo.State.PeerCertificates) == 0 {
			return nil, fmt.Errorf("no peer certificates")
		}
		wallet, err := checkCert(authInfo.State.PeerCertificates[0])
		if err != nil {
			return nil, err
		}
		if !ethcommon.IsHexAddress(wallet) {
			return nil, fmt.Errorf("%s is not a valid eth Address", wallet)
		}
		return auth.EthAuthInfo{TLS: authInfo, Wallet: ethcommon.HexToAddress(wallet)}, nil
	default:
		return nil, fmt.Errorf("unsupported AuthInfo %s %T", authInfo.AuthType(), authInfo)
	}
}

// NewTLS wraps TLS TransportCredentials from grpc to add custom logic
func NewTLS(c *tls.Config) credentials.TransportCredentials {
	tc := credentials.NewTLS(c)
	return tlsVerifier{TransportCredentials: tc}
}

// TODO: Will be removed in the nearest future in favor of TLS-wallet extraction and token-based authorization.
type SelfSignedWallet struct {
	Message string
}

func NewSelfSignedWallet(key *ecdsa.PrivateKey) (*SelfSignedWallet, error) {
	address := crypto.PubkeyToAddress(key.PublicKey).Hex()
	message := crypto.Keccak256([]byte(address))

	sign, err := crypto.Sign(message, key)
	if err != nil {
		return nil, err
	}

	signed := new(bytes.Buffer)
	signed.WriteString(address)
	signed.WriteByte('@')
	signed.WriteString(base32.StdEncoding.EncodeToString(sign))

	return &SelfSignedWallet{Message: signed.String()}, nil
}

// WalletAccess supplies PerRPCCredentials from a given self-signed wallet.
type WalletAccess struct {
	wallet *SelfSignedWallet
}

// NewWalletAccess constructs the new PerRPCCredentials using with given
// self-signed wallet.
func NewWalletAccess(wallet *SelfSignedWallet) credentials.PerRPCCredentials {
	return &WalletAccess{
		wallet: wallet,
	}
}

func (c WalletAccess) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"wallet": c.wallet.Message,
	}, nil
}

func (c WalletAccess) RequireTransportSecurity() bool {
	return true
}

func VerifySelfSignedWallet(signedWallet string) (string, error) {
	parts := strings.Split(signedWallet, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("malformed wallet provided")
	}

	address := []byte(parts[0])
	sign, err := ioutil.ReadAll(base32.NewDecoder(base32.StdEncoding, strings.NewReader(parts[1])))
	if err != nil {
		return "", err
	}

	recoveredPub, err := crypto.Ecrecover(crypto.Keccak256(address), sign)
	if err != nil {
		return "", err
	}

	pubKey := crypto.ToECDSAPub(recoveredPub)
	recoveredAddr := crypto.PubkeyToAddress(*pubKey).Hex()

	return recoveredAddr, nil
}
