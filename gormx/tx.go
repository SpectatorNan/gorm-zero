package gormx

import "database/sql"

type (
	beginnable func(*sql.DB) (trans, error)

	trans interface {
		Session
		Commit() error
		Rollback() error
	}

	txSession struct {
		*sql.Tx
	}
)