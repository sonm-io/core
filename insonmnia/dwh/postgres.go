package dwh

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
)

var (
	postgresSetupCommands = map[string]string{
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
		SupplierCertificates    BYTEA NOT NULL,
		ConsumerCertificates    BYTEA NOT NULL,
		ActiveChangeRequest     BOOLEAN NOT NULL,
		CPUSysbenchMulti		BIGINT NOT NULL,
		CPUSysbenchOne			BIGINT NOT NULL,
		CPUCores				BIGINT NOT NULL,
		RAMSize					BIGINT NOT NULL,
		StorageSize				BIGINT NOT NULL,
		NetTrafficIn			BIGINT NOT NULL,
		NetTrafficOut			BIGINT NOT NULL,
		GPUCount				BIGINT NOT NULL,
		GPUMem					BIGINT NOT NULL,
		GPUEthHashrate			BIGINT NOT NULL,
		GPUCashHashrate			BIGINT NOT NULL,
		GPURedshift				BIGINT NOT NULL
	)`,
		"createTableDealConditions": `
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
		"createTableDealPayments": `
	CREATE TABLE IF NOT EXISTS DealPayments (
		BillTS						INTEGER NOT NULL,
		PaidAmount					TEXT NOT NULL,
		DealID						TEXT NOT NULL REFERENCES Deals(Id) ON DELETE CASCADE,
		UNIQUE						(BillTS, PaidAmount, DealID)
	)`,
		"createTableChangeRequests": `
	CREATE TABLE IF NOT EXISTS DealChangeRequests (
		Id 							TEXT UNIQUE NOT NULL,
		CreatedTS					INTEGER NOT NULL,
		RequestType					TEXT NOT NULL,
		Duration 					INTEGER NOT NULL,
		Price						TEXT NOT NULL,
		Status						INTEGER NOT NULL,
		DealID						TEXT NOT NULL REFERENCES Deals(Id) ON DELETE CASCADE
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
		CreatorCertificates		BYTEA NOT NULL,
		CPUSysbenchMulti		BIGINT NOT NULL,
		CPUSysbenchOne			BIGINT NOT NULL,
		CPUCores				BIGINT NOT NULL,
		RAMSize					BIGINT NOT NULL,
		StorageSize				BIGINT NOT NULL,
		NetTrafficIn			BIGINT NOT NULL,
		NetTrafficOut			BIGINT NOT NULL,
		GPUCount				BIGINT NOT NULL,
		GPUMem					BIGINT NOT NULL,
		GPUEthHashrate			BIGINT NOT NULL,
		GPUCashHashrate			BIGINT NOT NULL,
		GPURedshift				BIGINT NOT NULL
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
		Value						BYTEA NOT NULL,
		ValidatorID					TEXT NOT NULL REFERENCES Validators(Id) ON DELETE CASCADE,
		UNIQUE						(OwnerID, ValidatorID, Attribute, Value)
	)`,
		"createTableProfiles": `
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
		"createTableMisc": `
	CREATE TABLE IF NOT EXISTS Misc (
		Id							BIGSERIAL PRIMARY KEY,
		LastKnownBlock				INTEGER NOT NULL
	)`,
	}
	postgresCreateIndex = "CREATE INDEX IF NOT EXISTS %s_%s ON %s (%s)"
	postgresCommands    = map[string]string{
		"insertDeal":                   `INSERT INTO Deals VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32)`,
		"updateDeal":                   `UPDATE Deals SET Duration = $1, Price = $2, StartTime = $3, EndTime = $4, Status = $5, BlockedBalance = $6, TotalPayout = $7, LastBillTS = $7 WHERE Id = $8`,
		"updateDealsSupplier":          `UPDATE Deals SET SupplierCertificates = $1 WHERE SupplierID = $2`,
		"updateDealsConsumer":          `UPDATE Deals SET ConsumerCertificates = $1 WHERE ConsumerID = $2`,
		"selectDealByID":               `SELECT * FROM Deals WHERE id = $1`,
		"deleteDeal":                   `DELETE FROM Deals WHERE Id = $1`,
		"insertOrder":                  `INSERT INTO Orders VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30)`,
		"selectOrderByID":              `SELECT * FROM Orders WHERE id = $1`,
		"updateOrders":                 `UPDATE Orders SET CreatorIdentityLevel = $1, CreatorName = $2, CreatorCountry = $3, CreatorCertificates = $4 WHERE AuthorID = $5`,
		"deleteOrder":                  `DELETE FROM Orders WHERE Id = $1`,
		"insertDealChangeRequest":      `INSERT INTO DealChangeRequests VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		"selectDealChangeRequests":     `SELECT * FROM DealChangeRequests WHERE DealID = $1 AND RequestType = $2 AND Status = $3`,
		"selectDealChangeRequestsByID": `SELECT * FROM DealChangeRequests WHERE DealID = $1`,
		"deleteDealChangeRequest":      `DELETE FROM DealChangeRequests WHERE Id = $1`,
		"updateDealChangeRequest":      `UPDATE DealChangeRequests SET Status = $1 WHERE Id = $2`,
		"insertDealCondition":          `INSERT INTO DealConditions(SupplierID, ConsumerID, MasterID, Duration, Price, StartTime, EndTime, TotalPayout, DealID) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		"updateDealConditionPayout":    `UPDATE DealConditions SET TotalPayout = $1 WHERE Id = $2`,
		"updateDealConditionEndTime":   `UPDATE DealConditions SET EndTime = $1 WHERE Id = $2`,
		"insertDealPayment":            `INSERT INTO DealPayments VALUES ($1, $2, $3)`,
		"insertWorker":                 `INSERT INTO Workers VALUES ($1, $2, $3)`,
		"updateWorker":                 `UPDATE Workers SET Confirmed = $1 WHERE MasterID = $2 AND WorkerID = $3`,
		"deleteWorker":                 `DELETE FROM Workers WHERE MasterID = $1 AND WorkerID = $2`,
		"insertBlacklistEntry":         `INSERT INTO Blacklists VALUES ($1, $2)`,
		"selectBlacklists":             `SELECT * FROM Blacklists WHERE AdderID = $1`,
		"deleteBlacklistEntry":         `DELETE FROM Blacklists WHERE AdderID = $1 AND AddeeID = $2`,
		"insertValidator":              `INSERT INTO Validators VALUES ($1, $2)`,
		"updateValidator":              `UPDATE Validators SET Level = $1 WHERE Id = $2`,
		"insertCertificate":            `INSERT INTO Certificates VALUES ($1, $2, $3, $4, $5)`,
		"selectCertificates":           `SELECT * FROM Certificates WHERE OwnerID = $1`,
		"insertProfileUserID":          `INSERT INTO Profiles (UserID, IdentityLevel, Name, Country, IsCorporation, IsProfessional, Certificates, ActiveAsks, ActiveBids ) VALUES ($1, 0, '', '', FALSE, FALSE, E'\\000', $2, $3)`,
		"selectProfileByID":            `SELECT * FROM Profiles WHERE UserID = $1`,
		"profileNotInBlacklist":        `AND UserID NOT IN (SELECT AddeeID FROM Blacklists WHERE AdderID = $ AND AddeeID = p.UserID)`,
		"profileInBlacklist":           `AND UserID IN (SELECT AddeeID FROM Blacklists WHERE AdderID = $ AND AddeeID = p.UserID)`,
		"updateProfile":                `UPDATE Profiles SET %s = $1 WHERE UserID = $2`,
		"selectLastKnownBlock":         `SELECT LastKnownBlock FROM Misc WHERE Id = 1`,
		"insertLastKnownBlock":         `INSERT INTO Misc(LastKnownBlock) VALUES ($1)`,
		"updateLastKnownBlock":         `UPDATE Misc SET LastKnownBlock = $1 WHERE Id = 1`,
	}
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

	for _, cmdName := range orderedSetupCommands {
		_, err = db.Exec(postgresSetupCommands[cmdName])
		if err != nil {
			return errors.Wrapf(err, "failed to %s (%s)", cmdName, w.cfg.Storage.Backend)
		}
	}

	for column := range DealsColumns {
		if err = createIndex(db, postgresCreateIndex, "Deals", column); err != nil {
			return err
		}
	}

	for _, column := range []string{"Id", "DealID", "RequestType", "Status"} {
		if err = createIndex(db, postgresCreateIndex, "DealChangeRequests", column); err != nil {
			return err
		}
	}

	for column := range DealConditionsColumns {
		if err = createIndex(db, postgresCreateIndex, "DealConditions", column); err != nil {
			return err
		}
	}

	for column := range OrdersColumns {
		if err = createIndex(db, postgresCreateIndex, "Orders", column); err != nil {
			return err
		}
	}

	for _, column := range []string{"MasterID", "WorkerID"} {
		if err = createIndex(db, postgresCreateIndex, "Workers", column); err != nil {
			return err
		}
	}

	for _, column := range []string{"AdderID", "AddeeID"} {
		if err = createIndex(db, postgresCreateIndex, "Blacklists", column); err != nil {
			return err
		}
	}

	if err = createIndex(db, postgresCreateIndex, "Validators", "Id"); err != nil {
		return err
	}

	if err = createIndex(db, postgresCreateIndex, "Certificates", "OwnerID"); err != nil {
		return err
	}

	for column := range ProfilesColumns {
		if err = createIndex(db, postgresCreateIndex, "Profiles", column); err != nil {
			return err
		}
	}

	w.db = db
	w.commands = postgresCommands
	w.runQuery = runQueryPostgres

	return nil
}

func runQueryPostgres(db *sql.DB, opts *queryOpts) (*sql.Rows, string, error) {
	var (
		query      = fmt.Sprintf("SELECT * FROM %s %s", opts.table, opts.selectAs)
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
