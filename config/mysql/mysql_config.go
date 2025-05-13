package mysql

import (
	"errors"
	"fmt"
	"github.com/SpectatorNan/gorm-zero/config"
	"github.com/SpectatorNan/gorm-zero/plugins"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type Mysql struct {
	Path          string // 服务器地址
	Port          int    `json:",default=3306"`                                               // 端口
	Config        string `json:",default=charset%3Dutf8mb4%26parseTime%3Dtrue%26loc%3DLocal"` // 高级配置
	Dbname        string // 数据库名
	Username      string // 数据库用户名
	Password      string // 数据库密码
	MaxIdleConns  int    `json:",default=10"` // 空闲中的最大连接数
	MaxOpenConns  int    `json:",default=10"` // 打开到数据库的最大连接数
	LogMode       string `json:",default=dev,options=dev|test|prod|silent"`
	LogColorful   bool   `json:",default=false"` // 是否开启日志高亮
	SlowThreshold int64  `json:",default=1000"`
}

func (m *Mysql) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + fmt.Sprintf("%d", m.Port) + ")/" + m.Dbname + "?" + m.Config
}

func (m *Mysql) GetGormLogMode() logger.LogLevel {
	return config.OverwriteGormLogMode(m.LogMode)
}

func (m *Mysql) GetSlowThreshold() time.Duration {
	return time.Duration(m.SlowThreshold) * time.Millisecond
}
func (m *Mysql) GetColorful() bool {
	return m.LogColorful
}

func Connect(m Mysql) (*gorm.DB, error) {
	if m.Dbname == "" {
		return nil, errors.New("database name is empty")
	}
	mysqlCfg := mysql.Config{
		DSN: m.Dsn(),
	}
	newLogger := config.NewDefaultGormLogger(&m)
	db, err := gorm.Open(mysql.New(mysqlCfg), &gorm.Config{
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

func ConnectWithConfig(m Mysql, cfg *gorm.Config) (*gorm.DB, error) {
	if m.Dbname == "" {
		return nil, errors.New("database name is empty")
	}
	mysqlCfg := mysql.Config{
		DSN: m.Dsn(),
	}
	db, err := gorm.Open(mysql.New(mysqlCfg), cfg)
	if err != nil {
		return nil, err
	}

	err = plugins.InitPlugins(db)
	if err != nil {
		return nil, err
	}

	sqldb, _ := db.DB()
	sqldb.SetMaxIdleConns(m.MaxIdleConns)
	sqldb.SetMaxOpenConns(m.MaxOpenConns)
	return db, nil

}
