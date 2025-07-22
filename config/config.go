package config

import (
	"github.com/SpectatorNan/gorm-zero/v2/logger"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type GormLogConfigI interface {
	GetGormLogMode() gormLogger.LogLevel
	GetSlowThreshold() time.Duration
	GetColorful() bool
}

func NewDefaultZeroLogger(cfg GormLogConfigI) gormLogger.Interface {
	newLogger := logger.NewZeroLog(
		gormLogger.Config{
			SlowThreshold:             cfg.GetSlowThreshold(), // 慢 SQL 阈值
			LogLevel:                  cfg.GetGormLogMode(),   // 日志级别
			IgnoreRecordNotFoundError: true,                   // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  cfg.GetColorful(),      // 禁用彩色打印
		},
	)
	return newLogger
}

func NewDefaultGormLogger(cfg GormLogConfigI) gormLogger.Interface {
	newLogger := gormLogger.New(
		log.New(os.Stderr, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		gormLogger.Config{
			SlowThreshold:             cfg.GetSlowThreshold(), // 慢 SQL 阈值
			LogLevel:                  cfg.GetGormLogMode(),   // 日志级别
			IgnoreRecordNotFoundError: true,                   // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  cfg.GetColorful(),      // 禁用彩色打印
		},
	)
	return newLogger
}

func OverwriteGormLogMode(mode string) gormLogger.LogLevel {
	switch mode {
	case "dev":
		return gormLogger.Info
	case "test":
		return gormLogger.Warn
	case "prod":
		return gormLogger.Error
	case "silent":
		return gormLogger.Silent
	default:
		return gormLogger.Info
	}
}
