package dwh

import (
	"database/sql"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type storage interface {
	CreateIndices(db *sql.DB) error
	InsertDeal(conn queryConn, deal *pb.Deal) error
	UpdateDeal(conn queryConn, deal *pb.Deal) error
	UpdateDealsSupplier(conn queryConn, profile *pb.Profile) error
	UpdateDealsConsumer(conn queryConn, profile *pb.Profile) error
	UpdateDealPayout(conn queryConn, dealID, payout *big.Int, billTS uint64) error
	DeleteDeal(conn queryConn, dealID *big.Int) error
	GetDealByID(conn queryConn, dealID *big.Int) (*pb.DWHDeal, error)
	GetDeals(conn queryConn, request *pb.DealsRequest) ([]*pb.DWHDeal, uint64, error)
	GetDealConditions(conn queryConn, request *pb.DealConditionsRequest) ([]*pb.DealCondition, uint64, error)
	InsertOrder(conn queryConn, order *pb.DWHOrder) error
	UpdateOrderStatus(conn queryConn, orderID *big.Int, status pb.OrderStatus) error
	UpdateOrders(conn queryConn, profile *pb.Profile) error
	DeleteOrder(conn queryConn, orderID *big.Int) error
	GetOrderByID(conn queryConn, orderID *big.Int) (*pb.DWHOrder, error)
	GetOrders(conn queryConn, request *pb.OrdersRequest) ([]*pb.DWHOrder, uint64, error)
	GetMatchingOrders(conn queryConn, request *pb.MatchingOrdersRequest) ([]*pb.DWHOrder, uint64, error)
	GetProfiles(conn queryConn, request *pb.ProfilesRequest) ([]*pb.Profile, uint64, error)
	InsertDealChangeRequest(conn queryConn, changeRequest *pb.DealChangeRequest) error
	UpdateDealChangeRequest(conn queryConn, changeRequest *pb.DealChangeRequest) error
	DeleteDealChangeRequest(conn queryConn, changeRequestID *big.Int) error
	GetDealChangeRequests(conn queryConn, changeRequest *pb.DealChangeRequest) ([]*pb.DealChangeRequest, error)
	GetDealChangeRequestsByDealID(conn queryConn, changeRequestID *big.Int) ([]*pb.DealChangeRequest, error)
	InsertDealCondition(conn queryConn, condition *pb.DealCondition) error
	UpdateDealConditionPayout(conn queryConn, dealConditionID uint64, payout *big.Int) error
	UpdateDealConditionEndTime(conn queryConn, dealConditionID, eventTS uint64) error
	CheckWorkerExists(conn queryConn, masterID, workerID common.Address) (bool, error)
	InsertWorker(conn queryConn, masterID, workerID common.Address) error
	UpdateWorker(conn queryConn, masterID, workerID common.Address) error
	DeleteWorker(conn queryConn, masterID, workerID common.Address) error
	GetMasterByWorker(conn queryConn, workerID common.Address) (common.Address, error)
	InsertBlacklistEntry(conn queryConn, adderID, addeeID common.Address) error
	DeleteBlacklistEntry(conn queryConn, removerID, removeeID common.Address) error
	GetBlacklist(conn queryConn, request *pb.BlacklistRequest) (*pb.BlacklistReply, error)
	GetBlacklistsContainingUser(conn queryConn, request *pb.BlacklistRequest) (*pb.BlacklistsContainingUserReply, error)
	InsertOrUpdateValidator(conn queryConn, validator *pb.Validator) error
	UpdateValidator(conn queryConn, validator *pb.Validator) error
	InsertCertificate(conn queryConn, certificate *pb.Certificate) error
	GetCertificates(conn queryConn, ownerID common.Address) ([]*pb.Certificate, error)
	InsertProfileUserID(conn queryConn, profile *pb.Profile) error
	GetProfileByID(conn queryConn, userID common.Address) (*pb.Profile, error)
	GetValidators(conn queryConn, request *pb.ValidatorsRequest) ([]*pb.Validator, uint64, error)
	GetWorkers(conn queryConn, request *pb.WorkersRequest) ([]*pb.DWHWorker, uint64, error)
	UpdateProfile(conn queryConn, userID common.Address, field string, value interface{}) error
	UpdateProfileStats(conn queryConn, userID common.Address, field string, value int) error
	GetLastKnownBlock(conn queryConn) (uint64, error)
	InsertLastKnownBlock(conn queryConn, blockNumber int64) error
	UpdateLastKnownBlock(conn queryConn, blockNumber int64) error
	StoreStaleID(conn queryConn, id *big.Int, entity string) error
	RemoveStaleID(conn queryConn, id *big.Int, entity string) error
	CheckStaleID(conn queryConn, id *big.Int, entity string) (bool, error)
}

type queryConn interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	// Any time a queryConn is created, `defer conn.Finish()` should be called.
	Finish() error
}

type simpleConn struct {
	db *sql.DB
}

func newSimpleConn(db *sql.DB) queryConn {
	return &simpleConn{db: db}
}

func (t *simpleConn) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.db.Exec(query, args...)
}

func (t *simpleConn) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.db.Query(query, args...)
}

func (t *simpleConn) Finish() error {
	return nil
}

// txConn implements covert transaction rollbacks/commits based on whether there was any errors
// while interacting with DB.
type txConn struct {
	tx        *sql.Tx
	logger    *zap.Logger
	hasErrors bool
}

func newTxConn(db *sql.DB, logger *zap.Logger) (queryConn, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &txConn{tx: tx, logger: logger}, nil
}

func (t *txConn) Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := t.tx.Exec(query, args...)
	if err != nil {
		t.hasErrors = true
		return nil, errors.Wrapf(err, "failed to exec %s", query)
	}
	return result, nil
}

func (t *txConn) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := t.tx.Query(query, args...)
	if err != nil {
		t.hasErrors = true
		return nil, errors.Wrapf(err, "failed to run %s", query)
	}
	return rows, nil
}

func (t *txConn) Finish() error {
	if t.hasErrors {
		if err := t.tx.Rollback(); err != nil {
			t.logger.Warn("transaction rollback failed")
			return err
		}
	} else {
		if err := t.tx.Commit(); err != nil {
			t.logger.Warn("transaction rollback failed")
			return err
		}
	}
	return nil
}
