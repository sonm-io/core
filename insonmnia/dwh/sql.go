package dwh

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
)

const (
	MaxLimit = 50
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
)

var (
	OrdersColumns = []string{
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
		"CPUSysbenchMulti",
		"CPUSysbenchOne",
		"CPUCores",
		"RAMSize",
		"StorageSize",
		"NetTrafficIn",
		"NetTrafficOut",
		"GPUCount",
		"GPUMem",
		"GPUEthHashrate",
		"GPUCashHashrate",
		"GPURedshift",
	}
	DealsColumns = []string{
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
		"CPUSysbenchMulti",
		"CPUSysbenchOne",
		"CPUCores",
		"RAMSize",
		"StorageSize",
		"NetTrafficIn",
		"NetTrafficOut",
		"GPUCount",
		"GPUMem",
		"GPUEthHashrate",
		"GPUCashHashrate",
		"GPURedshift",
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
		GPURedshift				INTEGER NOT NULL,
		FOREIGN KEY(Id)			REFERENCES Orders(Id) ON DELETE CASCADE
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
		UserID						TEXT UNIQUE NOT NULL,
		IdentityLevel				INTEGER NOT NULL,
		Name						TEXT NOT NULL,
		Country						TEXT NOT NULL,
		IsCorporation				INTEGER NOT NULL,
		IsProfessional				INTEGER NOT NULL,
		Certificates				BLOB NOT NULL
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
		"insertDealCondition":          `INSERT OR IGNORE INTO DealConditions VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"selectDealCondition":          `SELECT rowid, * FROM DealConditions WHERE DealID=? ORDER BY rowid DESC LIMIT 1`,
		"updateDealConditionPayout":    `UPDATE DealConditions SET TotalPayout=? WHERE rowid=?`,
		"updateDealConditionEndTime":   `UPDATE DealConditions SET EndTime=? WHERE rowid=?`,
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
		"insertProfileUserID":          `INSERT INTO Profiles VALUES (?, 0, "", "", 0, 0, "")`,
		"selectProfileByID":            `SELECT * FROM Profiles WHERE UserID=?`,
		"updateProfile":                `UPDATE Profiles SET %s=? WHERE UserID=?`,
		"selectLastKnownBlock":         `SELECT * FROM Misc`,
		"updateLastKnownBlock":         `INSERT OR REPLACE INTO Misc (rowid, LastKnownBlock) VALUES (1, ?)`,
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

	for cmdName, cmd := range sqliteSetupCommands {
		_, err = db.Exec(cmd)
		if err != nil {
			return errors.Wrapf(err, "failed to %s (%s)", cmdName, w.cfg.Storage.Backend)
		}
	}

	for _, column := range DealsColumns {
		cmd := fmt.Sprintf(sqliteCreateIndex, "Deals", column, "Deals", column)
		_, err = db.Exec(cmd)
		if err != nil {
			return errors.Wrapf(err, "failed to %s (%s)", cmd, w.cfg.Storage.Backend)
		}
	}

	for _, column := range OrdersColumns {
		cmd := fmt.Sprintf(sqliteCreateIndex, "Orders", column, "Orders", column)
		_, err = db.Exec(cmd)
		if err != nil {
			return errors.Wrapf(err, "failed to %s (%s)", cmd, w.cfg.Storage.Backend)
		}
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

func runQuery(db *sql.DB, table string, offset, limit uint64, orderColumn, orderType string, filters ...*filter) (
	*sql.Rows, string, error) {
	var (
		query      = fmt.Sprintf("SELECT * FROM %s", table)
		conditions []string
		values     []interface{}
	)
	for idx, filter := range filters {
		var condition string
		if filter.OpenBracket {
			condition += "("
		}
		condition += fmt.Sprintf("%s%s?", filter.Field, filter.CmpOperator)
		if filter.CloseBracket {
			condition += ")"
		}
		if idx != len(filters)-1 {
			condition += fmt.Sprintf(" %s", filter.BoolOperator)
		}
		conditions = append(conditions, condition)
		values = append(values, filter.Value)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " ")
	}

	if limit > MaxLimit || limit == 0 {
		limit = MaxLimit
	}
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderType)
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}
	query += ";"

	rows, err := db.Query(query, values...)
	if err != nil {
		return nil, query, errors.Wrapf(err, "query `%s` failed", query)
	}

	return rows, query, nil
}
