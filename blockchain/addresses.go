package blockchain

import "github.com/ethereum/go-ethereum/common"

const (
	sidechainSNMAddress          = "0x9b39604491deb0ce21044c272a84f7dc8138348c"
	masterchainSNMAddress        = "0xa2498b16a8fe7cd997f278d2419e3aa3b2b5854c"
	blacklistAddress             = "0xb8e23718cc2d56ae03439ae96204e90c0733f6d6"
	marketAddress                = "0x1bb0121a2ecd06d6a89b9a31748abbb1e532b75a"
	profileRegistryAddress       = "0x34a02e63b85eeaca27abc1178bdde0b06df2aee9"
	oracleUsdAddress             = "0x72818062eb6fe79d716d85b0620e2d59bcca4a8b"
	gatekeeperMasterchainAddress = "0x59b4b59eade970c9809044024453c2c43ff9e7b1"
	gatekeeperSidechainAddress   = "0xfecd969d3fb7347b784793a8954876908409700c"
	testnetFaucetAddress         = "0xeb031a9bb700fb609147d999de038ccfd9415def"
)

func MasterchainSNMAddr() common.Address {
	return common.HexToAddress(masterchainSNMAddress)
}

func SidechainSNMAddr() common.Address {
	return common.HexToAddress(sidechainSNMAddress)
}

func BlacklistAddr() common.Address {
	return common.HexToAddress(blacklistAddress)
}

func MarketAddr() common.Address {
	return common.HexToAddress(marketAddress)
}

func ProfileRegistryAddr() common.Address {
	return common.HexToAddress(profileRegistryAddress)
}

func OracleUsdAddr() common.Address {
	return common.HexToAddress(oracleUsdAddress)
}

func GatekeeperSidechainAddr() common.Address {
	return common.HexToAddress(gatekeeperSidechainAddress)
}

func GatekeeperMasterchainAddr() common.Address {
	return common.HexToAddress(gatekeeperMasterchainAddress)
}

func TestnetFaucetAddr() common.Address {
	return common.HexToAddress(testnetFaucetAddress)
}
