// transaction
package gorm

import (
	"context"
	"database/sql"
	"strings"

	"github.com/satori/go.uuid"
)

type SqlDB struct {
	*sql.DB
}

type sqlTransaction interface {
	Commit() error
	Rollback() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) (*sql.Row, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) (*sql.Row, error)
	Stmt(stmt *sql.Stmt) *sql.Stmt
	StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
}

func (db *SqlDB) Begin() (sqlTx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &SqlTx{Tx: tx}, nil
}

type SqlTx struct {
	*sql.Tx

	sps []string
}

func (tx *SqlTx) Begin() (sqlTx, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	id := strings.Replace(uid.String(), "-", "", -1)

	sql := "savepoint " + id
	_, err = tx.Exec(sql)
	if err != nil {
		return nil, err
	}
	tx.sps = append(tx.sps, id)
	return tx, nil
}

func (tx *SqlTx) Commit() error {
	if len(tx.sps) == 0 {
		return tx.Tx.Commit()
	}

	n := len(tx.sps)
	sql := "release savepoint " + tx.sps[n-1]
	_, err := tx.Exec(sql)
	if err != nil {
		return err
	}
	tx.sps = tx.sps[:n-1]
	return nil
}

func (tx *SqlTx) Rollback() error {
	if len(tx.sps) == 0 {
		return tx.Tx.Rollback()
	}

	n := len(tx.sps)
	sql := "rollback to savepoint " + tx.sps[n-1]
	_, err := tx.Exec(sql)
	if err != nil {
		return err
	}
	tx.sps = tx.sps[:n-1]
	return nil
}
