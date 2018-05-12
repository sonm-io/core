package dwh

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
)

func setupPostgres(w *DWH) error {
	db, err := sql.Open(w.cfg.Storage.Backend, w.cfg.Storage.Endpoint)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	var (
		tInfo         = newTablesInfo(w.numBenchmarks)
		commands      = newPostgresCommands(tInfo, w.numBenchmarks)
		setupCommands = newPostgresSetupCommands()
	)
	if err = setupCommands.SetupTables(db); err != nil {
		return errors.Wrap(err, "failed to setup tables")
	}

	w.db = db
	w.commands = commands
	w.tablesInfo = tInfo
	w.queryRunner = newPostgresQueryRunner(db, tInfo)

	if w.cfg.ColdStart != nil {
		go coldStart(w, setupCommands.CreateIndices)
	} else {
		if err := setupCommands.CreateIndices(db, tInfo); err != nil {
			return errors.Wrap(err, "failed to CreateIndices")
		}
	}

	return nil
}

func newPostgresCommands(tInfo *tablesInfo, numBenchmarks uint64) *SQLCommands {
	commands := &SQLCommands{
		insertDeal:                   `INSERT INTO Deals(%s) VALUES (%s)`,
		updateDeal:                   `UPDATE Deals SET Duration = $1, Price = $2, StartTime = $3, EndTime = $4, Status = $5, BlockedBalance = $6, TotalPayout = $7, LastBillTS = $7 WHERE Id = $8`,
		updateDealsSupplier:          `UPDATE Deals SET SupplierCertificates = $1 WHERE SupplierID = $2`,
		updateDealsConsumer:          `UPDATE Deals SET ConsumerCertificates = $1 WHERE ConsumerID = $2`,
		updateDealPayout:             `UPDATE Deals SET TotalPayout = $1 WHERE Id = $2`,
		selectDealByID:               `SELECT %s FROM Deals WHERE id = $1`,
		deleteDeal:                   `DELETE FROM Deals WHERE Id = $1`,
		insertOrder:                  `INSERT INTO Orders(%s) VALUES (%s)`,
		selectOrderByID:              `SELECT %s FROM Orders WHERE id = $1`,
		updateOrders:                 `UPDATE Orders SET CreatorIdentityLevel = $1, CreatorName = $2, CreatorCountry = $3, CreatorCertificates = $4 WHERE AuthorID = $5`,
		updateOrderStatus:            `UPDATE Orders SET Status = $1 WHERE Id = $2`,
		deleteOrder:                  `DELETE FROM Orders WHERE Id = $1`,
		insertDealChangeRequest:      `INSERT INTO DealChangeRequests VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		selectDealChangeRequests:     `SELECT * FROM DealChangeRequests WHERE DealID = $1 AND RequestType = $2 AND Status = $3`,
		selectDealChangeRequestsByID: `SELECT * FROM DealChangeRequests WHERE DealID = $1`,
		deleteDealChangeRequest:      `DELETE FROM DealChangeRequests WHERE Id = $1`,
		updateDealChangeRequest:      `UPDATE DealChangeRequests SET Status = $1 WHERE Id = $2`,
		insertDealCondition:          `INSERT INTO DealConditions(SupplierID, ConsumerID, MasterID, Duration, Price, StartTime, EndTime, TotalPayout, DealID) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		updateDealConditionPayout:    `UPDATE DealConditions SET TotalPayout = $1 WHERE Id = $2`,
		updateDealConditionEndTime:   `UPDATE DealConditions SET EndTime = $1 WHERE Id = $2`,
		insertDealPayment:            `INSERT INTO DealPayments VALUES ($1, $2, $3)`,
		insertWorker:                 `INSERT INTO Workers VALUES ($1, $2, $3)`,
		updateWorker:                 `UPDATE Workers SET Confirmed = $1 WHERE MasterID = $2 AND WorkerID = $3`,
		deleteWorker:                 `DELETE FROM Workers WHERE MasterID = $1 AND WorkerID = $2`,
		insertBlacklistEntry:         `INSERT INTO Blacklists VALUES ($1, $2)`,
		selectBlacklists:             `SELECT * FROM Blacklists WHERE AdderID = $1`,
		deleteBlacklistEntry:         `DELETE FROM Blacklists WHERE AdderID = $1 AND AddeeID = $2`,
		insertValidator:              `INSERT INTO Validators VALUES ($1, $2)`,
		updateValidator:              `UPDATE Validators SET Level = $1 WHERE Id = $2`,
		insertCertificate:            `INSERT INTO Certificates VALUES ($1, $2, $3, $4, $5)`,
		selectCertificates:           `SELECT * FROM Certificates WHERE OwnerID = $1`,
		insertProfileUserID:          `INSERT INTO Profiles (UserID, IdentityLevel, Name, Country, IsCorporation, IsProfessional, Certificates, ActiveAsks, ActiveBids ) VALUES ($1, 0, '', '', FALSE, FALSE, $2, $3, $4)`,
		selectProfileByID:            `SELECT * FROM Profiles WHERE UserID = $1`,
		profileNotInBlacklist:        `AND UserID NOT IN (SELECT AddeeID FROM Blacklists WHERE AdderID = $ AND AddeeID = p.UserID)`,
		profileInBlacklist:           `AND UserID IN (SELECT AddeeID FROM Blacklists WHERE AdderID = $ AND AddeeID = p.UserID)`,
		updateProfile:                `UPDATE Profiles SET %s = $1 WHERE UserID = $2`,
		selectLastKnownBlock:         `SELECT LastKnownBlock FROM Misc WHERE Id = 1`,
		insertLastKnownBlock:         `INSERT INTO Misc(LastKnownBlock) VALUES ($1)`,
		updateLastKnownBlock:         `UPDATE Misc SET LastKnownBlock = $1 WHERE Id = 1`,
	}

	format := func(argID uint64, lastArg bool) string {
		if lastArg {
			return fmt.Sprintf("$%d", argID)
		}
		return fmt.Sprintf("$%d, ", argID)
	}
	commands.Finalize(numBenchmarks, tInfo, format)

	return commands
}

func newPostgresSetupCommands() *SQLSetupCommands {
	setupCommands := &SQLSetupCommands{
		createTableDeals: `
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
		SupplierCertificates    BYTEA NOT NULL,
		ConsumerCertificates    BYTEA NOT NULL,
		ActiveChangeRequest     BOOLEAN NOT NULL`,
		createTableDealConditions: `
	CREATE TABLE IF NOT EXISTS DealConditions (
		Id							BIGSERIAL PRIMARY KEY,
		SupplierID					TEXT NOT NULL,
		ConsumerID					TEXT NOT NULL,
		MasterID					TEXT NOT NULL,
		Duration 					INTEGER NOT NULL,
		Price						TEXT NOT NULL,
		StartTime					INTEGER NOT NULL,
		EndTime						INTEGER NOT NULL,
		TotalPayout					TEXT NOT NULL,
		DealID						TEXT NOT NULL REFERENCES Deals(Id) ON DELETE CASCADE
	)`,
		createTableDealPayments: `
	CREATE TABLE IF NOT EXISTS DealPayments (
		BillTS						INTEGER NOT NULL,
		PaidAmount					TEXT NOT NULL,
		DealID						TEXT NOT NULL REFERENCES Deals(Id) ON DELETE CASCADE,
		UNIQUE						(BillTS, PaidAmount, DealID)
	)`,
		createTableChangeRequests: `
	CREATE TABLE IF NOT EXISTS DealChangeRequests (
		Id 							TEXT UNIQUE NOT NULL,
		CreatedTS					INTEGER NOT NULL,
		RequestType					TEXT NOT NULL,
		Duration 					INTEGER NOT NULL,
		Price						TEXT NOT NULL,
		Status						INTEGER NOT NULL,
		DealID						TEXT NOT NULL REFERENCES Deals(Id) ON DELETE CASCADE
	)`,
		createTableOrders: `
	CREATE TABLE IF NOT EXISTS Orders (
		Id						TEXT UNIQUE NOT NULL,
		CreatedTS				INTEGER NOT NULL,
		DealID					TEXT NOT NULL,
		Type					INTEGER NOT NULL,
		Status					INTEGER NOT NULL,
		AuthorID				TEXT NOT NULL,
		CounterpartyID			TEXT NOT NULL,
		Duration 				BIGINT NOT NULL,
		Price					TEXT NOT NULL,
		Netflags				INTEGER NOT NULL,
		IdentityLevel			INTEGER NOT NULL,
		Blacklist				TEXT NOT NULL,
		Tag						BYTEA NOT NULL,
		FrozenSum				TEXT NOT NULL,
		CreatorIdentityLevel	INTEGER NOT NULL,
		CreatorName				TEXT NOT NULL,
		CreatorCountry			TEXT NOT NULL,
		CreatorCertificates		BYTEA NOT NULL`,
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
		Value						BYTEA NOT NULL,
		ValidatorID					TEXT NOT NULL REFERENCES Validators(Id) ON DELETE CASCADE,
		UNIQUE						(OwnerID, ValidatorID, Attribute, Value)
	)`,
		createTableProfiles: `
	CREATE TABLE IF NOT EXISTS Profiles (
		Id							BIGSERIAL PRIMARY KEY,
		UserID						TEXT UNIQUE NOT NULL,
		IdentityLevel				INTEGER NOT NULL,
		Name						TEXT NOT NULL,
		Country						TEXT NOT NULL,
		IsCorporation				BOOLEAN NOT NULL,
		IsProfessional				BOOLEAN NOT NULL,
		Certificates				BYTEA NOT NULL,
		ActiveAsks					INTEGER NOT NULL,
		ActiveBids					INTEGER NOT NULL
	)`,
		createTableMisc: `
	CREATE TABLE IF NOT EXISTS Misc (
		Id							BIGSERIAL PRIMARY KEY,
		LastKnownBlock				INTEGER NOT NULL
	)`,
		createIndex: `CREATE INDEX IF NOT EXISTS %s_%s ON %s (%s)`,
	}

	benchmarkColumns := make([]string, NumMaxBenchmarks)
	for benchmarkID := uint64(0); benchmarkID < NumMaxBenchmarks; benchmarkID++ {
		benchmarkColumns[benchmarkID] = fmt.Sprintf("%s BIGINT NOT NULL", getBenchmarkColumn(uint64(benchmarkID)))
	}
	setupCommands.createTableDeals = strings.Join(
		append([]string{setupCommands.createTableDeals}, benchmarkColumns...), ",\n") + ")"
	setupCommands.createTableOrders = strings.Join(
		append([]string{setupCommands.createTableOrders}, benchmarkColumns...), ",\n") + ")"

	return setupCommands
}

type postgresQueryRunner struct {
	db         *sql.DB
	tablesInfo *tablesInfo
}

func newPostgresQueryRunner(db *sql.DB, tInfo *tablesInfo) QueryRunner {
	return &postgresQueryRunner{
		db:         db,
		tablesInfo: tInfo,
	}
}

func (r *postgresQueryRunner) Run(opts *queryOpts) (*sql.Rows, uint64, error) {
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
		numFilters = 1
	)
	for idx, filter := range opts.filters {
		var condition string
		if filter.OpenBracket {
			condition += "("
		}
		condition += fmt.Sprintf("%s %s $%d", filter.Field, filter.CmpOperator, numFilters)
		if filter.CloseBracket {
			condition += ")"
		}
		if idx != len(opts.filters)-1 {
			condition += fmt.Sprintf(" %s", filter.BoolOperator)
		}
		conditions = append(conditions, condition)
		values = append(values, filter.Value)
		numFilters++
	}
	if len(conditions) > 0 {
		if opts.customFilter != nil {
			clause := strings.Replace(opts.customFilter.clause, "$", fmt.Sprintf("$%d", numFilters), 1)
			conditions = append(conditions, clause)
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
		countRows, err := r.db.Query(countQuery, values...)
		if err != nil {
			return nil, 0, errors.Wrapf(err, "count query `%s` failed", countQuery)
		}
		for countRows.Next() {
			countRows.Scan(&count)
		}
	}

	rows, err := r.db.Query(query, values...)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "query `%s` failed", query)
	}

	return rows, count, nil
}
