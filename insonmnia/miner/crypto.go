package miner

import (
	"fmt"
	"net"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/util"
	"google.golang.org/grpc/credentials"
)

type walletAuthenticator struct {
	credentials.TransportCredentials
	Wallet common.Address
}

func (w *walletAuthenticator) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, authInfo, err := w.TransportCredentials.ServerHandshake(conn)
	if err != nil {
		return nil, nil, err
	}

	switch authInfo := authInfo.(type) {
	case util.EthAuthInfo:
		if !util.EqualAddresses(authInfo.Wallet, w.Wallet) {
			return nil, nil, fmt.Errorf("authorization failed: expected %s, actual %s", w.Wallet, authInfo.Wallet)
		}
	default:
		return nil, nil, fmt.Errorf("unsupported AuthInfo %s %T", authInfo.AuthType(), authInfo)
	}

	return conn, authInfo, nil
}

func newWalletAuthenticator(c credentials.TransportCredentials, wallet common.Address) credentials.TransportCredentials {
	return &walletAuthenticator{c, wallet}
}

func parseHubEndpoint(endpoint string) (string, common.Address, error) {
	parsed := strings.SplitN(endpoint, "@", 2)
	if len(parsed) != 2 {
		return "", common.Address{}, errInvalidEndpointFormat
	}

	ethAddr := parsed[0]
	socketAddr := parsed[1]

	if !common.IsHexAddress(ethAddr) {
		return "", common.Address{}, errInvalidEthAddrFormat
	}

	return socketAddr, common.HexToAddress(ethAddr), nil
}
