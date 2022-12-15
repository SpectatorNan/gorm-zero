package gormc

import (
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type GormLogConfigI interface {
	GetGormLogMode() logger.LogLevel
	GetSlowThreshold() time.Duration
	GetColorful() bool
}

func newDefaultGormLogger(cfg GormLogConfigI) logger.Interface {
	newLogger := logger.New(
		log.New(os.Stderr, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             cfg.GetSlowThreshold(), // 慢 SQL 阈值
			LogLevel:                  cfg.GetGormLogMode(),   // 日志级别
			IgnoreRecordNotFoundError: true,                   // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  cfg.GetColorful(),      // 禁用彩色打印
		},
	)
	return newLogger
}

func overwriteGormLogMode(mode string) logger.LogLevel {
	switch mode {
	case "dev":
		return logger.Info
	case "test":
		return logger.Warn
	case "prod":
		return logger.Error
	case "silent":
		return logger.Silent
	default:
		return logger.Info
	}
}
