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
