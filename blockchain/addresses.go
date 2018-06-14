package blockchain

import "github.com/ethereum/go-ethereum/common"

// private chain
const SNMSidechainAddress string = "0x26524b1234e361eb4e3ddf7600d41271620fcb0a"

const BlacklistAddress string = "0x9ad1e969ec5842ee5d67414536813e224ceb56b1"

const MarketAddress string = "0x51a1d1d1821b4109eb84edc4ca8ec814b1fe9876"

const ProfileRegistryAddress string = "0x1b3a50ee228b040e1b00ef7e7f99058be2684408"

const OracleUsdAddress string = "0x1f995e52dcbec7c0d00d45b8b1bf43b29dd12b5b"

const GatekeeperMasterchainAddress string = "0xbc29310be3693949094ce452b11829dbccca7d49"

const GatekeeperSidechainAddress string = "0x9414922e778a0038058e9ea786e9474a89ad1ec0"

// rinkeby
const SNMAddress string = "0x06bda3cf79946e8b32a0bb6a3daa174b577c55b5"

func SNMAddr() common.Address {
	return common.HexToAddress(SNMAddress)
}

func BlacklistAddr() common.Address {
	return common.HexToAddress(BlacklistAddress)
}

func MarketAddr() common.Address {
	return common.HexToAddress(MarketAddress)
}

func ProfileRegistryAddr() common.Address {
	return common.HexToAddress(ProfileRegistryAddress)
}

func OracleUsdAddr() common.Address {
	return common.HexToAddress(OracleUsdAddress)
}

func SNMSidechainAddr() common.Address {
	return common.HexToAddress(SNMSidechainAddress)
}

func GatekeeperSidechainAddr() common.Address {
	return common.HexToAddress(GatekeeperSidechainAddress)
}

func GatekeeperMasterchainAddr() common.Address {
	return common.HexToAddress(GatekeeperMasterchainAddress)
}
