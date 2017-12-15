package util

import (
	"crypto/tls"
	"fmt"
	"net"

	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// EthAuthInfo implements credentials.AuthInfo
// It provides access to a wallet of a connected user
type EthAuthInfo struct {
	TLS    credentials.TLSInfo
	Wallet ethcommon.Address
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
		if !ethcommon.IsHexAddress(wallet) {
			return nil, fmt.Errorf("%s is not a valid eth Address", wallet)
		}
		return EthAuthInfo{TLS: authInfo, Wallet: ethcommon.HexToAddress(wallet)}, nil
	default:
		return nil, fmt.Errorf("unsupported AuthInfo %s %T", authInfo.AuthType(), authInfo)
	}
}

// NewTLS wraps TLS TransportCredentials from grpc to add custom logic
func NewTLS(c *tls.Config) credentials.TransportCredentials {
	tc := credentials.NewTLS(c)
	return tlsVerifier{TransportCredentials: tc}
}

type WalletAuthenticator struct {
	credentials.TransportCredentials
	Wallet ethcommon.Address
}

func (w *WalletAuthenticator) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, authInfo, err := w.TransportCredentials.ServerHandshake(conn)
	if err != nil {
		return nil, nil, err
	}

	switch authInfo := authInfo.(type) {
	case EthAuthInfo:
		if !EqualAddresses(authInfo.Wallet, w.Wallet) {
			return nil, nil, fmt.Errorf("authorization failed: expected %s, actual %s", w.Wallet, authInfo.Wallet)
		}
	default:
		return nil, nil, fmt.Errorf("unsupported AuthInfo %s %T", authInfo.AuthType(), authInfo)
	}

	return conn, authInfo, nil
}

func (w *WalletAuthenticator) ClientHandshake(ctx context.Context, arg string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, authInfo, err := w.TransportCredentials.ClientHandshake(ctx, arg, conn)
	if err != nil {
		return nil, nil, err
	}

	if err := w.compareWallets(authInfo); err != nil {
		return nil, nil, err
	}

	return conn, authInfo, nil
}

func (w *WalletAuthenticator) compareWallets(authInfo credentials.AuthInfo) error {
	switch authInfo := authInfo.(type) {
	case EthAuthInfo:
		if !EqualAddresses(authInfo.Wallet, w.Wallet) {
			return fmt.Errorf("authorization failed: expected %s, actual %s", w.Wallet, authInfo.Wallet)
		}
	default:
		return fmt.Errorf("unsupported AuthInfo %s %T", authInfo.AuthType(), authInfo)
	}

	return nil
}

func NewWalletAuthenticator(c credentials.TransportCredentials, wallet ethcommon.Address) credentials.TransportCredentials {
	return &WalletAuthenticator{c, wallet}
}

func ParseEndpoint(endpoint string) (string, ethcommon.Address, error) {
	parsed := strings.SplitN(endpoint, "@", 2)
	if len(parsed) != 2 {
		return "", ethcommon.Address{}, errors.New("invalid Ethereum address format")
	}

	ethAddr := parsed[0]
	socketAddr := parsed[1]

	if !ethcommon.IsHexAddress(ethAddr) {
		return "", ethcommon.Address{}, errors.New("invalid Ethereum address format")
	}

	return socketAddr, ethcommon.HexToAddress(ethAddr), nil
}

func MakeWalletAuthenticatedClient(ctx context.Context, creds credentials.TransportCredentials, endpoint string) (*grpc.ClientConn, error) {
	sockaddr, ethAddr, err := ParseEndpoint(endpoint)
	if err != nil {
		conn, err := MakeGrpcClient(ctx, endpoint, creds)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	locatorCreds := NewWalletAuthenticator(creds, ethAddr)

	conn, err := MakeGrpcClient(ctx, sockaddr, locatorCreds)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
