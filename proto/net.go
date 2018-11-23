package sonm

import (
	"fmt"
	"net"
	"strconv"

	"github.com/sonm-io/core/util/netutil"
)

func NewAddr(addr net.Addr) (*Addr, error) {
	socketAddr, err := NewSocketAddr(addr.String())
	if err != nil {
		return nil, err
	}

	return &Addr{
		Protocol: addr.Network(),
		Addr:     socketAddr,
	}, nil
}

func (m *Addr) IsValid() bool {
	return m != nil && m.Addr != nil
}

func (m *Addr) IntoTCP() (net.Addr, error) {
	if m.Protocol != "tcp" {
		return nil, fmt.Errorf("invalid protocol: %s", m.Protocol)
	}
	return m.Addr.IntoTCP()
}

func (m *Addr) IntoUDP() (*net.UDPAddr, error) {
	return m.Addr.IntoUDP()
}

// IsPrivate returns true if this address can't be reached from the Internet directly.
func (m *Addr) IsPrivate() bool {
	return m.Addr.IsPrivate()
}

func NewSocketAddr(endpoint string) (*SocketAddr, error) {
	host, portString, err := net.SplitHostPort(endpoint)
	if err != nil {
		return nil, err
	}
	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		return nil, err
	}
	return &SocketAddr{
		Addr: host,
		Port: uint32(port),
	}, nil
}

// IsPrivate returns true if this address can't be reached from the Internet directly.
func (m *SocketAddr) IsPrivate() bool {
	return netutil.IsPrivateIP(net.ParseIP(m.Addr))
}

func (m *SocketAddr) IntoTCP() (net.Addr, error) {
	return net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", m.Addr, m.Port))
}

func (m *SocketAddr) IntoUDP() (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", m.Addr, m.Port))
}
