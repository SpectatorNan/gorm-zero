package conn

import (
	"database/sql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

type ConnOld[T any] struct {
	db *gorm.DB
	//Repository[T]
	CommonRepo[T]
	opts []gen.DOOption // Store options
}

// NewConn creates a new connection with the given gorm.DB and options.
func NewConnOld1[T any](db *gorm.DB, opts ...gen.DOOption) ConnOld[T] {

	repo := CommonRepo[T]{}
	var model T
	repo.UseDB(db, opts...)
	repo.UseModel(&model)
	return NewConnOldWithRepo[T](db, repo, opts...)
}

func NewConnOldWithRepo[T any](db *gorm.DB, repo CommonRepo[T], opts ...gen.DOOption) ConnOld[T] {
	conn := ConnOld[T]{
		db:         db,
		CommonRepo: repo,
		opts:       opts,
	}

	return conn
}

func (c ConnOld[T]) clone(db *gorm.DB) ConnOld[T] {
	repo := c.CommonRepo
	repo.ReplaceDB(db)
	return ConnOld[T]{
		db:         db,
		CommonRepo: repo,
		opts:       c.opts,
	}
}

// do functions
//func (c ConnOld[T]) UseTable(newTableName string) {
//	c.UseTable(newTableName)
//}

//func (c ConnOld[T]) As(alias string) {
//	c.DO = *(c.DO.As(alias).(*gen.DO))
//}

// transaction
func (c ConnOld[T]) Transaction(fc func(tx *ConnTx) error, opts ...*sql.TxOptions) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		connTx := &ConnTx{
			db:    tx,
			Error: tx.Error,
			opts:  c.opts,
		}
		if connTx.Error != nil {
			return connTx.Error
		}
		return fc(connTx)
	}, opts...)
}

func (c ConnOld[T]) Begin(opts ...*sql.TxOptions) *ConnTx {
	tx := c.db.Begin(opts...)
	return &ConnTx{
		db:    tx,
		Error: tx.Error,
		opts:  c.opts,
	}
}
