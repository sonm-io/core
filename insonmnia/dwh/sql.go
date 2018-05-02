package dwh

import (
	"database/sql"
	"fmt"
	"strings"

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
		"sqlite3": setupSQLite,
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

var (
	sqliteSetupCommands = map[string]string{
		"createTableDeals": `
	CREATE TABLE IF NOT EXISTS Deals (
		Id						TEXT UNIQUE NOT NULL,
		SupplierID				TEXT NOT NULL,
		ConsumerID				TEXT NOT NULL,
		MasterID				TEXT NOT NULL,
		AskID					TEXT NOT NULL,
		BidID					TEXT NOT NULL,
		Duration 				INTEGER NOT NULL,
		Price					TEXT NOT NULL,
		StartTime				INTEGER NOT NULL,
		EndTime					INTEGER NOT NULL,
		Status					INTEGER NOT NULL,
		BlockedBalance			TEXT NOT NULL,
		TotalPayout				TEXT NOT NULL,
		LastBillTS				INTEGER NOT NULL,
		Netflags				INTEGER NOT NULL,
		AskIdentityLevel		INTEGER NOT NULL,
		BidIdentityLevel		INTEGER NOT NULL,
		SupplierCertificates    BLOB NOT NULL,
		ConsumerCertificates    BLOB NOT NULL,
		ActiveChangeRequest     INTEGER NOT NULL,
		CPUSysbenchMulti		INTEGER NOT NULL,
		CPUSysbenchOne			INTEGER NOT NULL,
		CPUCores				INTEGER NOT NULL,
		RAMSize					INTEGER NOT NULL,
		StorageSize				INTEGER NOT NULL,
		NetTrafficIn			INTEGER NOT NULL,
		NetTrafficOut			INTEGER NOT NULL,
		GPUCount				INTEGER NOT NULL,
		GPUMem					INTEGER NOT NULL,
		GPUEthHashrate			INTEGER NOT NULL,
		GPUCashHashrate			INTEGER NOT NULL,
		GPURedshift				INTEGER NOT NULL
	)`,
		"createTableDealConditions": `
	CREATE TABLE IF NOT EXISTS DealConditions (
		Id							INTEGER PRIMARY KEY AUTOINCREMENT,
		SupplierID					TEXT NOT NULL,
		ConsumerID					TEXT NOT NULL,
		MasterID					TEXT NOT NULL,
		Duration 					INTEGER NOT NULL,
		Price						TEXT NOT NULL,
		StartTime					INTEGER NOT NULL,
		EndTime						INTEGER NOT NULL,
		TotalPayout					TEXT NOT NULL,
		DealID						TEXT NOT NULL,
		FOREIGN KEY (DealID)		REFERENCES Deals(Id) ON DELETE CASCADE
	)`,
		"createTableDealPayments": `
	CREATE TABLE IF NOT EXISTS DealPayments (
		BillTS						INTEGER NOT NULL,
		PaidAmount					TEXT NOT NULL,
		DealID						TEXT NOT NULL,
		UNIQUE						(BillTS, PaidAmount, DealID),
		FOREIGN KEY (DealID) 		REFERENCES Deals(Id) ON DELETE CASCADE
	)`,
		"createTableChangeRequests": `
	CREATE TABLE IF NOT EXISTS DealChangeRequests (
		Id 							TEXT UNIQUE NOT NULL,
		CreatedTS					INTEGER NOT NULL,
		RequestType					TEXT NOT NULL,
		Duration 					INTEGER NOT NULL,
		Price						TEXT NOT NULL,
		Status						INTEGER NOT NULL,
		DealID						TEXT NOT NULL,
		FOREIGN KEY (DealID)		REFERENCES Deals(Id) ON DELETE CASCADE
	)`,
		"createTableOrders": `
	CREATE TABLE IF NOT EXISTS Orders (
		Id						TEXT UNIQUE NOT NULL,
		CreatedTS				INTEGER NOT NULL,
		DealID					TEXT NOT NULL,
		Type					INTEGER NOT NULL,
		Status					INTEGER NOT NULL,
		AuthorID				TEXT NOT NULL,
		CounterpartyID			TEXT NOT NULL,
		Duration 				INTEGER NOT NULL,
		Price					TEXT NOT NULL,
		Netflags				INTEGER NOT NULL,
		IdentityLevel			INTEGER NOT NULL,
		Blacklist				TEXT NOT NULL,
		Tag						TEXT NOT NULL,
		FrozenSum				TEXT NOT NULL,
		CreatorIdentityLevel	INTEGER NOT NULL,
		CreatorName				TEXT NOT NULL,
		CreatorCountry			TEXT NOT NULL,
		CreatorCertificates		BLOB NOT NULL,
		CPUSysbenchMulti		INTEGER NOT NULL,
		CPUSysbenchOne			INTEGER NOT NULL,
		CPUCores				INTEGER NOT NULL,
		RAMSize					INTEGER NOT NULL,
		StorageSize				INTEGER NOT NULL,
		NetTrafficIn			INTEGER NOT NULL,
		NetTrafficOut			INTEGER NOT NULL,
		GPUCount				INTEGER NOT NULL,
		GPUMem					INTEGER NOT NULL,
		GPUEthHashrate			INTEGER NOT NULL,
		GPUCashHashrate			INTEGER NOT NULL,
		GPURedshift				INTEGER NOT NULL
		)`,
		"createTableWorkers": `
	CREATE TABLE IF NOT EXISTS Workers (
		MasterID					TEXT NOT NULL,
		WorkerID					TEXT NOT NULL,
		Confirmed					INTEGER NOT NULL,
		UNIQUE						(MasterID, WorkerID)
	)`,
		"createTableBlacklists": `
	CREATE TABLE IF NOT EXISTS Blacklists (
		AdderID						TEXT NOT NULL,
		AddeeID						TEXT NOT NULL,
		UNIQUE						(AdderID, AddeeID)
	)`,
		"createTableValidators": `
	CREATE TABLE IF NOT EXISTS Validators (
		Id							TEXT UNIQUE NOT NULL,
		Level						INTEGER NOT NULL
	)`,
		"createTableCertificates": `
	CREATE TABLE IF NOT EXISTS Certificates (
		OwnerID						TEXT NOT NULL,
		Attribute					INTEGER NOT NULL,
		AttributeLevel				INTEGER NOT NULL,
		Value						BLOB NOT NULL,
		ValidatorID					TEXT NOT NULL,
		UNIQUE						(OwnerID, ValidatorID, Attribute),
		FOREIGN KEY (ValidatorID)	REFERENCES Validators(Id) ON DELETE CASCADE
	)`,
		"createTableProfiles": `
	CREATE TABLE IF NOT EXISTS Profiles (
		Id							INTEGER PRIMARY KEY AUTOINCREMENT,
		UserID						TEXT UNIQUE NOT NULL,
		IdentityLevel				INTEGER NOT NULL,
		Name						TEXT NOT NULL,
		Country						TEXT NOT NULL,
		IsCorporation				INTEGER NOT NULL,
		IsProfessional				INTEGER NOT NULL,
		Certificates				BLOB NOT NULL,
		ActiveAsks					INTEGER NOT NULL,
		ActiveBids					INTEGER NOT NULL
	)`,
		"createTableMisc": `
	CREATE TABLE IF NOT EXISTS Misc (
		LastKnownBlock				INTEGER NOT NULL
	)`,
	}
	sqliteCreateIndex = "CREATE INDEX IF NOT EXISTS %s_%s ON %s (%s)"
	sqliteCommands    = map[string]string{
		"insertDeal":                   `INSERT OR IGNORE INTO Deals VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"updateDeal":                   `UPDATE Deals SET Duration=?, Price=?, StartTime=?, EndTime=?, Status=?, BlockedBalance=?, TotalPayout=?, LastBillTS=? WHERE Id=?`,
		"updateDealsSupplier":          `UPDATE Deals SET SupplierCertificates=? WHERE SupplierID=?`,
		"updateDealsConsumer":          `UPDATE Deals SET ConsumerCertificates=? WHERE ConsumerID=?`,
		"selectDealByID":               `SELECT * FROM Deals WHERE id=?`,
		"deleteDeal":                   `DELETE FROM Deals WHERE Id=?`,
		"insertOrder":                  `INSERT OR IGNORE INTO Orders VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"selectOrderByID":              `SELECT * FROM Orders WHERE id=?`,
		"updateOrders":                 `UPDATE Orders SET CreatorIdentityLevel=?, CreatorName=?, CreatorCountry=?, CreatorCertificates=? WHERE AuthorID=?`,
		"deleteOrder":                  `DELETE FROM Orders WHERE Id=?`,
		"insertDealChangeRequest":      `INSERT OR IGNORE INTO DealChangeRequests VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"selectDealChangeRequests":     `SELECT * FROM DealChangeRequests WHERE DealID=? AND RequestType=? AND Status=?`,
		"selectDealChangeRequestsByID": `SELECT * FROM DealChangeRequests WHERE DealID=?`,
		"deleteDealChangeRequest":      `DELETE FROM DealChangeRequests WHERE Id=?`,
		"updateDealChangeRequest":      `UPDATE DealChangeRequests SET Status=? WHERE Id=?`,
		"insertDealCondition":          `INSERT OR IGNORE INTO DealConditions VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"updateDealConditionPayout":    `UPDATE DealConditions SET TotalPayout=? WHERE Id=?`,
		"updateDealConditionEndTime":   `UPDATE DealConditions SET EndTime=? WHERE Id=?`,
		"insertDealPayment":            `INSERT OR IGNORE INTO DealPayments VALUES (?, ?, ?)`,
		"insertWorker":                 `INSERT OR IGNORE INTO Workers VALUES (?, ?, ?)`,
		"updateWorker":                 `UPDATE Workers SET Confirmed=? WHERE MasterID=? AND WorkerID=?`,
		"deleteWorker":                 `DELETE FROM Workers WHERE MasterID=? AND WorkerID=?`,
		"insertBlacklistEntry":         `INSERT OR IGNORE INTO Blacklists VALUES (?, ?)`,
		"selectBlacklists":             `SELECT * FROM Blacklists WHERE AdderID=?`,
		"deleteBlacklistEntry":         `DELETE FROM Blacklists WHERE AdderID=? AND AddeeID=?`,
		"insertValidator":              `INSERT OR IGNORE INTO Validators VALUES (?, ?)`,
		"updateValidator":              `UPDATE Validators SET Level=? WHERE Id=?`,
		"insertCertificate":            `INSERT OR IGNORE INTO Certificates VALUES (?, ?, ?, ?, ?)`,
		"selectCertificates":           `SELECT * FROM Certificates WHERE OwnerID=?`,
		"insertProfileUserID":          `INSERT INTO Profiles VALUES (?, ?, 0, "", "", 0, 0, "", ?, ?)`,
		"selectProfileByID":            `SELECT * FROM Profiles WHERE UserID=?`,
		"profileNotInBlacklist":        `AND UserID NOT IN (SELECT AddeeID FROM Blacklists WHERE AdderID=? AND AddeeID = p.UserID)`,
		"profileInBlacklist":           `AND UserID IN (SELECT AddeeID FROM Blacklists WHERE AdderID=? AND AddeeID = p.UserID)`,
		"updateProfile":                `UPDATE Profiles SET %s=? WHERE UserID=?`,
		"selectLastKnownBlock":         `SELECT * FROM Misc`,
		"updateLastKnownBlock":         `UPDATE Misc Set LastKnownBlock=?`,
	}
)

func setupSQLite(w *DWH) error {
	db, err := sql.Open(w.cfg.Storage.Backend, w.cfg.Storage.Endpoint)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	_, err = db.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		return errors.Wrapf(err, "failed to enable foreign key support (%s)", w.cfg.Storage.Backend)
	}

	for _, cmdName := range orderedSetupCommands {
		_, err = db.Exec(sqliteSetupCommands[cmdName])
		if err != nil {
			return errors.Wrapf(err, "failed to %s (%s)", cmdName, w.cfg.Storage.Backend)
		}
	}

	for column := range DealsColumns {
		if err = createIndex(db, sqliteCreateIndex, "Deals", column); err != nil {
			return err
		}
	}

	for _, column := range []string{"Id", "DealID", "RequestType", "Status"} {
		if err = createIndex(db, sqliteCreateIndex, "DealChangeRequests", column); err != nil {
			return err
		}
	}

	for column := range DealConditionsColumns {
		if err = createIndex(db, sqliteCreateIndex, "DealConditions", column); err != nil {
			return err
		}
	}

	for column := range OrdersColumns {
		if err = createIndex(db, sqliteCreateIndex, "Orders", column); err != nil {
			return err
		}
	}

	for _, column := range []string{"MasterID", "WorkerID"} {
		if err = createIndex(db, sqliteCreateIndex, "Workers", column); err != nil {
			return err
		}
	}

	for _, column := range []string{"AdderID", "AddeeID"} {
		if err = createIndex(db, sqliteCreateIndex, "Blacklists", column); err != nil {
			return err
		}
	}

	if err = createIndex(db, sqliteCreateIndex, "Validators", "Id"); err != nil {
		return err
	}

	if err = createIndex(db, sqliteCreateIndex, "Certificates", "OwnerID"); err != nil {
		return err
	}

	for column := range ProfilesColumns {
		if err = createIndex(db, sqliteCreateIndex, "Profiles", column); err != nil {
			return err
		}
	}

	_, err = db.Exec("INSERT INTO Misc VALUES (0)")
	if err != nil {
		return err
	}

	w.db = db
	w.commands = sqliteCommands

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
}

func runQuery(db *sql.DB, opts *queryOpts) (*sql.Rows, string, error) {
	var (
		query      = fmt.Sprintf("SELECT * FROM %s %s", opts.table, opts.selectAs)
		conditions []string
		values     []interface{}
	)
	for idx, filter := range opts.filters {
		var condition string
		if filter.OpenBracket {
			condition += "("
		}
		condition += fmt.Sprintf("%s%s?", filter.Field, filter.CmpOperator)
		if filter.CloseBracket {
			condition += ")"
		}
		if idx != len(opts.filters)-1 {
			condition += fmt.Sprintf(" %s", filter.BoolOperator)
		}
		conditions = append(conditions, condition)
		values = append(values, filter.Value)
	}
	if len(conditions) > 0 {
		if opts.customFilter != nil {
			conditions = append(conditions, opts.customFilter.clause)
			values = append(values, opts.customFilter.values...)
		}

		query += " WHERE " + strings.Join(conditions, " ")
	}

	if opts.limit > MaxLimit || opts.limit == 0 {
		opts.limit = MaxLimit
	}

	if len(opts.sortings) > 0 {
		query += fmt.Sprintf(" ORDER BY ")
		var sortsFlat []string
		for _, sort := range opts.sortings {
			sortsFlat = append(sortsFlat, fmt.Sprintf("%s %s", sort.Field, pb.SortingOrder_name[int32(sort.Order)]))
		}
		query += strings.Join(sortsFlat, ", ")
	}

	query += fmt.Sprintf(" LIMIT %d", opts.limit)
	if opts.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", opts.offset)
	}
	query += ";"

	rows, err := db.Query(query, values...)
	if err != nil {
		return nil, query, errors.Wrapf(err, "query `%s` failed", query)
	}

	return rows, query, nil
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
