package dwh

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

const (
	MaxLimit         = 200
	NumMaxBenchmarks = 128
)

const (
	gte = ">="
	lte = "<="
	eq  = "="
)

var (
	opsTranslator = map[pb.CmpOp]string{
		pb.CmpOp_GTE: gte,
		pb.CmpOp_LTE: lte,
		pb.CmpOp_EQ:  eq,
	}
	setupDBCallbacks = map[string]func(*DWH) error{
		"sqlite3":  setupSQLite,
		"postgres": setupPostgres,
	}
)

type QueryRunner interface {
	Run(opts *queryOpts) (*sql.Rows, uint64, error)
}

type setupIndices func(w *DWH) error

type tablesInfo struct {
	DealColumns             []string
	DealColumnsSet          map[string]bool
	NumDealColumns          uint64
	OrderColumns            []string
	OrderColumnsSet         map[string]bool
	NumOrderColumns         uint64
	ProfileColumnsSet       map[string]bool
	DealConditionColumnsSet map[string]bool
}

func newTablesInfo(numBenchmarks uint64) *tablesInfo {
	dealColumns := []string{
		"Id",
		"SupplierID",
		"ConsumerID",
		"MasterID",
		"AskID",
		"BidID",
		"Duration",
		"Price",
		"StartTime",
		"EndTime",
		"Status",
		"BlockedBalance",
		"TotalPayout",
		"LastBillTS",
		"Netflags",
		"AskIdentityLevel",
		"BidIdentityLevel",
		"SupplierCertificates",
		"ConsumerCertificates",
		"ActiveChangeRequest",
	}
	orderColumns := []string{
		"Id",
		"CreatedTS",
		"DealID",
		"Type",
		"Status",
		"AuthorID",
		"CounterpartyID",
		"Duration",
		"Price",
		"Netflags",
		"IdentityLevel",
		"Blacklist",
		"Tag",
		"FrozenSum",
		"CreatorIdentityLevel",
		"CreatorName",
		"CreatorCountry",
		"CreatorCertificates",
	}
	dealConditionColumns := []string{
		"Id",
		"SupplierID",
		"ConsumerID",
		"MasterID",
		"Duration",
		"Price",
		"StartTime",
		"EndTime",
		"TotalPayout",
		"DealID",
	}
	profileColumns := []string{
		"Id",
		"UserID",
		"IdentityLevel",
		"Name",
		"Country",
		"IsCorporation",
		"IsProfessional",
		"Certificates",
	}

	out := &tablesInfo{
		DealColumns:             dealColumns,
		DealColumnsSet:          stringSliceToSet(dealColumns),
		NumDealColumns:          uint64(len(dealColumns)),
		OrderColumns:            orderColumns,
		OrderColumnsSet:         stringSliceToSet(orderColumns),
		NumOrderColumns:         uint64(len(orderColumns)),
		DealConditionColumnsSet: stringSliceToSet(dealConditionColumns),
		ProfileColumnsSet:       stringSliceToSet(profileColumns),
	}

	for benchmarkID := uint64(0); benchmarkID < numBenchmarks; benchmarkID++ {
		out.DealColumns = append(out.DealColumns, getBenchmarkColumn(uint64(benchmarkID)))
		out.DealColumnsSet[getBenchmarkColumn(uint64(benchmarkID))] = true
		out.OrderColumns = append(out.OrderColumns, getBenchmarkColumn(uint64(benchmarkID)))
		out.OrderColumnsSet[getBenchmarkColumn(uint64(benchmarkID))] = true
	}

	return out
}

type SQLCommands struct {
	insertDeal                   string
	updateDeal                   string
	updateDealsSupplier          string
	updateDealsConsumer          string
	updateDealPayout             string
	selectDealByID               string
	deleteDeal                   string
	insertOrder                  string
	selectOrderByID              string
	updateOrderStatus            string
	updateOrders                 string
	deleteOrder                  string
	insertDealChangeRequest      string
	selectDealChangeRequests     string
	selectDealChangeRequestsByID string
	deleteDealChangeRequest      string
	updateDealChangeRequest      string
	insertDealCondition          string
	updateDealConditionPayout    string
	updateDealConditionEndTime   string
	insertDealPayment            string
	insertWorker                 string
	updateWorker                 string
	deleteWorker                 string
	insertBlacklistEntry         string
	selectBlacklists             string
	deleteBlacklistEntry         string
	insertValidator              string
	updateValidator              string
	insertCertificate            string
	selectCertificates           string
	insertProfileUserID          string
	selectProfileByID            string
	profileNotInBlacklist        string
	profileInBlacklist           string
	updateProfile                string
	selectLastKnownBlock         string
	insertLastKnownBlock         string
	updateLastKnownBlock         string
}

type SQLSetupCommands struct {
	createTableDeals          string
	createTableDealConditions string
	createTableDealPayments   string
	createTableChangeRequests string
	createTableOrders         string
	createTableWorkers        string
	createTableBlacklists     string
	createTableValidators     string
	createTableCertificates   string
	createTableProfiles       string
	createTableMisc           string
}

func (c *SQLSetupCommands) Exec(db *sql.DB) error {
	_, err := db.Exec(c.createTableDeals)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableDeals)
	}

	_, err = db.Exec(c.createTableDealConditions)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableDealConditions)
	}

	_, err = db.Exec(c.createTableDealPayments)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableDealPayments)
	}

	_, err = db.Exec(c.createTableChangeRequests)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableChangeRequests)
	}

	_, err = db.Exec(c.createTableOrders)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableOrders)
	}

	_, err = db.Exec(c.createTableWorkers)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableWorkers)
	}

	_, err = db.Exec(c.createTableBlacklists)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableBlacklists)
	}

	_, err = db.Exec(c.createTableValidators)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableValidators)
	}

	_, err = db.Exec(c.createTableCertificates)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableCertificates)
	}

	_, err = db.Exec(c.createTableProfiles)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableProfiles)
	}

	_, err = db.Exec(c.createTableMisc)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableMisc)
	}

	return nil
}

type filter struct {
	Field        string
	CmpOperator  string
	BoolOperator string
	OpenBracket  bool
	CloseBracket bool
	Value        interface{}
}

func newFilter(field string, cmpOperator string, value interface{}, boolOperator string) *filter {
	return &filter{
		Field:        field,
		CmpOperator:  cmpOperator,
		BoolOperator: boolOperator,
		Value:        value,
	}
}

func newNetflagsFilter(operator pb.CmpOp, value uint64) *filter {
	switch operator {
	case pb.CmpOp_GTE:
		return newFilter("Netflags", fmt.Sprintf(" | ~%d = ", value), -1, "AND")
	case pb.CmpOp_LTE:
		return newFilter("", fmt.Sprintf("%d | ~Netflags = ", value), -1, "AND")
	default:
		return newFilter("Netflags", eq, value, "AND")
	}
}

type customFilter struct {
	clause string
	values []interface{}
}

type queryOpts struct {
	table        string
	filters      []*filter
	sortings     []*pb.SortingOption
	offset       uint64
	limit        uint64
	customFilter *customFilter
	selectAs     string
	withCount    bool
}

func createIndex(db *sql.DB, command, table, column string) error {
	cmd := fmt.Sprintf(command, table, column, table, column)
	_, err := db.Exec(cmd)
	if err != nil {
		return errors.Wrapf(err, "failed to %s (%s)", cmd)
	}

	return nil
}

func filterSortings(sortings []*pb.SortingOption, columns map[string]bool) (out []*pb.SortingOption) {
	for _, sorting := range sortings {
		if columns[sorting.Field] {
			out = append(out, sorting)
		}
	}

	return out
}

func coldStart(w *DWH, setupIndicesCb setupIndices) {
	if w.cfg.ColdStart.UpToBlock == 0 {
		w.logger.Info("UpToBlock == 0, creating indices right now")
		if err := setupIndicesCb(w); err != nil {
			w.logger.Error("failed to setupIndicesCb, exiting", zap.Error(err))
			w.Stop()
		} else {
			w.logger.Info("successfully created indices")
		}

		return
	}
	w.logger.Info("creating indices after block", zap.Uint64("block_number", w.cfg.ColdStart.UpToBlock))
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("stopped coldStart routine")
			return
		case <-ticker.C:
			lastBlock, err := w.getLastKnownBlockTS()
			if err != nil {
				w.logger.Info("failed to getLastKnownBlockTS (coldStart), retrying")
				continue
			}
			w.logger.Info("current block (waiting to create indices)", zap.Uint64("block_number", lastBlock))
			if lastBlock >= w.cfg.ColdStart.UpToBlock {
				w.logger.Info("creating indices")
				if err := setupIndicesCb(w); err != nil {
					w.logger.Error("failed to setupIndicesCb (coldStart), retrying", zap.Error(err))
					continue
				}
				w.logger.Info("successfully created indices")
				return
			}
		}
	}
}
