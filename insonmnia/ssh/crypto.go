package ssh

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type SSHIdentity struct {
	Addr common.Address
	Sign []byte
}

func (m *SSHIdentity) String() string {
	return fmt.Sprintf("%s@%s", m.Addr.Hex(), base32.StdEncoding.EncodeToString(m.Sign))
}

func NewSSHIdentity(key *ecdsa.PrivateKey) (*SSHIdentity, error) {
	addr := crypto.PubkeyToAddress(key.PublicKey)
	hash := chainhash.DoubleHashB(addr.Bytes())
	sign, err := crypto.Sign(hash, key)
	if err != nil {
		return nil, err
	}

	return &SSHIdentity{Addr: addr, Sign: sign}, nil
}

func ParseSSHIdentity(v string) (*SSHIdentity, error) {
	parts := strings.SplitN(v, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format")
	}

	sign, err := base32.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	return &SSHIdentity{
		Addr: common.HexToAddress(parts[0]),
		Sign: sign,
	}, nil
}

func (m *SSHIdentity) Verify() error {
	hash := chainhash.DoubleHashB(m.Addr.Bytes())
	publicKey, err := crypto.SigToPub(hash, m.Sign)
	if err != nil {
		return fmt.Errorf("invalid signature")
	}

	if !bytes.Equal(crypto.PubkeyToAddress(*publicKey).Bytes(), m.Addr.Bytes()) {
		return fmt.Errorf("invalid signature for provided ETH address")
	}

	return nil
}
