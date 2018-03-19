package auth

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type Endpoint struct {
	EthAddress common.Address
	Endpoint   string
}

// NewEndpoint constructs a new verified endpoint from the given string
// representation.
// The format is <ethAddress>@<endpoint>, for example 8125721C2413d99a33E351e1F6Bb4e56b6b633FD@127.0.0.1:9090.
func NewEndpoint(endpoint string) (*Endpoint, error) {
	parsed := strings.SplitN(endpoint, "@", 2)
	if len(parsed) != 2 {
		return nil, fmt.Errorf("invalid Ethereum address format")
	}

	ethAddress := parsed[0]
	hostPort := parsed[1]

	if !common.IsHexAddress(ethAddress) {
		return nil, fmt.Errorf("invalid Ethereum address format")
	}

	return &Endpoint{
		EthAddress: common.HexToAddress(ethAddress),
		Endpoint:   hostPort,
	}, nil
}

func (m Endpoint) String() string {
	return fmt.Sprintf("%s:%s", m.EthAddress.Hex(), m.Endpoint)
}
