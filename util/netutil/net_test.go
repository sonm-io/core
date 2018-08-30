package netutil

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPrivateIPv4(t *testing.T) {
	tests := []struct {
		ip     string
		isPriv bool
	}{
		{ip: "127.0.0.1", isPriv: true},
		{ip: "127.1.2.3", isPriv: true},
		{ip: "10.20.30.40", isPriv: true},
		{ip: "192.168.0.1", isPriv: true},
		{ip: "172.16.0.1", isPriv: true},
		{ip: "169.254.0.1", isPriv: true},
		{ip: "169.254.1.0", isPriv: true},
		{ip: "169.254.123.222", isPriv: true},
		{ip: "169.254.255.255", isPriv: true},
		{ip: "224.0.0.1", isPriv: true},

		{ip: "1.2.3.4", isPriv: false},
		{ip: "0.0.0.0", isPriv: false},
	}

	for _, cc := range tests {
		b := isPrivateIPv4(net.ParseIP(cc.ip))
		assert.Equal(t, b, cc.isPriv, cc.ip)
	}
}

func TestIsPrivateIPv6(t *testing.T) {
	tests := []struct {
		ip     string
		isPriv bool
	}{
		{ip: "::", isPriv: false},
		{ip: "::1", isPriv: true},
		{ip: "fe80::", isPriv: true},
		{ip: "feaf::", isPriv: true},
		{ip: "fc00::", isPriv: true},
	}

	for _, cc := range tests {
		b := isPrivateIPv6(net.ParseIP(cc.ip))
		assert.Equal(t, b, cc.isPriv, cc.ip)
	}
}

func TestSortedIPs(t *testing.T) {
	ips := []string{
		"192.168.70.17",
		"46.148.198.133",
		"fd21:f7bb:61b8:9e37::1",
		"2001:db8::68",
	}

	sortedIPs := []string{
		"2001:db8::68",
		"46.148.198.133",
		"fd21:f7bb:61b8:9e37::1",
		"192.168.70.17",
	}

	assert.Equal(t, sortedIPs, SortedIPs(ips))
}
