package gormx

import (
	"gorm.io/gorm"
)

type (
	beginnable func(*gorm.DB) (trans, error)

	trans interface {
		Session
		Commit() *gorm.DB
		Rollback() *gorm.DB
	}

	txSession struct {
		Tx *gorm.DB
	}
)

func (tx txSession) Commit() *gorm.DB {
	return tx.Tx.Commit()
}
func (tx txSession) Rollback() *gorm.DB {
	return tx.Tx.Rollback()
}
func begin(db *gorm.DB) (trans, error) {
	tx := db.Begin()

	return txSession{
		Tx: tx,
	}, nil
}