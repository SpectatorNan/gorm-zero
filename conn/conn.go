package conn

import (
	"database/sql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

type Conn[T any] struct {
	db *gorm.DB
	Repository[T]
	opts []gen.DOOption // Store options
}

// NewConn creates a new connection with the given gorm.DB and options.
func NewConn[T any](db *gorm.DB, opts ...gen.DOOption) Conn[T] {

	repo := CommonRepo[T]{}
	var model T
	repo.UseDB(db, opts...)
	repo.UseModel(&model)
	return NewConnWithRepo(db, repo, opts...)
}

func NewConnWithRepo[T any](db *gorm.DB, repo Repository[T], opts ...gen.DOOption) Conn[T] {
	conn := Conn[T]{
		db:         db,
		Repository: repo,
		opts:       opts,
	}

	return conn
}

func (c Conn[T]) clone(db *gorm.DB) Conn[T] {
	repo := c.Repository
	repo.ReplaceDB(db)
	return Conn[T]{
		db:         db,
		Repository: repo,
		opts:       c.opts,
	}
}

func (c Conn[T]) Transaction(fc func(tx *ConnTx) error, opts ...*sql.TxOptions) error {
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

func (c Conn[T]) Begin(opts ...*sql.TxOptions) *ConnTx {
	tx := c.db.Begin(opts...)
	return &ConnTx{
		db:    tx,
		Error: tx.Error,
		opts:  c.opts,
	}
}
