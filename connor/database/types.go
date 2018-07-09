package database

import "time"

type PoolDb struct {
	DealID                 int64     `db:"DEAL_ID"`
	PoolID                 string    `db:"POOL_ID"`
	WorkerReportedHashrate float64   `db:"W_REP_HASH"`
	WorkerAvgHashrate      float64   `db:"W_AVG_HASH"`
	BadGuy                 int64     `db:"BAD_GUY"`
	Iterations             int64     `db:"ITERATIONS"`
	TimeStart              time.Time `db:"TIME_START"`
	TimeUpdate             time.Time `db:"TIME_UPDATE"`
}

type BlackListDb struct {
	MasterID       string `db:"MASTER_ID"`
	FailSupplierId string `db:"FAIL_SUPPLIER_ID"`
	BanStatus      int64  `db:"BAN_STATUS"`
	DealId         int64  `db:"DEAL_ID"`
}

type TokenDb struct {
	ID              int64     `db:"TOKEN_ID"`
	Name            string    `db:"NAME"`
	UsdPrice        float64   `db:"USD_PRICE"`
	NetHashesPerSec float64   `db:"NET_HASHES_SEC"`
	BlockTime       int64     `db:"BLOCK_TIME"`
	BlockReward     float64   `db:"BLOCK_REWARD"`
	ProfitPerMonth  float64   `db:"PROFIT_PER_MONTH_USD"`
	ProfitSNM       float64   `db:"PROFIT_SNM"`
	DateTime        time.Time `db:"DATE_TIME"`
}

type DealDb struct {
	DealID              int64 `db:"ID"`
	Status              int64 `db:"STATUS"`
	Price               int64 `db:"PRICE"`
	AskID               int64 `db:"ASK_ID"`
	BidID               int64 `db:"BID_ID"`
	DeployStatus        int64 `db:"DEPLOY_STATUS"`
	ChangeRequestStatus int64 `db:"CHANGE_REQUEST_STATUS"`
}

type OrderDb struct {
	OrderID         int64     `db:"ID"`
	Price           int64     `db:"PRICE"`
	Hashrate        uint64    `db:"HASHRATE"`
	StartTime       time.Time `db:"START_TIME"`
	ButterflyEffect int64     `db:"BUTTERFLY_EFFECT"`
	ActualStep      float64   `db:"ACTUAL_STEP"`
}
