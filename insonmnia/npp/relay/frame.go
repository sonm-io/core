package relay

import (
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gogo/protobuf/proto"
	"github.com/sonm-io/core/proto"
)

// SignedETHAddr represents self-signed ETH address.
type SignedETHAddr struct {
	addr common.Address
	sign []byte
}

// NewSignedAddr constructs a new self-signed Ethereum address using the specified
// private key.
func NewSignedAddr(key *ecdsa.PrivateKey) (SignedETHAddr, error) {
	addr := crypto.PubkeyToAddress(key.PublicKey)
	hash := chainhash.DoubleHashB(addr.Bytes())
	sign, err := crypto.Sign(hash, key)
	if err != nil {
		return SignedETHAddr{}, err
	}

	m := SignedETHAddr{
		addr: addr,
		sign: sign,
	}

	return m, nil
}

func (m *SignedETHAddr) Addr() common.Address {
	return m.addr
}

func newDiscover(addr common.Address) *sonm.HandshakeRequest {
	return &sonm.HandshakeRequest{
		PeerType: sonm.PeerType_DISCOVER,
		Addr:     addr.Bytes(),
	}
}

func newDiscoverResponse(addr string) *sonm.DiscoverResponse {
	return &sonm.DiscoverResponse{
		Addr: addr,
	}
}

func newServerHandshake(addr SignedETHAddr) *sonm.HandshakeRequest {
	return &sonm.HandshakeRequest{
		PeerType: sonm.PeerType_SERVER,
		Addr:     addr.addr.Bytes(),
		Sign:     addr.sign,
	}
}

func newClientHandshake(addr common.Address, uuid string) *sonm.HandshakeRequest {
	return &sonm.HandshakeRequest{
		PeerType: sonm.PeerType_CLIENT,
		Addr:     addr.Bytes(),
		UUID:     uuid,
	}
}

func sendFrame(wr io.Writer, message proto.Message) error {
	frame, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	if err := binary.Write(wr, binary.BigEndian, uint16(len(frame))); err != nil {
		return err
	}

	var bytesSent int
	for bytesSent < len(frame) {
		n, err := wr.Write(frame[bytesSent:])
		if err != nil {
			return err
		}

		bytesSent += n
	}

	return nil
}

func recvFrame(rd io.Reader, message proto.Message) error {
	var size uint16
	if err := binary.Read(rd, binary.BigEndian, &size); err != nil {
		return err
	}

	if size > 4096 {
		return fmt.Errorf("message too large")
	}

	var buf [4096]byte
	var bytesRead int
	for bytesRead < int(size) {
		n, err := rd.Read(buf[bytesRead:size])
		if err != nil {
			return err
		}

		bytesRead += n
	}

	if err := proto.Unmarshal(buf[:bytesRead], message); err != nil {
		return err
	}

	return nil
}
