package miner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareAddr(t *testing.T) {
	addr1 := "1000000000000000000000000000000000000001"
	addr2 := "1000000000000000000000000000000000000002"
	addr3 := "1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1"
	addr4 := "1AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA1"

	assert.False(t, compareEthAddr(addr1, addr2))
	assert.True(t, compareEthAddr(addr1, addr1))
	assert.True(t, compareEthAddr(addr3, addr4))

	assert.False(t, compareEthAddr("0x"+addr1, "0x"+addr2))
	assert.True(t, compareEthAddr("0x"+addr1, "0x"+addr1))
	assert.True(t, compareEthAddr("0x"+addr1, addr1))
	assert.True(t, compareEthAddr(addr1, "0x"+addr1))

	assert.True(t, compareEthAddr("0x"+addr3, "0x"+addr4))
	assert.True(t, compareEthAddr("0x"+addr3, "0x"+addr3))
	assert.True(t, compareEthAddr("0x"+addr4, addr4))
	assert.True(t, compareEthAddr(addr3, "0x"+addr3))
}
