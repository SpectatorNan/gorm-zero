package gormx

import (
	"database/sql"
	"errors"
	"github.com/tal-tech/go-zero/core/syncx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"sync"
	"time"
)

type DriverType int
const (
	MysqlDriver DriverType = iota
	PostgresDriver
)

const (
	maxIdleConns = 64
	maxOpenConns = 64
	maxLifetime  = time.Minute
)

var connManager = syncx.NewResourceManager()

type pingedDB struct {
	gormDB *gorm.DB
	*sql.DB
	once sync.Once
}
func getCachedSqlConn(driver DriverType, server string) (*pingedDB, error) {
	val, err := connManager.GetResource(server, func() (io.Closer, error) {
		gormDB, conn, err := newDBConnection(driver, server)
		if err != nil {
			return nil, err
		}

		return &pingedDB{
			gormDB: gormDB,
			DB: conn,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(*pingedDB), nil
}
func getSqlConn(driver DriverType, server string) (*gorm.DB, error) {
	pdb, err := getCachedSqlConn(driver, server)
	if err != nil {
		return nil, err
	}
	// gorm has auto ping
	/*
	pdb.once.Do(func() {
		err = pdb.Ping()
	})
	*/
	if err != nil {
		return nil, err
	}

	return pdb.gormDB, nil
}
func newMysqlConnect(datasource string, dbCfg gorm.Config) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(datasource), &dbCfg)
	if err != nil {
		return nil, nil, err
	}

	conn, err := db.DB()

	// we need to do this until the issue https://github.com/golang/go/issues/9851 get fixed
	// discussed here https://github.com/go-sql-driver/mysql/issues/257
	// if the discussed SetMaxIdleTimeout methods added, we'll change this behavior
	// 8 means we can't have more than 8 goroutines to concurrently access the same database.
	conn.SetMaxIdleConns(maxIdleConns)
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetConnMaxLifetime(maxLifetime)

	return db, conn, nil
}
func newPostgresConnect(datasource string, dbCfg gorm.Config) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(datasource), &dbCfg)
	if err != nil {
		return nil, nil, err
	}

	conn, err := db.DB()

	// we need to do this until the issue https://github.com/golang/go/issues/9851 get fixed
	// discussed here https://github.com/go-sql-driver/mysql/issues/257
	// if the discussed SetMaxIdleTimeout methods added, we'll change this behavior
	// 8 means we can't have more than 8 goroutines to concurrently access the same database.
	conn.SetMaxIdleConns(maxIdleConns)
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetConnMaxLifetime(maxLifetime)

	return db, conn, nil
}
func newDBConnection(driver DriverType, server string) (*gorm.DB, *sql.DB, error) {
	switch driver {
	case MysqlDriver:
		return newMysqlConnect(server, gorm.Config{})
	case PostgresDriver:
		return newPostgresConnect(server, gorm.Config{})
	default:
		return nil, nil, errors.New("un support db driver")
	}
}

/*
func newDBConnection(driverName, datasource string) (*sql.DB, error) {

	conn, err := sql.Open(driverName, datasource)
	if err != nil {
		return nil, err
	}

	// we need to do this until the issue https://github.com/golang/go/issues/9851 get fixed
	// discussed here https://github.com/go-sql-driver/mysql/issues/257
	// if the discussed SetMaxIdleTimeout methods added, we'll change this behavior
	// 8 means we can't have more than 8 goroutines to concurrently access the same database.
	conn.SetMaxIdleConns(maxIdleConns)
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetConnMaxLifetime(maxLifetime)

	return conn, nil
}
*/