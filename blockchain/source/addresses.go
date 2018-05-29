package market

import "github.com/ethereum/go-ethereum/common"

// private chain
const SNMSidechainAddress string = "0x26524b1234e361eb4e3ddf7600d41271620fcb0a"

const GatekeeperLiveAddress string = "0xbc29310be3693949094ce452b11829dbccca7d49"

const GatekeeperSidechainAddress string = "0x9414922e778a0038058e9ea786e9474a89ad1ec0"

// rinkeby
const SNMAddress string = "0x06bda3cf79946e8b32a0bb6a3daa174b577c55b5"

func SNMAddr() common.Address {
	return common.HexToAddress(SNMAddress)
}

func SNMSidechainAddr() common.Address {
	return common.HexToAddress(SNMSidechainAddress)
}

func GatekeeperSidechainAddr() common.Address {
	return common.HexToAddress(GatekeeperSidechainAddress)
}

func GatekeeperLiveAddr() common.Address {
	return common.HexToAddress(GatekeeperLiveAddress)
}
