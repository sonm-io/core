package auth

import (
	"bytes"
	"context"
	"fmt"
	"net"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// EthAuthInfo implements credentials.AuthInfo
// It provides access to a wallet of a connected user
type EthAuthInfo struct {
	TLS    credentials.TLSInfo
	Wallet common.Address
}

// AuthType implements credentials.AuthInfo interface
func (e EthAuthInfo) AuthType() string {
	return "ETH+" + e.TLS.AuthType()
}

func ExtractWalletFromContext(ctx context.Context) (*common.Address, error) {
	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no peer info")
	}

	switch auth := peerInfo.AuthInfo.(type) {
	case EthAuthInfo:
		return &auth.Wallet, nil
	default:
		return nil, fmt.Errorf("unknown auth info %T", peerInfo.AuthInfo)
	}
}

type WalletAuthenticator struct {
	credentials.TransportCredentials
	Wallet common.Address
}

func (w *WalletAuthenticator) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, authInfo, err := w.TransportCredentials.ServerHandshake(conn)
	if err != nil {
		return nil, nil, err
	}

	if err := w.compareWallets(authInfo); err != nil {
		return nil, nil, err
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
		if !equalAddresses(authInfo.Wallet, w.Wallet) {
			return fmt.Errorf("authorization failed: expected %s, actual %s",
				w.Wallet.Hex(), authInfo.Wallet.Hex())
		}
	default:
		return fmt.Errorf("unsupported AuthInfo %s %T", authInfo.AuthType(), authInfo)
	}

	return nil
}

func NewWalletAuthenticator(c credentials.TransportCredentials, wallet common.Address) credentials.TransportCredentials {
	return &WalletAuthenticator{c, wallet}
}

// TODO: Left for backward compabitility, prune later.
func equalAddresses(a, b common.Address) bool {
	return bytes.Equal(a.Bytes(), b.Bytes())
}

// EqualAddresses compares the two given ETH addresses for equality.
func EqualAddresses(a, b common.Address) bool {
	return equalAddresses(a, b)
}
