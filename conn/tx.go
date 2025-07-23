package conn

import (
	"gorm.io/gen"
	"gorm.io/gorm"
)

type ConnTx struct {
	db    *gorm.DB
	Error error
	opts  []gen.DOOption // Store options
}

func NewConnFromTx[T any](tx *ConnTx) Conn[T] {
	return NewConn[T](tx.db, tx.opts...)
}

func (c *ConnTx) Commit() error {
	return c.db.Commit().Error
}

func (c *ConnTx) Rollback() error {
	return c.db.Rollback().Error
}

func (c *ConnTx) SavePoint(name string) error {
	return c.db.SavePoint(name).Error
}

func (c *ConnTx) RollbackTo(name string) error {
	return c.db.RollbackTo(name).Error
}

func (c *ConnTx) UnderlyingDB() *gorm.DB {
	return c.db
}
