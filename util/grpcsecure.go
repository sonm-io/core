package util

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/insonmnia/auth"
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

// NewTLS wraps TLS TransportCredentials from gRPC to add custom verification
// logic.
func NewTLS(cfg *tls.Config) credentials.TransportCredentials {
	return tlsVerifier{TransportCredentials: credentials.NewTLS(cfg)}
}
