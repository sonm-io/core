package antifraud

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestIsWhitelisted(t *testing.T) {
	a := &antiFraud{cfg: Config{
		Whitelist: []common.Address{
			common.HexToAddress("0x38FeB5FE6fb1EECb2e0b990214BC6f0ACa1eCE6A"),
			common.HexToAddress("0xeEDC24cdcA34fcDFCeaAd84a7C6bA3301D5D707f"),
			common.HexToAddress("0xBAdc46A8ca03bD0D458242E76Bc2810DC5f12908"),
		},
	}}

	notInList := a.isAddressWhitelisted(common.HexToAddress("0x2b7D9Af99CE1Bec0dc32aB115B4DA82BF15B501E"))
	assert.False(t, notInList)

	inList := a.isAddressWhitelisted(common.HexToAddress("0xeEDC24cdcA34fcDFCeaAd84a7C6bA3301D5D707f"))
	assert.True(t, inList)
}
