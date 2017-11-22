package miner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
