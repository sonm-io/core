package dwh

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

type queryConn interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	// Any time a queryConn is created, `defer conn.Finish()` should be called.
	Finish() error
}

type simpleConn struct {
	db *sql.DB
}

func newSimpleConn(db *sql.DB) queryConn {
	return &simpleConn{db: db}
}

func (t *simpleConn) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.db.Exec(query, args...)
}

func (t *simpleConn) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.db.Query(query, args...)
}

func (t *simpleConn) Finish() error {
	return nil
}

// txConn implements covert transaction rollbacks/commits based on whether there was any errors
// while interacting with DB.
type txConn struct {
	tx        *sql.Tx
	logger    *zap.Logger
	hasErrors bool
}

func newTxConn(db *sql.DB, logger *zap.Logger) (queryConn, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &txConn{tx: tx, logger: logger}, nil
}

func (t *txConn) Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := t.tx.Exec(query, args...)
	if err != nil {
		t.hasErrors = true
		return nil, fmt.Errorf("failed to exec %s: %v", query, err)
	}
	return result, nil
}

func (t *txConn) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := t.tx.Query(query, args...)
	if err != nil {
		t.hasErrors = true
		return nil, fmt.Errorf("failed to run %s: %v", query, err)
	}
	return rows, nil
}

func (t *txConn) Finish() error {
	if t.hasErrors {
		if err := t.tx.Rollback(); err != nil {
			t.logger.Warn("transaction rollback failed")
			return err
		}
	} else {
		if err := t.tx.Commit(); err != nil {
			t.logger.Warn("transaction rollback failed")
			return err
		}
	}
	return nil
}
