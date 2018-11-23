package auth

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

const separator = "@"

// Addr represents a parsed unified address. It just encapsulates an ability
// to contain either ETH common address, network address or both of them
// separated by "@".
//
// No assumption is given for the real meaning of the addresses it contains.
// It's the user responsibility to verify that the given network address is
// owned by the specified ETH address if both of them provided.
//
// We use the following common address scheme:
// - When only ETH address specified - "0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD".
//   Usually this is intended to be used for real address resolution using
//   Rendezvous server.
// - When only network address specified - "localhost:8080", "127.0.0.1:10000", etc.
//   This can be used if you really know that the target is located in a
//   well-known place in the network.
// - When both ETH and network addresses specified - "0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD@localhost:8080".
//   This can be represented as a combination of first two cases. The difference
//   is in how you treat its real meaning. For example, the network address may
//   be used as a hint for NPP dialer.
//   It's the user's responsibility to verify that the given network address
//   is really matches the ETH address provided, usually, using TLS certificates.
type Addr struct {
	eth     *common.Address
	netAddr string
}

// ParseAddr parses the specified string into an unified address.
func ParseAddr(addr string) (*Addr, error) {
	parts := strings.SplitN(addr, separator, -1)

	switch len(parts) {
	case 1:
		addr := parts[0]

		if common.IsHexAddress(addr) {
			return NewETHAddr(common.HexToAddress(addr)), nil
		} else {
			return &Addr{netAddr: addr}, nil
		}
	case 2:
		ethAddr := parts[0]
		netAddr := parts[1]

		if !common.IsHexAddress(ethAddr) {
			return nil, errInvalidETHAddressFormat()
		}

		eth := common.HexToAddress(ethAddr)
		return &Addr{
			eth:     &eth,
			netAddr: netAddr,
		}, nil
	default:
		return nil, errInvalidETHAddressFormat()
	}
}

// NewETHAddr constructs a new unified address from the given ETH address.
func NewETHAddr(addr common.Address) *Addr {
	return &Addr{
		eth: &addr,
	}
}

func (m *Addr) ETH() (common.Address, error) {
	if m.eth == nil {
		return common.Address{}, fmt.Errorf("no Ethereum address specified")
	}

	return *m.eth, nil
}

func (m *Addr) Addr() (string, error) {
	if len(m.netAddr) == 0 {
		return "", fmt.Errorf("no network address specified")
	}

	return m.netAddr, nil
}

func (m Addr) String() string {
	if m.eth != nil {
		if len(m.netAddr) > 0 {
			return fmt.Sprintf("%s@%s", m.eth.Hex(), m.netAddr)
		}

		return m.eth.Hex()
	}

	return m.netAddr
}

func (m Addr) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m *Addr) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var addr string
	if err := unmarshal(&addr); err != nil {
		return err
	}

	value, err := ParseAddr(addr)
	if err != nil {
		return fmt.Errorf("cannot convert `%s` into an `auth.Addr` address: %s", addr, err)
	}

	*m = *value
	return nil
}

func errInvalidETHAddressFormat() error {
	return fmt.Errorf("invalid Ethereum address format")
}
