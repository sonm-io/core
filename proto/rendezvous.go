package sonm

import (
	"github.com/pkg/errors"
)

func (m *ConnectRequest) Validate() error {
	if m.Protocol == "" {
		m.Protocol = "tcp"
	}
	if m.ID == "" {
		return errors.New("destination ID s required")
	}

	return nil
}

func (m *RendezvousReply) Empty() bool {
	return (m.PublicAddr == nil || m.PublicAddr.Addr == nil) && len(m.PrivateAddrs) == 0
}
