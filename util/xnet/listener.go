package xnet

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const (
	minSleepInterval = 5 * time.Millisecond
	maxSleepInterval = 1 * time.Second
	sleepMultiplier  = 2
)

type BackPressureListener struct {
	net.Listener

	Log *zap.Logger
}

func (m *BackPressureListener) Accept() (net.Conn, error) {
	interval := minSleepInterval

	for {
		conn, err := m.Listener.Accept()
		if err == nil {
			return conn, nil
		}

		if netError, ok := err.(net.Error); ok && netError.Temporary() {
			if max := maxSleepInterval; interval > max {
				interval = max
			}

			m.Log.Warn("failed to accept connection", zap.Error(netError), zap.Duration("sleep", interval))
			time.Sleep(interval)

			interval *= sleepMultiplier
		} else {
			return nil, err
		}
	}
}

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
