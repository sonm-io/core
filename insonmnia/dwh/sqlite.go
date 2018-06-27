package dwh

import (
	"github.com/Masterminds/squirrel"
)

func newSQLiteStorage(numBenchmarks uint64) *sqlStorage {
	tInfo := newTablesInfo(numBenchmarks)
	storage := &sqlStorage{
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
		MasterID				TEXT NOT NULL,
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
		Level						INTEGER NOT NULL,
		Name						TEXT NOT NULL DEFAULT '',
		Logo						TEXT NOT NULL DEFAULT '',
		URL							TEXT NOT NULL DEFAULT '',
		Description					TEXT NOT NULL DEFAULT '',
		Price						TEXT NOT NULL DEFAULT '0'
	)`,
			createTableCertificates: `
	CREATE TABLE IF NOT EXISTS Certificates (
	    Id						    TEXT NOT NULL, 
		OwnerID						TEXT NOT NULL,
		Attribute					INTEGER NOT NULL,
		AttributeLevel				INTEGER NOT NULL,
		Value						BLOB NOT NULL,
		ValidatorID					TEXT NOT NULL,
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
		tablesInfo:    tInfo,
		builder: func() squirrel.StatementBuilderType {
			return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
		},
	}

	return storage
}
