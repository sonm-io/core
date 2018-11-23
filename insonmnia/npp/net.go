package npp

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"github.com/libp2p/go-reuseport"
	"github.com/sonm-io/core/util/netutil"
)

const (
	protocol = "tcp"
)

type Port uint16

// DialContext allows to dial the remote peer with SO_REUSEPORT and
// SO_REUSEADDR options configured.
func DialContext(ctx context.Context, network, laddr, raddr string) (net.Conn, error) {
	if !reuseport.Available() {
		return nil, syscall.ENOPROTOOPT
	}

	var dialer reuseport.Dialer
	if laddr != "" {
		localAddr, err := reuseport.ResolveAddr(network, laddr)
		if err != nil {
			return nil, err
		}
		dialer.D.LocalAddr = localAddr
	}

	return dialer.DialContext(ctx, network, raddr)
}

// PrivateAddrs collects and returns private addresses of a network interfaces
// the socket bind on.
func privateAddrs(addr net.Addr) ([]net.Addr, error) {
	ip, port, err := netutil.SplitHostPort(addr.String())
	if err != nil {
		return nil, err
	}

	if !ip.IsUnspecified() {
		return []net.Addr{addr}, nil
	}

	// For unspecified bind addresses, actual address set has to be resolved.
	// In other words, unspecified means every available (but still private)
	// address for the host.
	ips, err := netutil.GetAvailableIPs()
	if err != nil {
		return nil, err
	}

	var addrs []net.Addr
	for _, ip := range ips {
		addr, err := net.ResolveTCPAddr(protocol, fmt.Sprintf("%s:%d", ip, uint16(port)))
		if err != nil {
			return nil, err
		}

		addrs = append(addrs, addr)
	}

	return addrs, nil
}
