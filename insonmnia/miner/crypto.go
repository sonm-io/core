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
	Wallet string
}

func (w *walletAuthenticator) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, authInfo, err := w.TransportCredentials.ServerHandshake(conn)
	if err != nil {
		return nil, nil, err
	}

	switch authInfo := authInfo.(type) {
	case util.EthAuthInfo:
		if !compareEthAddr(authInfo.Wallet, w.Wallet) {
			return nil, nil, fmt.Errorf("authorization failed: expected %s, actual %s", w.Wallet, authInfo.Wallet[2:])
		}
	default:
		return nil, nil, fmt.Errorf("unsupported AuthInfo %s %T", authInfo.AuthType(), authInfo)
	}

	return conn, authInfo, nil
}

func newWalletAuthenticator(c credentials.TransportCredentials, wallet string) credentials.TransportCredentials {
	return &walletAuthenticator{c, wallet}
}

func parseHubEndpoint(endpoint string) (string, string, error) {
	parsed := strings.SplitN(endpoint, "@", 2)
	if len(parsed) != 2 {
		return "", "", errInvalidEndpointFormat
	}

	ethAddr := parsed[0]
	socketAddr := parsed[1]

	if len(ethAddr) <= 2 {
		return "", "", errInvalidEthAddrFormat
	}
	if ethAddr[:2] == "0x" {
		ethAddr = ethAddr[2:]
	}
	return socketAddr, ethAddr, nil
}

func compareEthAddr(a, b string) bool {
	s1 := common.HexToAddress(a)
	s2 := common.HexToAddress(b)

	return s1.String() == s2.String()
}
