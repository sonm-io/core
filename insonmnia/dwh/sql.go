package dwh

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
)

const (
	MaxLimit = 200
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
)

var (
	OrdersColumns = map[string]bool{
		"Id":                   true,
		"CreatedTS":            true,
		"DealID":               true,
		"Type":                 true,
		"Status":               true,
		"AuthorID":             true,
		"CounterpartyID":       true,
		"Duration":             true,
		"Price":                true,
		"Netflags":             true,
		"IdentityLevel":        true,
		"Blacklist":            true,
		"Tag":                  true,
		"FrozenSum":            true,
		"CreatorIdentityLevel": true,
		"CreatorName":          true,
		"CreatorCountry":       true,
		"CreatorCertificates":  true,
		"CPUSysbenchMulti":     true,
		"CPUSysbenchOne":       true,
		"CPUCores":             true,
		"RAMSize":              true,
		"StorageSize":          true,
		"NetTrafficIn":         true,
		"NetTrafficOut":        true,
		"GPUCount":             true,
		"GPUMem":               true,
		"GPUEthHashrate":       true,
		"GPUCashHashrate":      true,
		"GPURedshift":          true,
	}
	DealsColumns = map[string]bool{
		"Id":                   true,
		"SupplierID":           true,
		"ConsumerID":           true,
		"MasterID":             true,
		"AskID":                true,
		"BidID":                true,
		"Duration":             true,
		"Price":                true,
		"StartTime":            true,
		"EndTime":              true,
		"Status":               true,
		"BlockedBalance":       true,
		"TotalPayout":          true,
		"LastBillTS":           true,
		"Netflags":             true,
		"AskIdentityLevel":     true,
		"BidIdentityLevel":     true,
		"SupplierCertificates": true,
		"ConsumerCertificates": true,
		"ActiveChangeRequest":  true,
		"CPUSysbenchMulti":     true,
		"CPUSysbenchOne":       true,
		"CPUCores":             true,
		"RAMSize":              true,
		"StorageSize":          true,
		"NetTrafficIn":         true,
		"NetTrafficOut":        true,
		"GPUCount":             true,
		"GPUMem":               true,
		"GPUEthHashrate":       true,
		"GPUCashHashrate":      true,
		"GPURedshift":          true,
	}
	DealConditionsColumns = map[string]bool{
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
	ProfilesColumns = map[string]bool{
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
