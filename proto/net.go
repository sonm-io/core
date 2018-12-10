package sonm

import (
	"fmt"
	"net"
	"strings"

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

func TransformNetAddrs(addrs []net.Addr) ([]*Addr, error) {
	transformed := make([]*Addr, 0, len(addrs))

	for _, addr := range addrs {
		transformedAddr, err := NewAddr(addr)
		if err != nil {
			return nil, err
		}

		transformed = append(transformed, transformedAddr)
	}

	return transformed, nil
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
	host, port, err := netutil.SplitHostPort(endpoint)
	if err != nil {
		return nil, err
	}

	return &SocketAddr{
		Addr: host.String(),
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

func FormatAddrs(addrs ...*Addr) string {
	formatted := make([]string, len(addrs))
	for id, addr := range addrs {
		formatted[id] = fmt.Sprintf("%s:%d", addr.GetAddr().GetAddr(), addr.GetAddr().GetPort())
	}

	return fmt.Sprintf("[%s]", strings.Join(formatted, ", "))
}
