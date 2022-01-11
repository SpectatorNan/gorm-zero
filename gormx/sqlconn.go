package gormx

import (
	"database/sql"
	"github.com/tal-tech/go-zero/core/breaker"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/timex"
	"gorm.io/gorm"
)

var ErrNotFound = gorm.ErrRecordNotFound

type (
	Session interface {
		First(v interface{}) error
		QueryRow(v interface{}) error
		QueryRows(v interface{}) error
	}

	SqlConn interface {
		Session
		RawDB() (*sql.DB, error)
		Transact(func(session Session) error) error
	}

	// SqlOption defines the method to customize a sql connection.
	SqlOption func(*commonSqlConn)

	// thread-safe
	// Because CORBA doesn't support PREPARE, so we need to combine the
	// query arguments into one string and do underlying query without arguments
	commonSqlConn struct {
		connProv connProvider
		onError  func(error)
		beginTx  beginnable
		brk      breaker.Breaker
		accept   func(error) bool
	}

	connProvider func() (*gorm.DB, error)
)

// NewSqlConn returns a SqlConn with given driver name and datasource.
func NewSqlConn(driverName DriverType, datasource string, opts ...SqlOption) SqlConn {
	conn := &commonSqlConn{
		connProv: func() (*gorm.DB, error) {
			return getSqlConn(driverName, datasource)
		},
		onError: func(err error) {
			logInstanceError(datasource, err)
		},
		beginTx: begin,
		brk:     breaker.NewBreaker(),
	}
	for _, opt := range opts {
		opt(conn)
	}

	return conn
}

// NewSqlConnFromDB returns a SqlConn with the given sql.DB.
// Use it with caution, it's provided for other ORM to interact with.
func NewSqlConnFromDB(db *sql.DB, opts ...SqlOption) SqlConn {
	// wait implement
	/*
		conn := &commonSqlConn{
			connProv: func() (*sql.DB, error) {
				return db, nil
			},
			onError: func(err error) {
				logx.Errorf("Error on getting sql instance: %v", err)
			},
			beginTx: begin,
			brk:     breaker.NewBreaker(),
		}
		for _, opt := range opts {
			opt(conn)
		}

		return conn
	*/
	return nil
}

func (db *commonSqlConn) acceptable(err error) bool {
	ok := err == nil || err == gorm.ErrRecordNotFound || err == gorm.errtr
	if db.accept == nil {
		return ok
	}

	return ok || db.accept(err)
}

func (db *commonSqlConn) queryRows(engine func(db *gorm.DB) error, dest interface{}) error {
	var qerr error
	return db.brk.DoWithAcceptable(func() error {
		conn, err := db.connProv()
		if err != nil {
			db.onError(err)
			return err
		}
		return db.exec(func() error {
			qerr = conn.Find(&dest).Error
			return qerr
		})
		//return query(conn, func(rows *sql.Rows) error {
		//	qerr = scanner(rows)
		//	return qerr
		//}, q, args...)
	}, func(err error) bool {
		return qerr == err || db.acceptable(err)
	})
}

func (db *commonSqlConn) exec(execFn func() error) error {

	startTime := timex.Now()
	err := execFn()
	duration := timex.Since(startTime)
	if duration > slowThreshold.Load() {
		logx.WithDuration(duration).Slowf("[SQL] query: slowcall - %s", stmt)
	} else {
		logx.WithDuration(duration).Infof("sql query: %s", stmt)
	}
	if err != nil {
		logSqlError(stmt, err)
		return err
	}
	return nil
}
