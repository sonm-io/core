package auth

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

const separator = "@"

type Addr struct {
	eth     *common.Address
	netAddr string
}

func NewAddr(addr string) (*Addr, error) {
	parts := strings.SplitN(addr, separator, -1)

	switch len(parts) {
	case 1:
		addr := parts[0]

		if common.IsHexAddress(addr) {
			eth := common.HexToAddress(addr)
			return &Addr{eth: &eth}, nil
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
func NewETHAddr(addr common.Address) Addr {
	return Addr{
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

	value, err := NewAddr(addr)
	if err != nil {
		return fmt.Errorf("cannot convert `%s` into an `auth.Addr` address: %s", addr, err)
	}

	*m = *value
	return nil
}

func errInvalidETHAddressFormat() error {
	return fmt.Errorf("invalid Ethereum address format")
}
