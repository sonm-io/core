package netutil

import (
	"net"
	"strconv"
)

// TODO
type Port uint16

// TODO
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
