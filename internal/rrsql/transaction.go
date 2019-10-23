package rrsql

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
)

// TxFn is the function that really executes sql queries
type TxFn func(*sqlx.Tx) error

// WithTransaction is a wrapper that creates db transaction and handles rollback/commit based on the
// error object returned by the `TxFn`
func WithTransaction(db *sqlx.DB, fn TxFn) (err error) {
	tx, err := db.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			tx.Rollback()
			panic(p)
		} else if err != nil {
			// something went wrong, rollback
			tx.Rollback()
		} else {
			// all good, commit
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

type PipelineStmt struct {
	Query        string
	Args         []interface{}
	NamedArgs    interface{}
	NamedExec    bool
	RowsAffected bool
	LastInsertId bool
}

// Runs the supplied statements within the transaction. If any statement fails, the transaction
// is rolled back, and the original error is returned.
func RunPipeline(tx *sqlx.Tx, stmts ...*PipelineStmt) (int64, sql.Result, error) {
	var res sql.Result
	var err error
	var lastInsId, rowCnt int64

	for _, ps := range stmts {

		if lastInsId != 0 {
			lastInsIdString := strconv.Itoa(int(lastInsId))
			ps.Query = strings.Replace(ps.Query, config.Config.SQL.TrasactionIDPlaceholder, lastInsIdString, -1)
		}

		if ps.NamedExec {
			res, err = tx.NamedExec(ps.Query, ps.NamedArgs)
		} else {
			res, err = tx.Exec(ps.Query, ps.Args...)
		}
		if err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				return lastInsId, nil, DuplicateError
			} else {
				return lastInsId, nil, err
			}
		}
		if ps.RowsAffected {
			rowCnt, err = res.RowsAffected()
			if rowCnt > 1 {
				return lastInsId, nil, MultipleRowAffectedError
			} else if rowCnt == 0 {
				return lastInsId, nil, ItemNotFoundError
			}
		}
		if ps.LastInsertId {
			lastInsId, err = res.LastInsertId()
			if err != nil {
				return lastInsId, nil, err
			}
		}
	}

	return lastInsId, res, nil
}
