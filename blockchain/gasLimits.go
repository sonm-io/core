package blockchain

const (
	getTokensGasLimit           = 70000
	transferGasLimit            = 70000
	transferFromGasLimit        = 80000
	increaseApprovalGasLimit    = 70000
	decreaseApprovalGasLimit    = 70000
	approveGasLimit             = 70000
	placeOrderGasLimit          = 650000
	cancelOrderGasLimit         = 300000
	quickBuyGasLimit            = 1300000
	openDealGasLimit            = 800000
	closeDealGasLimit           = 250000
	billGasLimit                = 300000
	createChangeRequestGasLimit = 300000
	cancelChangeRequestGasLimit = 170000
	registerWorkerGasLimit      = 100000
	confirmWorkerGasLimit       = 100000
	removeWorkerGasLimit        = 100000
	addMasterGasLimit           = 100000
	removeMasterGasLimit        = 100000
	payinGasLimit               = 100000
	payoutGasLimit              = 100000
)

func devicesGasLimit(size uint64) uint64 {
	return 400000 + size/32*31000
}
