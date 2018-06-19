package blockchain

import "github.com/ethereum/go-ethereum/common"

const (
	sidechainSNMAddress          = "0x26524b1234e361eb4e3ddf7600d41271620fcb0a"
	masterchainSNMAddress        = "0x06bda3cf79946e8b32a0bb6a3daa174b577c55b5"
	blacklistAddress             = "0x9ad1e969ec5842ee5d67414536813e224ceb56b1"
	marketAddress                = "0x92d964348b59aaefc467123a593bb76289da2ba1"
	profileRegistryAddress       = "0xc12c8a95d3b3436952719f0c9b662d14b7bd3a12"
	oracleUsdAddress             = "0x1f995e52dcbec7c0d00d45b8b1bf43b29dd12b5b"
	gatekeeperMasterchainAddress = "0xbc29310be3693949094ce452b11829dbccca7d49"
	gatekeeperSidechainAddress   = "0x9414922e778a0038058e9ea786e9474a89ad1ec0"
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
