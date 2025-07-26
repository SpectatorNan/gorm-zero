package conn

import (
	"database/sql"
	"gorm.io/gorm"
)

type Conn struct {
	db *gorm.DB
}

func NewConn(db *gorm.DB) *Conn {
	return &Conn{
		db: db,
	}
}

func (c Conn) DB() *gorm.DB {
	return c.db
}
func (c Conn) Transaction(fc func(tx *ConnTx) error, opts ...*sql.TxOptions) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		connTx := &ConnTx{
			db:    tx,
			Error: tx.Error,
			opts:  nil, // No options for Conn
		}
		if connTx.Error != nil {
			return connTx.Error
		}
		return fc(connTx)
	}, opts...)
}
