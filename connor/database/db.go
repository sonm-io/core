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

func (d *Database) DeleteOrder(id int64) error {
	_, err := d.connect.Exec("DELETE FROM ORDERS WHERE ID=?", id)
	if err != nil {
		return err
	}
	return nil

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

func (d *Database) CreateOrderDB() error {
	_, err := d.connect.Exec(orders)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) CreatePoolDB() error {
	_, err := d.connect.Exec(pools)
	if err != nil {
		return err
	}
	// TODO: check error in result
	return nil
}
func (d *Database) CreateDealsDB() error {
	_, err := d.connect.Exec(deals)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateBlacklistDB() error {
	_, err := d.connect.Exec(blacklist)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) CreateTokenDb() error {
	_, err := d.connect.Exec(tokens)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) CreateAllTables() error {
	if err := d.CreateTokenDb(); err != nil {
		return fmt.Errorf("cannot create token DB %v", err)
	}
	if err := d.CreateDealsDB(); err != nil {
		return fmt.Errorf("cannot create deals DB %v", err)
	}
	if err := d.CreateOrderDB(); err != nil {
		return fmt.Errorf("cannot create orders DB %v", err)
	}
	if err := d.CreatePoolDB(); err != nil {
		return fmt.Errorf("cannot create pool DB %v", err)
	}
	if err := d.CreateBlacklistDB(); err != nil {
		return fmt.Errorf("cannot create blacklist DB %v", err)
	}
	return nil
}

func (d *Database) SaveDealIntoDB(deal *DealDb) error {
	_, err := d.connect.Exec(deals)
	if err != nil {
		return err
	}
	tx := d.connect.MustBegin()
	tx.NamedExec(insertDeals, deal)
	tx.Commit()
	return nil
}

func (d *Database) SaveBlacklistIntoDB(blacklistData *BlackListDb) error {
	_, err := d.connect.Exec(blacklist)
	if err != nil {
		return err
	}
	tx := d.connect.MustBegin()
	tx.NamedExec(insertBlackList, blacklistData)
	tx.Commit()
	return nil
}

func (d *Database) SaveOrderIntoDB(order *OrderDb) error {
	_, err := d.connect.Exec(orders)
	if err != nil {
		return err
	}
	tx := d.connect.MustBegin()
	tx.NamedExec(insertOrders, order)
	tx.Commit()
	return nil
}
func (d *Database) SaveProfitToken(token *TokenDb) error {
	_, err := d.connect.Exec(tokens)
	if err != nil {
		return err
	}
	tx := d.connect.MustBegin()
	tx.NamedExec(insertToken, token)
	tx.Commit()
	return nil
}
func (d *Database) SavePoolIntoDB(pool *PoolDb) error {
	_, err := d.connect.Exec(pools)
	if err != nil {
		return err
	}
	tx := d.connect.MustBegin()
	tx.NamedExec(insertPools, pool)
	tx.Commit()
	return nil
}

func (d *Database) UpdateDealStatusDb(id int64, status sonm.DealStatus) error {
	_, err := d.connect.Exec(updateStatusDeal, status, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) UpdateOrderInDB(id int64, bfly int64) error {
	_, err := d.connect.Exec(updateOrders, bfly, id)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) UpdateDeployStatusDealInDB(id int64, deployStatus int64) error {
	_, err := d.connect.Exec(updateDeployStatusDeal, deployStatus, id)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) UpdateBadGayStatusInPoolDB(id int64, badGuy int64, timeUpdate time.Time) error {
	_, err := d.connect.Exec(updateStatusPoolDB, badGuy, timeUpdate, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) UpdateChangeRequestStatusDealDB(id int64, status sonm.ChangeRequestStatus, price int64) error {
	_, err := d.connect.Exec(updateCRStatusDeal, status, price, id)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) ReturnChangeRequestStatusDealDB(id int64, status sonm.ChangeRequestStatus) error {
	_, err := d.connect.Exec(returnCRStatusDeal, status, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) UpdateReportedHashratePoolDB(id string, reportedHashrate float64, timeUpdate time.Time) error {
	_, err := d.connect.Exec(updateReportedHashrate, reportedHashrate, timeUpdate, id)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) UpdateAvgPoolDB(id string, avgHashrate float64, timeUpdate time.Time) error {
	_, err := d.connect.Exec(updateAvgPool, avgHashrate, timeUpdate, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) UpdateBanStatusBlackListDB(masterID string, banStatus int64) error {
	_, err := d.connect.Exec(updateBlackList, banStatus, masterID)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) UpdateIterationPoolDB(iteration int64, id int64) error {
	_, err := d.connect.Exec(updateIterationPool, iteration, id)
	if err != nil {
		return err
	}
	return nil
}
func (d *Database) SetDestroyDealPoolDB(status int64, id int64) error {
	_, err := d.connect.Exec(setDestroyStatusDeal, status, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetCountFromDB() (counts int64, err error) {
	rows, err := d.connect.Query(getCountFromDb)
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
	rows, err := d.connect.Query(getLastActualStep)
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
	rows, err := d.connect.Query(getOrders)
	if err != nil {
		return nil, err
	}
	orders := make([]*OrderDb, 0)
	defer rows.Close()
	for rows.Next() {
		order := new(OrderDb)
		err := rows.Scan(&order.OrderID, &order.Price, &order.Hashrate, &order.StartTime, &order.ButterflyEffect, &order.ActualStep)
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
func (d *Database) GetBlacklistFromDb(failSupplierID string) (string, error) {
	rows, err := d.connect.Query(getSupplierIDFromBlackList, failSupplierID)
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

func (d *Database) GetWorkerFromPoolDb(dealID string) (string, error) {
	rows, err := d.connect.Query(getWorkerIDFromPool, dealID)
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

func (d *Database) GetChangeRequestStatus(dealId int64)(int64, error){
	rows, err := d.connect.Query(getChangeRequestStatus, dealId)
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

func (d *Database) GetCountFailSupplierFromDb(masterID string) (int64, error) {
	rows, err := d.connect.Query(getCountSupplierIDFromBlackList, masterID)
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

func (d *Database) GetDealsFromDB() ([]*DealDb, error) {
	rows, err := d.connect.Query(getDeals)
	if err != nil {
		return nil, err
	}
	deals := make([]*DealDb, 0)
	defer rows.Close()
	for rows.Next() {
		deal := new(DealDb)
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
func (d *Database) GetWorkersFromDB() ([]*PoolDb, error) {
	rows, err := d.connect.Query(getWorkersFromPool)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var workers []*PoolDb
	for rows.Next() {
		worker := new(PoolDb)
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
