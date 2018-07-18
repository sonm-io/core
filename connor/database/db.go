package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sonm-io/core/proto"
)

type Database struct {
	connect *sqlx.DB
}

func NewDatabaseConnect(driver, dataSource string) (*Database, error) {
	var err error
	d := &Database{}
	d.connect, err = sqlx.Connect(driver, dataSource)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Database) CreateOrdersTable() error {
	_, err := d.connect.Exec(createOrdersTableSQL)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) CreatePoolTable() error {
	_, err := d.connect.Exec(createPoolsTableSQL)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) CreateDealsTable() error {
	_, err := d.connect.Exec(createDealsTableSQL)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) CreateBlacklistTable() error {
	_, err := d.connect.Exec(createBlacklistTableSQL)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) CreateTokensTable() error {
	_, err := d.connect.Exec(createTokensTableSQL)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) CreateAllTables() error {
	if err := d.CreateTokensTable(); err != nil {
		return fmt.Errorf("cannot create token DB: %v", err)
	}
	if err := d.CreateDealsTable(); err != nil {
		return fmt.Errorf("cannot create deals DB: %v", err)
	}
	if err := d.CreateOrdersTable(); err != nil {
		return fmt.Errorf("cannot create orders DB: %v", err)
	}
	if err := d.CreatePoolTable(); err != nil {
		return fmt.Errorf("cannot create pool DB: %v", err)
	}
	if err := d.CreateBlacklistTable(); err != nil {
		return fmt.Errorf("cannot create blacklist DB: %v", err)
	}
	return nil
}

func (d *Database) SaveProfitToken(token *TokenDb) error {
	_, err := d.connect.Exec(createTokensTableSQL)
	if err != nil {
		return err
	}
	tx := d.connect.MustBegin()
	tx.NamedExec(insertTokenQuery, token)
	tx.Commit()
	return nil
}
func (d *Database) SaveTestOrderIntoDB(order *OrderDb) error {
	_, err := d.connect.Exec(createOrdersTableSQL)
	if err != nil {
		return err
	}
	tx := d.connect.MustBegin()
	tx.NamedExec(insertTestOrderQuery, order)
	tx.Commit()
	return nil
}
func (d *Database) SaveOrderIntoDB(order *OrderDb) error {
	_, err := d.connect.Exec(createOrdersTableSQL)
	if err != nil {
		return err
	}
	tx := d.connect.MustBegin()
	tx.NamedExec(insertOrderQuery, order)
	tx.Commit()
	return nil
}
func (d *Database) SaveDealIntoDB(deal *DealDB) error {
	_, err := d.connect.Exec(createDealsTableSQL)
	if err != nil {
		return err
	}
	tx := d.connect.MustBegin()
	tx.NamedExec(insertDealsQuery, deal)
	tx.Commit()
	return nil
}
func (d *Database) SavePoolIntoDB(pool *PoolDB) error {
	_, err := d.connect.Exec(createPoolsTableSQL)
	if err != nil {
		return err
	}
	tx := d.connect.MustBegin()
	tx.NamedExec(insertPoolsQuery, pool)
	tx.Commit()
	return nil
}
func (d *Database) SaveBlacklistIntoDB(blacklistData *BlackListDb) error {
	_, err := d.connect.Exec(createBlacklistTableSQL)
	if err != nil {
		return err
	}
	tx := d.connect.MustBegin()
	tx.NamedExec(insertBlackListQuery, blacklistData)
	tx.Commit()
	return nil
}

func (d *Database) UpdateOrderInDB(id int64, bfly int64) error {
	_, err := d.connect.Exec(updateOrdersQuery, bfly, id)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) UpdateDeployAndDealStatusDB(id int64, deployStatus int64, status sonm.DealStatus) error {
	_, err := d.connect.Exec(updateDeployAndDealStatusQuery, deployStatus, status, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) UpdateDeployStatusDealInDB(id int64, deployStatus int64) error {
	_, err := d.connect.Exec(updateDeployStatusDealQuery, deployStatus, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) UpdateBadGayStatusInPoolDB(id int64, badGuy int64, timeUpdate time.Time) error {
	_, err := d.connect.Exec(updateStatusInPoolQuery, badGuy, timeUpdate, id)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) UpdateChangeRequestStatusDealDB(id int64, status sonm.ChangeRequestStatus, price int64) error {
	_, err := d.connect.Exec(updateDealCRPriceQuery, status, price, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) UpdateReportedHashratePoolDB(id string, reportedHashrate float64, timeUpdate time.Time) error {
	_, err := d.connect.Exec(updateReportedHashrateQuery, reportedHashrate, timeUpdate, id)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) UpdateAvgPoolDB(id string, avgHashrate float64, timeUpdate time.Time) error {
	_, err := d.connect.Exec(updateAvgPoolQuery, avgHashrate, timeUpdate, id)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) UpdateBanStatusBlackListDB(masterID string, banStatus int64) error {
	_, err := d.connect.Exec(updateBlacklistQuery, banStatus, masterID)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) UpdateIterationPoolDB(iteration int64, id int64) error {
	_, err := d.connect.Exec(updateIterationPoolQuery, iteration, id)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) UpdateDestroyDealPoolDB(status int64, id int64) error {
	_, err := d.connect.Exec(updateDealDeployStatusQuery, status, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetCountFromDB() (counts int64, err error) {
	rows, err := d.connect.Query(selectCountFromQuery)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var count int64
		err = rows.Scan(&count)
		if err != nil {
			return 0, err
		}
		return count, nil
	}
	return 0, fmt.Errorf("")
}
func (d *Database) GetLastActualStepFromDb() (float64, error) {
	rows, err := d.connect.Query(selectLastActualStepQuery)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var max float64
		err = rows.Scan(&max)
		if err != nil {
			return 0, err
		}
		return max, nil
	}
	return 0, nil
}
func (d *Database) GetOrdersFromDB() ([]*OrderDb, error) {
	rows, err := d.connect.Query(selectAllOrdersQuery)
	if err != nil {
		return nil, err
	}
	orders := make([]*OrderDb, 0)
	defer rows.Close()
	for rows.Next() {
		order := new(OrderDb)
		err := rows.Scan(&order.OrderID, &order.Price, &order.Hashrate, &order.StartTime, &order.Status, &order.ActualStep)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orders, err
}
func (d *Database) GetDealsFromDB() ([]*DealDB, error) {
	rows, err := d.connect.Query(selectAllDealsQuery)
	if err != nil {
		return nil, err
	}
	deals := make([]*DealDB, 0)
	defer rows.Close()
	for rows.Next() {
		deal := new(DealDB)
		err := rows.Scan(&deal.DealID, &deal.Status, &deal.Price, &deal.AskID, &deal.BidID, &deal.DeployStatus, &deal.ChangeRequestStatus)
		if err != nil {
			return nil, err
		}
		deals = append(deals, deal)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return deals, err
}
func (d *Database) GetWorkersFromDB() ([]*PoolDB, error) {
	rows, err := d.connect.Query(selectAllWorkersQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var workers []*PoolDB
	for rows.Next() {
		worker := new(PoolDB)
		err := rows.Scan(&worker.DealID, &worker.PoolID, &worker.WorkerReportedHashrate,
			&worker.WorkerAvgHashrate, &worker.BadGuy, &worker.Iterations, &worker.TimeStart, &worker.TimeUpdate)
		if err != nil {
			return nil, err
		}
		workers = append(workers, worker)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return workers, err
}
func (d *Database) GetChangeRequestStatusDealDB(id int64, status sonm.ChangeRequestStatus) error {
	_, err := d.connect.Exec(updateDealCRQuery, status, id)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) GetFailSupplierFromBlacklistDb(failSupplierID string) (string, error) {
	rows, err := d.connect.Query(selectSupplierIDFromBlackListQuery, failSupplierID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var failSupplier string
		err = rows.Scan(&failSupplier)
		if err != nil {
			return "", err
		}
		return failSupplier, nil
	}
	return "", nil
}
func (d *Database) GetMasterBlacklist(masterId string) (string, error) {
	rows, err := d.connect.Query(selectMasterIDFromBlackListQuery, masterId)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var masterId string
		err = rows.Scan(&masterId)
		if err != nil {
			return "", err
		}
		return masterId, nil
	}
	return "", nil
}
func (d *Database) GetWorkerFromPoolDB(dealID string) (string, error) {
	rows, err := d.connect.Query(selectWorkerIDFromPoolQuery, dealID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var dealID string
		err = rows.Scan(&dealID)
		if err != nil {
			return "", err
		}
		return dealID, nil
	}
	return "already in Pool!", nil
}
func (d *Database) GetChangeRequestStatus(dealId int64) (int64, error) {
	rows, err := d.connect.Query(selectChangeRequestStatusQuery, dealId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var changeRequestStatus int64
		err = rows.Scan(&changeRequestStatus)
		if err != nil {
			return 0, err
		}
		return changeRequestStatus, nil
	}
	return 0, nil
}
func (d *Database) GetDeployStatus(dealId int64) (int64, error) {
	rows, err := d.connect.Query(selectDeployStatusQuery, dealId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var deployStatus int64
		err = rows.Scan(&deployStatus)
		if err != nil {
			return 0, err
		}
		return deployStatus, nil
	}
	return 0, nil
}
func (d *Database) GetCountFailSupplierFromDb(masterID string) (int64, error) {
	rows, err := d.connect.Query(selectSupplierCountFromBlackListQuery, masterID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var failSupplier int64
		err = rows.Scan(&failSupplier)
		if err != nil {
			return 0, err
		}
		return failSupplier, nil
	}
	return 0, nil
}
