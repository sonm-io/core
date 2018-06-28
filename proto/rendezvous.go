package sonm

import (
	"fmt"
)

const (
	DefaultNPPProtocol = "tcp"
)

func (m *PublishRequest) Validate() error {
	if m.Protocol == "" {
		m.Protocol = DefaultNPPProtocol
	}

	return nil
}

func (m *ConnectRequest) Validate() error {
	if m.Protocol == "" {
		m.Protocol = DefaultNPPProtocol
	}
	if len(m.ID) != 20 {
		return fmt.Errorf("destination ID must have exactly 20 bytes format")
	}

	return nil
}

func (m *RendezvousReply) Empty() bool {
	return (m.PublicAddr == nil || m.PublicAddr.Addr == nil) && len(m.PrivateAddrs) == 0
}
