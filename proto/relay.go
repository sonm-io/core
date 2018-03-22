package sonm

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/ethereum/go-ethereum/crypto"
)

// Validate validates the incoming handshake request.
func (m *HandshakeRequest) Validate() error {
	if len(m.Addr) != 20 {
		return fmt.Errorf("address must have exactly 20 bytes format")
	}

	switch m.PeerType {
	case PeerType_SERVER:
		if len(m.Sign) == 0 {
			return fmt.Errorf("sign field must not be empty for server-side handshake")
		}

		hash := chainhash.DoubleHashB(m.Addr)
		publicKey, err := crypto.SigToPub(hash, m.Sign)
		if err != nil {
			return fmt.Errorf("invalid signature")
		}

		if !bytes.Equal(crypto.PubkeyToAddress(*publicKey).Bytes(), m.Addr) {
			return fmt.Errorf("invalid signature for provided ETH address")
		}
	case PeerType_CLIENT:
		if len(m.Sign) != 0 {
			return fmt.Errorf("sign field must be empty for client-side handshake")
		}
	default:
		return fmt.Errorf("unknown peer type: %s", m.PeerType)
	}

	return nil
}

// HasUUID returns true if a request has UUID provided.
func (m *HandshakeRequest) HasUUID() bool {
	return len(m.UUID) != 0
}
