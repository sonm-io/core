package netutil

import (
	"errors"
	"fmt"
	"net"
	"sort"
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

// TCPAddr wraps net.TCPAddr allowing to initialize itself using YAML
// unmarshaller.
type TCPAddr struct {
	net.TCPAddr
}

func (m *TCPAddr) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var addr string
	if err := unmarshal(&addr); err != nil {
		return err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return fmt.Errorf("cannot convert `%s` into a TCP address: %s", addr, err)
	}

	m.TCPAddr = *tcpAddr
	return nil
}

func GetPublicIPs() ([]net.IP, error) {
	available, err := GetAvailableIPs()
	if err != nil {
		return nil, err
	}

	var out []net.IP
	for _, ip := range available {
		if !IsPrivateIP(ip) {
			out = append(out, ip)
		}
	}

	return out, nil
}

func GetAvailableIPs() (availableIPs []net.IP, err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.IsGlobalUnicast() {
				availableIPs = append(availableIPs, ip)
			}
		}
	}
	if len(availableIPs) == 0 {
		return nil, errors.New("could not determine a single unicast addr, check networking")
	}

	return availableIPs, nil
}

func IsPrivateIP(ip net.IP) bool {
	return isPrivateIPv4(ip) || isPrivateIPv6(ip)
}

func isPrivateIPv4(ip net.IP) bool {
	_, private24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
	_, private20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
	_, private16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")
	return private24BitBlock.Contains(ip) ||
		private20BitBlock.Contains(ip) ||
		private16BitBlock.Contains(ip) ||
		ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast()
}

func isPrivateIPv6(ip net.IP) bool {
	_, block, _ := net.ParseCIDR("fc00::/7")

	return block.Contains(ip) || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

func LookupTCPHostPort(hostport string) ([]net.Addr, error) {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return nil, err
	}

	addrs, err := net.LookupHost(host)
	if err != nil {
		return nil, err
	}

	netAddrs := make([]net.Addr, len(addrs))
	for id, addr := range addrs {
		tcpAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(addr, port))
		if err != nil {
			return nil, err
		}

		netAddrs[id] = tcpAddr
	}

	return netAddrs, nil
}

func SortedIPs(ips []string) []string {
	var sorted sortableIPs
	for _, strIP := range ips {
		if ip := net.ParseIP(strIP); ip != nil {
			sorted = append(sorted, ip)
		}
	}
	sort.Sort(sorted)

	out := make([]string, len(sorted))
	for idx, ip := range sorted {
		out[idx] = ip.String()
	}

	return out
}

// Sorting is implemented as follows: first come all public IPs (IPv6 before IPv4), then
// all private IPs (IPv6 before IPv4).
type sortableIPs []net.IP

func (s sortableIPs) Len() int      { return len(s) }
func (s sortableIPs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortableIPs) Less(i, j int) bool {
	iIsPrivate, jIsPrivate := IsPrivateIP(s[i]), IsPrivateIP(s[j])
	if iIsPrivate && !jIsPrivate {
		return false
	}

	if jIsPrivate && !iIsPrivate {
		return true
	}

	// Both are private, check for family.
	iIsIPv4, jIsIPv4 := IsIPv4(s[i]), IsIPv4(s[j])
	if iIsIPv4 && !jIsIPv4 {
		return false
	}

	return true
}

func IsIPv4(ip net.IP) bool {
	return len(ip) == net.IPv4len
}
