package dwh

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"sync"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

const (
	MaxLimit         = 200
	NumMaxBenchmarks = 100
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
	orderedSetupCommands = []string{
		"createTableDeals",
		"createTableDealConditions",
		"createTableDealPayments",
		"createTableChangeRequests",
		"createTableOrders",
		"createTableWorkers",
		"createTableBlacklists",
		"createTableValidators",
		"createTableCertificates",
		"createTableProfiles",
		"createTableMisc",
	}
	finalizeColumnsOnce  = &sync.Once{}
	finalizeCommandsOnce = &sync.Once{}
)

var (
	DealColumns = []string{
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
	DealColumnsSet = stringSliceToSet(DealColumns)
	NumDealColumns = len(DealColumns)
	OrderColumns   = []string{
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
	OrderColumnsSet         = stringSliceToSet(OrderColumns)
	NumOrderColumns         = len(OrderColumns)
	DealConditionColumnsSet = map[string]bool{
		"Id":          true,
		"SupplierID":  true,
		"ConsumerID":  true,
		"MasterID":    true,
		"Duration":    true,
		"Price":       true,
		"StartTime":   true,
		"EndTime":     true,
		"TotalPayout": true,
		"DealID":      true,
	}
	ProfilesColumnsSet = map[string]bool{
		"Id":             true,
		"UserID":         true,
		"IdentityLevel":  true,
		"Name":           true,
		"Country":        true,
		"IsCorporation":  true,
		"IsProfessional": true,
		"Certificates":   true,
	}
)

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

type QueryRunner func(db *sql.DB, opts *queryOpts) (*sql.Rows, uint64, error)

func getBenchmarkColumn(id uint64) string {
	return fmt.Sprintf("Benchmark%d", id)
}

type setupIndices func(w *DWH) error

func coldStart(w *DWH, setupIndicesCb setupIndices) {
	if w.cfg.ColdStart.UpToBlock < 1 {
		w.logger.Info("UpToBlock < 1, creating indices right now")
		if err := setupIndicesCb(w); err != nil {
			w.logger.Error("failed to setupIndicesCb", zap.Error(err))
			os.Exit(1)
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
				w.logger.Info("failed to getLastKnownBlockTS (coldStart)")
				continue
			}
			w.logger.Info("current block (waiting to create indices)", zap.Uint64("block_number", lastBlock))
			if lastBlock >= w.cfg.ColdStart.UpToBlock {
				w.logger.Info("creating indices")
				if err := setupIndicesCb(w); err != nil {
					w.logger.Error("failed to setupIndicesCb (coldStart)", zap.Error(err))
				}
			}
		}
	}
}

func finalizeTableColumns(numBenchmarks int) {
	for benchmarkID := 0; benchmarkID < numBenchmarks; benchmarkID++ {
		DealColumns = append(DealColumns, getBenchmarkColumn(uint64(benchmarkID)))
		DealColumnsSet[getBenchmarkColumn(uint64(benchmarkID))] = true
		OrderColumns = append(OrderColumns, getBenchmarkColumn(uint64(benchmarkID)))
		OrderColumnsSet[getBenchmarkColumn(uint64(benchmarkID))] = true
	}
}
