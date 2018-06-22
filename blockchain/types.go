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

type MultiSigTransactionData struct {
	To       common.Address
	Value    *big.Int
	Data     []byte
	Executed bool
}

type ConfirmationData struct {
	Sender        common.Address
	TransactionId *big.Int
}

type RevocationData struct {
	Sender        common.Address
	TransactionId *big.Int
}

type SubmissionData struct {
	TransactionId *big.Int
}

type ExecutionData struct {
	TransactionId *big.Int
}

type ExecutionFailureData struct {
	TransactionId *big.Int
}

type OwnerAdditionData struct {
	Owner common.Address
}

type OwnerRemovalData struct {
	Owner common.Address
}

type SimpleGateTransaction struct {
	To       common.Address
	Value    *big.Int
	TxNumber *big.Int
}
