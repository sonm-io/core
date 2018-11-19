package xnet

import (
	"fmt"
	"net"
	"strconv"
)

// Listen announces on the loopback network address.
//
// The network must be "tcp", "tcp4" or "tcp6".
func ListenLoopback(network string, port uint16) ([]net.Listener, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("unexpected network type: %s", network)
	}

	ips, err := LookupLoopbackIP()
	if err != nil {
		return nil, err
	}

	onFail := func(listeners []net.Listener) {
		for _, listener := range listeners {
			listener.Close()
		}
	}

	var listeners []net.Listener
	for _, ip := range ips {
		listener, err := net.Listen(network, net.JoinHostPort(ip.String(), strconv.Itoa(int(port))))
		if err != nil {
			onFail(listeners)
			return nil, err
		}

		listeners = append(listeners, listener)
	}

	return listeners, nil
}

func ListenPacketLoopback(network string, port uint16) ([]net.PacketConn, error) {
	if network != "udp" && network != "udp4" && network != "udp6" {
		return nil, fmt.Errorf("unexpected network type: %s", network)
	}

	ips, err := LookupLoopbackIP()
	if err != nil {
		return nil, err
	}

	onFail := func(listeners []net.PacketConn) {
		for _, listener := range listeners {
			listener.Close()
		}
	}

	var listeners []net.PacketConn
	for _, ip := range ips {
		listener, err := net.ListenPacket(network, net.JoinHostPort(ip.String(), strconv.Itoa(int(port))))
		if err != nil {
			onFail(listeners)
			return nil, err
		}

		listeners = append(listeners, listener)
	}

	return listeners, nil
}
