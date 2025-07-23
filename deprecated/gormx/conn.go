package gormx

import (
	"context"
	"gorm.io/gorm"
)

type (

	// ExecCtxFn defines the sql exec method.
	ExecCtxFn func(conn *gorm.DB) error

	Conn struct {
		*gorm.DB
	}
)

func NewConn(db *gorm.DB) Conn {
	return Conn{
		DB: db,
	}
}

//func (cc Conn) WithContextDB(ctx context.Context) *gorm.DB {
//	return cc.DB.WithContext(ctx)
//}

// Exec runs exec with given gorm.DB.
func (cc Conn) Exec(exec ExecCtxFn) error {
	return cc.ExecCtx(context.Background(), exec)
}

// ExecCtx runs exec with given gorm.DB.
func (cc Conn) ExecCtx(ctx context.Context, execCtx ExecCtxFn) (err error) {
	ctx, span := startSpan(ctx, "ExecCtx")
	defer func() {
		endSpan(span, err)
	}()
	return execCtx(cc.DB.WithContext(ctx))
}

