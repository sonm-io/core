package dwh

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
)

func (w *DWH) setupSQLite(db *sql.DB, numBenchmarks uint64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	_, err := db.Exec(`PRAGMA foreign_keys=ON`)
	if err != nil {
		db.Close()
		return errors.Wrapf(err, "failed to enable foreign key support (%s)", w.cfg.Storage.Backend)
	}

	store := newSQLiteStorage(newTablesInfo(numBenchmarks), numBenchmarks)
	if err := store.Setup(db); err != nil {
		return errors.Wrap(err, "failed to setup store")
	}

	w.storage = store

	return nil
}

func newSQLiteStorage(tInfo *tablesInfo, numBenchmarks uint64) *sqlStorage {
	formatCb := func(_ uint64, lastArg bool) string {
		if lastArg {
			return "?"
		}
		return "?, "
	}

	store := &sqlStorage{
		commands: &sqlCommands{
			insertDeal:                   makeInsertDealQuery(`INSERT INTO Deals(%s) VALUES (%s)`, formatCb, numBenchmarks, tInfo),
			updateDeal:                   `UPDATE Deals SET Duration=?, Price=?, StartTime=?, EndTime=?, Status=?, BlockedBalance=?, TotalPayout=?, LastBillTS=? WHERE Id=?`,
			updateDealsSupplier:          `UPDATE Deals SET SupplierCertificates=? WHERE SupplierID=?`,
			updateDealsConsumer:          `UPDATE Deals SET ConsumerCertificates=? WHERE ConsumerID=?`,
			updateDealPayout:             `UPDATE Deals SET TotalPayout = ?, LastBillTS = ? WHERE Id = ?`,
			selectDealByID:               makeSelectDealByIDQuery(`SELECT %s FROM Deals WHERE id=?`, tInfo),
			deleteDeal:                   `DELETE FROM Deals WHERE Id=?`,
			insertOrder:                  makeInsertOrderQuery(`INSERT INTO Orders(%s) VALUES (%s)`, formatCb, numBenchmarks, tInfo),
			selectOrderByID:              makeSelectOrderByIDQuery(`SELECT %s FROM Orders WHERE id=?`, tInfo),
			updateOrderStatus:            `UPDATE Orders SET Status=? WHERE Id=?`,
			updateOrders:                 `UPDATE Orders SET CreatorIdentityLevel=?, CreatorName=?, CreatorCountry=?, CreatorCertificates=? WHERE AuthorID=?`,
			deleteOrder:                  `DELETE FROM Orders WHERE Id=?`,
			insertDealChangeRequest:      `INSERT INTO DealChangeRequests VALUES (?, ?, ?, ?, ?, ?, ?)`,
			selectDealChangeRequests:     `SELECT * FROM DealChangeRequests WHERE DealID=? AND RequestType=? AND Status=?`,
			selectDealChangeRequestsByID: `SELECT * FROM DealChangeRequests WHERE DealID=?`,
			deleteDealChangeRequest:      `DELETE FROM DealChangeRequests WHERE Id=?`,
			updateDealChangeRequest:      `UPDATE DealChangeRequests SET Status=? WHERE Id=?`,
			insertDealCondition:          `INSERT INTO DealConditions VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			updateDealConditionPayout:    `UPDATE DealConditions SET TotalPayout=? WHERE Id=?`,
			updateDealConditionEndTime:   `UPDATE DealConditions SET EndTime=? WHERE Id=?`,
			insertDealPayment:            `INSERT INTO DealPayments VALUES (?, ?, ?)`,
			insertWorker:                 `INSERT INTO Workers VALUES (?, ?, ?)`,
			updateWorker:                 `UPDATE Workers SET Confirmed=? WHERE MasterID=? AND WorkerID=?`,
			deleteWorker:                 `DELETE FROM Workers WHERE MasterID=? AND WorkerID=?`,
			insertBlacklistEntry:         `INSERT INTO Blacklists VALUES (?, ?)`,
			selectBlacklists:             `SELECT * FROM Blacklists WHERE AdderID=?`,
			deleteBlacklistEntry:         `DELETE FROM Blacklists WHERE AdderID=? AND AddeeID=?`,
			insertValidator:              `INSERT INTO Validators VALUES (?, ?)`,
			updateValidator:              `UPDATE Validators SET Level=? WHERE Id=?`,
			insertCertificate:            `INSERT INTO Certificates VALUES (?, ?, ?, ?, ?)`,
			selectCertificates:           `SELECT * FROM Certificates WHERE OwnerID=?`,
			insertProfileUserID:          `INSERT INTO Profiles VALUES (NULL, ?, 0, "", "", 0, 0, ?, ?, ?)`,
			selectProfileByID:            `SELECT * FROM Profiles WHERE UserID=?`,
			profileNotInBlacklist:        `AND UserID NOT IN (SELECT AddeeID FROM Blacklists WHERE AdderID=? AND AddeeID = p.UserID)`,
			profileInBlacklist:           `AND UserID IN (SELECT AddeeID FROM Blacklists WHERE AdderID=? AND AddeeID = p.UserID)`,
			updateProfile:                `UPDATE Profiles SET %s=? WHERE UserID=?`,
			selectLastKnownBlock:         `SELECT LastKnownBlock FROM Misc WHERE Id=1`,
			insertLastKnownBlock:         `INSERT INTO Misc VALUES (NULL, ?)`,
			updateLastKnownBlock:         `UPDATE Misc Set LastKnownBlock=? WHERE Id=1`,
			storeStaleID:                 `INSERT INTO StaleIDs VALUES (?)`,
			removeStaleID:                `DELETE FROM StaleIDs WHERE Id = ?`,
			checkStaleID:                 `SELECT * FROM StaleIDs WHERE Id = ?`,
		},
		setupCommands: &sqlSetupCommands{
			// Incomplete, modified during setup.
			createTableDeals: makeTableWithBenchmarks(`
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
		ActiveChangeRequest     INTEGER NOT NULL`, `INTEGER DEFAULT 0`),
			createTableDealConditions: `
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
			createTableDealPayments: `
	CREATE TABLE IF NOT EXISTS DealPayments (
		BillTS						INTEGER NOT NULL,
		PaidAmount					TEXT NOT NULL,
		DealID						TEXT NOT NULL,
		UNIQUE						(BillTS, PaidAmount, DealID),
		FOREIGN KEY (DealID) 		REFERENCES Deals(Id) ON DELETE CASCADE
	)`,
			createTableChangeRequests: `
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
			// Incomplete, modified during setup.
			createTableOrders: makeTableWithBenchmarks(`
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
		Tag						BLOB NOT NULL,
		FrozenSum				TEXT NOT NULL,
		CreatorIdentityLevel	INTEGER NOT NULL,
		CreatorName				TEXT NOT NULL,
		CreatorCountry			TEXT NOT NULL,
		CreatorCertificates		BLOB NOT NULL`, `INTEGER DEFAULT 0`),
			createTableWorkers: `
	CREATE TABLE IF NOT EXISTS Workers (
		MasterID					TEXT NOT NULL,
		WorkerID					TEXT NOT NULL,
		Confirmed					INTEGER NOT NULL,
		UNIQUE						(MasterID, WorkerID)
	)`,
			createTableBlacklists: `
	CREATE TABLE IF NOT EXISTS Blacklists (
		AdderID						TEXT NOT NULL,
		AddeeID						TEXT NOT NULL,
		UNIQUE						(AdderID, AddeeID)
	)`,
			createTableValidators: `
	CREATE TABLE IF NOT EXISTS Validators (
		Id							TEXT UNIQUE NOT NULL,
		Level						INTEGER NOT NULL
	)`,
			createTableCertificates: `
	CREATE TABLE IF NOT EXISTS Certificates (
		OwnerID						TEXT NOT NULL,
		Attribute					INTEGER NOT NULL,
		AttributeLevel				INTEGER NOT NULL,
		Value						BLOB NOT NULL,
		ValidatorID					TEXT NOT NULL,
		UNIQUE						(OwnerID, ValidatorID, Attribute, Value),
		FOREIGN KEY (ValidatorID)	REFERENCES Validators(Id) ON DELETE CASCADE
	)`,
			createTableProfiles: `
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
			createTableStaleIDs: `
	CREATE TABLE IF NOT EXISTS StaleIDs (
		Id 							TEXT NOT NULL
	)`,
			createTableMisc: `
	CREATE TABLE IF NOT EXISTS Misc (
		Id							INTEGER PRIMARY KEY AUTOINCREMENT,
		LastKnownBlock				INTEGER NOT NULL
	)`,
			createIndexCmd: `CREATE INDEX IF NOT EXISTS %s_%s ON %s (%s)`,
			tablesInfo:     tInfo,
		},
		numBenchmarks: numBenchmarks,
		queryRunner:   newSQLiteQueryRunner(tInfo),
		tablesInfo:    tInfo,
		formatCb:      formatCb,
	}

	return store
}

type sqliteQueryRunner struct {
	tablesInfo *tablesInfo
}

func newSQLiteQueryRunner(tInfo *tablesInfo) queryRunner {
	return &sqliteQueryRunner{tablesInfo: tInfo}
}

func (r *sqliteQueryRunner) Run(conn queryConn, opts *queryOpts) (*sql.Rows, uint64, error) {
	var columns string
	switch opts.table {
	case "Deals":
		columns = strings.Join(r.tablesInfo.DealColumns, ", ")
	case "Orders":
		columns = strings.Join(r.tablesInfo.OrderColumns, ", ")
	default:
		columns = "*"
	}
	var (
		query      = fmt.Sprintf("SELECT %s FROM %s %s", columns, opts.table, opts.selectAs)
		countQuery = fmt.Sprintf("SELECT count(*) FROM %s %s", opts.table, opts.selectAs)
		conditions []string
		values     []interface{}
	)
	for idx, filter := range opts.filters {
		var condition string
		if filter.OpenBracket {
			condition += "("
		}
		condition += fmt.Sprintf("%s %s ?", filter.Field, filter.CmpOperator)
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
		countQuery += " WHERE " + strings.Join(conditions, " ")
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
	countQuery += ";"

	var count uint64
	if opts.withCount {
		countRows, err := conn.Query(countQuery, values...)
		if err != nil {
			return nil, 0, errors.Wrapf(err, "count query `%s` failed", countQuery)
		}
		for countRows.Next() {
			countRows.Scan(&count)
		}
	}

	rows, err := conn.Query(query, values...)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "query `%s` failed", query)
	}

	return rows, count, nil
}
