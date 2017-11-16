package util

import (
	"crypto/tls"
	"fmt"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
)

// EthAuthInfo implements credentials.AuthInfo
// It provides access to a wallet of a connected user
type EthAuthInfo struct {
	TLS    credentials.TLSInfo
	Wallet string
}

// AuthType implements credentials.AuthInfo interface
func (e EthAuthInfo) AuthType() string {
	return "ETH+" + e.TLS.AuthType()
}

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
		return EthAuthInfo{TLS: authInfo, Wallet: wallet}, nil
	default:
		return nil, fmt.Errorf("unsupported AuthInfo %s %T", authInfo.AuthType(), authInfo)
	}
}

// NewTLS wraps TLS TransportCredentials from grpc to add custom logic
func NewTLS(c *tls.Config) credentials.TransportCredentials {
	tc := credentials.NewTLS(c)
	return tlsVerifier{TransportCredentials: tc}
}
