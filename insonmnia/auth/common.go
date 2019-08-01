package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

var (
	LeakedInsecureKey = common.HexToAddress("0x8125721c2413d99a33e351e1f6bb4e56b6b633fd")
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

type Peer struct {
	*peer.Peer
	Addr common.Address
}

func FromContext(ctx context.Context) (*Peer, error) {
	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errNoPeerInfo()
	}

	addr, err := ExtractWalletFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &Peer{
		Peer: peerInfo,
		Addr: *addr,
	}, nil
}

func errNoPeerInfo() error {
	return errors.New("no peer info provided")
}

func ExtractWalletFromContext(ctx context.Context) (*common.Address, error) {
	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errNoPeerInfo()
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
