package xnet

import (
	"net"
)

// LookupLoopbackIP looks up loopback interfaces on the host using the local
// resolver.
// It returns a slice of that host's IPv6 and IPv4 addresses.
func LookupLoopbackIP() ([]net.IP, error) {
	links, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	// Distinguishing between IPv6 and IPv4 addresses if required to stick to
	// the RFC 6724, which requires IPv6 to be before IPv4.
	var loV6Addrs []net.IP
	var loV4Addrs []net.IP

	for _, link := range links {
		// Ignore down and non-loopback links. However, link-local links
		// like under the "fe80::/10" subnet for example still need to be
		// filtered out later.
		if link.Flags&(net.FlagUp|net.FlagLoopback) != (net.FlagUp | net.FlagLoopback) {
			continue
		}

		addrs, err := link.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			if ip.IsLoopback() {
				if ip := ip.To16(); ip != nil {
					loV6Addrs = append(loV6Addrs, ip)
					continue
				}

				if ip := ip.To4(); ip != nil {
					loV4Addrs = append(loV4Addrs, ip)
					continue
				}
			}
		}
	}

	return append(loV6Addrs, loV4Addrs...), nil
}
