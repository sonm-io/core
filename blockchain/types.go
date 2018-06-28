package blockchain

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Event struct {
	Data        interface{}
	BlockNumber uint64
	TS          uint64
}

type DealOpenedData struct {
	ID *big.Int
}

type DealUpdatedData struct {
	ID *big.Int
}

type DealChangeRequestSentData struct {
	ID *big.Int
}

type DealChangeRequestUpdatedData struct {
	ID *big.Int
}

type OrderPlacedData struct {
	ID *big.Int
}

type OrderUpdatedData struct {
	ID *big.Int
}

type BilledData struct {
	DealID     *big.Int `json:"dealID"`
	PaidAmount *big.Int `json:"paidAmount"`
}

type WorkerAnnouncedData struct {
	WorkerID common.Address
	MasterID common.Address
}

type WorkerConfirmedData struct {
	WorkerID common.Address
	MasterID common.Address
}

type WorkerRemovedData struct {
	WorkerID common.Address
	MasterID common.Address
}

type ErrorData struct {
	Err   error
	Topic string
}

type AddedToBlacklistData struct {
	AdderID common.Address
	AddeeID common.Address
}

type RemovedFromBlacklistData struct {
	RemoverID common.Address
	RemoveeID common.Address
}

type ValidatorCreatedData struct {
	ID common.Address
}

type ValidatorDeletedData struct {
	ID common.Address
}

type CertificateCreatedData struct {
	ID *big.Int
}

type NumBenchmarksUpdatedData struct {
	NumBenchmarks uint64
}

type PayoutResult int

const (
	UNKNOWN   PayoutResult = 0
	Committed PayoutResult = 1
	Payouted  PayoutResult = 2
)

type GateTx struct {
	// From token transfer sender
	From common.Address
	// Number is sequence number of transaction
	// defines to unique transaction.
	// That sequence realized in smart contract
	Number *big.Int
	// Value of transferring tokens
	Value *big.Int
	// BlockNumber timestamp of commitment Payin transaction
	// used for calculate duration of stay transaction
	BlockNumber uint64
}

// GateTxState present state of payout transaction
// used for verify transactions
type GateTxState struct {
	CommitTS *big.Int
	Paid     bool
	Keeper   common.Address
}

type Keeper struct {
	Address    common.Address
	DayLimit   *big.Int
	LastDay    *big.Int
	SpentToday *big.Int
	Frozen     bool
}
