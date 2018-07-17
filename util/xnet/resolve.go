package xnet

import "net"

// LookupLoopbackIP looks up loopback interfaces on the host using the local
// resolver.
// It returns a slice of that host's IPv6 and IPv4 addresses.
func LookupLoopbackIP() ([]net.IP, error) {
	// Possible implementation may use netlink to avoid using local
	// resolver.
	return net.LookupIP("localhost")
}
