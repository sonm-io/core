package blockchain

import "fmt"

const (
	defaultContractRegistryAddr = "0xd1a6f3d1ae33b4b19565a6b283d7a05c5a0decb0"

	sidechainSNMAddressKey          = "sidechainSNMAddress"
	masterchainSNMAddressKey        = "masterchainSNMAddress"
	blacklistAddressKey             = "blacklistAddress"
	profileRegistryAddressKey       = "profileRegistryAddress"
	oracleUsdAddressKey             = "oracleUsdAddress"
	gatekeeperMasterchainAddressKey = "gatekeeperMasterchainAddress"
	gatekeeperSidechainAddressKey   = "gatekeeperSidechainAddress"
	testnetFaucetAddressKey         = "testnetFaucetAddress"
	oracleMultiSigAddressKey        = "oracleMultiSigAddress"
	devicesStorageAddressKey        = "devicesStorageAddress"
)

func marketAddressKey(version uint) (string, error) {
	switch version {
	case 1:
		return "marketAddress", nil
	case 2:
		return "marketV2Address", nil
	default:
		return "", fmt.Errorf("invalid market version %d", version)
	}
}
