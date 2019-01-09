package blockchain

import (
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode"

	"github.com/ethereum/go-ethereum/common"
)

type Event struct {
	Data         interface{}
	BlockNumber  uint64
	TxIndex      uint64
	ReceiptIndex uint64
	TS           uint64
}

func (m *Event) PrecedesOrEquals(other *Event) bool {
	var comparators = []func(x, y *Event) bool{
		func(x, y *Event) bool { return x.BlockNumber < y.BlockNumber },
		func(x, y *Event) bool { return x.TxIndex < y.TxIndex },
		func(x, y *Event) bool { return x.ReceiptIndex <= y.ReceiptIndex },
	}
	for _, comparator := range comparators {
		switch {
		case comparator(m, other):
			return true
		case comparator(other, m):
			return false
		}
	}

	return false
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

type CertificateUpdatedData struct {
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

type Balance struct {
	Eth *big.Int
	SNM *big.Int
}

type GasPrice struct {
	*big.Int
}

func (m *GasPrice) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return fmt.Errorf("empty input")
	}

	// Remove all whitespace characters.
	normalized := strings.Replace(string(text), " ", "", -1)

	var number, unit string
	id := strings.IndexFunc(normalized, unicode.IsLetter)
	if id == -1 {
		number, unit = normalized, "wei"
	} else {
		number, unit = normalized[:id], normalized[id:]
	}

	value, ok := big.NewInt(0).SetString(number, 10)
	if !ok {
		return fmt.Errorf("failed to parse numeric value")
	}

	unitETH, err := UnitFromString(unit)
	if err != nil {
		return err
	}

	m.Int = big.NewInt(0).Mul(value, unitETH.Int)

	return nil
}

type PayoutResult int

const (
	UNKNOWN PayoutResult = iota
	Committed
	Payouted
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
	CommitTS time.Time
	Paid     bool
	Keeper   common.Address
}

type Keeper struct {
	Address    common.Address
	DayLimit   *big.Int
	LastDay    time.Time
	SpentToday *big.Int
	Frozen     bool
}
