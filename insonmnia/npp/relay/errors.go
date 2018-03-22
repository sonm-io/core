package relay

import (
	"fmt"

	"github.com/sonm-io/core/proto"
)

const (
	ErrInvalidHandshake int32 = iota
	ErrUnknownPeerType
	ErrTimeout
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
