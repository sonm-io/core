package netutil

import (
	"net"
	"strconv"
)

// Port describes a transport layer port.
type Port uint16

// SplitHostPort splits a network address.
//
// A literal address or host name for IPv6 must be enclosed in square
// brackets, as in "[::1]:80".
//
// Unlike net.SplitHostPort this function also performs type transformations
// to represent an IP address as a commonly-used interface and a port as
// an uint16.
func SplitHostPort(hostport string) (net.IP, Port, error) {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return nil, 0, err
	}

	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, 0, err
	}

	return net.ParseIP(host), Port(p), nil
}

func ExtractHost(hostport string) (net.IP, error) {
	host, _, err := SplitHostPort(hostport)
	return host, err
}

func ExtractPort(hostport string) (Port, error) {
	_, port, err := SplitHostPort(hostport)
	return port, err
}
