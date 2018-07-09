package relay

import (
	"fmt"

	"github.com/sonm-io/core/proto"
)

const (
	ErrInvalidHandshake int32 = iota
	ErrUnknownPeerType
	ErrTimeout
	ErrNoPeer
	ErrWrongNode
	ErrEmptyContinuum
)

type protocolError struct {
	code        int32
	description string
}

func newProtocolError(code int32, err error) *protocolError {
	return &protocolError{
		code:        code,
		description: err.Error(),
	}
}

func (m *protocolError) Error() string {
	return fmt.Sprintf("[%d] %s", m.code, m.description)
}

func errInvalidHandshake(err error) error {
	return newProtocolError(ErrInvalidHandshake, fmt.Errorf("invalid handshake: %s", err.Error()))
}

func errUnknownType(ty sonm.PeerType) error {
	return newProtocolError(ErrUnknownPeerType, fmt.Errorf("unknown handshake peer type: %s", ty))
}

func errTimeout() error {
	return newProtocolError(ErrTimeout, fmt.Errorf("timed out"))
}

func errNoPeer() error {
	return newProtocolError(ErrNoPeer, fmt.Errorf("no peer found"))
}

func errWrongNode() error {
	return newProtocolError(ErrWrongNode, fmt.Errorf("peer connected to wrong node"))
}

func errEmptyContinuum() error {
	return newProtocolError(ErrEmptyContinuum, fmt.Errorf("no nodes in the continuum"))
}
