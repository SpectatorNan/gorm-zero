package gormc

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type PgSql struct {
	Username      string
	Password      string
	Path          string
	Port          int    `json:",default=5432"`
	SslMode       string `json:",default=disable,options=disable|enable"`
	TimeZone      string `json:",default=Asia/Shanghai"`
	Dbname        string
	MaxIdleConns  int    `json:",default=10"`                               // 空闲中的最大连接数
	MaxOpenConns  int    `json:",default=10"`                               // 打开到数据库的最大连接数
	LogMode       string `json:",default=dev,options=dev|test|prod|silent"` // 是否开启Gorm全局日志
	LogZap        bool   // 是否通过zap写入日志文件
	SlowThreshold int64  `json:",default=1000"`
}

func (m *PgSql) Dsn() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s TimeZone=%s", m.Username, m.Password, m.Dbname, m.Path, m.Port, m.SslMode, m.TimeZone)
}
func (m *PgSql) GetGormLogMode() logger.LogLevel {
	return overwriteGormLogMode(m.LogMode)
}

func (m *PgSql) GetSlowThreshold() time.Duration {
	return time.Duration(m.SlowThreshold) * time.Millisecond
}
func (m *PgSql) GetColorful() bool {
	return true
}

func ConnectPgSql(m PgSql) (*gorm.DB, error) {
	if m.Dbname == "" {
		return nil, errors.New("database name is empty")
	}
	newLogger := newDefaultGormLogger(&m)
	pgsqlCfg := postgres.Config{
		DSN:                  m.Dsn(),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}
	db, err := gorm.Open(postgres.New(pgsqlCfg), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	} else {
		sqldb, _ := db.DB()
		sqldb.SetMaxIdleConns(m.MaxIdleConns)
		sqldb.SetMaxOpenConns(m.MaxOpenConns)
		return db, nil
	}
}
