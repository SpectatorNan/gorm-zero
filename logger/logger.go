package logger

import (
	"context"
	"errors"
	"fmt"
	"github.com/SpectatorNan/gorm-zero/utils"
	"github.com/zeromicro/go-zero/core/logx"
	gormLogger "gorm.io/gorm/logger"
	"time"
)

func NewZeroLog(config gormLogger.Config) gormLogger.Interface {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		infoStr = gormLogger.Green + "%s\n" + gormLogger.Reset + gormLogger.Green + "[info] " + gormLogger.Reset
		warnStr = gormLogger.BlueBold + "%s\n" + gormLogger.Reset + gormLogger.Magenta + "[warn] " + gormLogger.Reset
		errStr = gormLogger.Magenta + "%s\n" + gormLogger.Reset + gormLogger.Red + "[error] " + gormLogger.Reset
		traceStr = gormLogger.Green + "%s\n" + gormLogger.Reset + gormLogger.Yellow + "[%.3fms] " + gormLogger.BlueBold + "[rows:%v]" + gormLogger.Reset + " %s"
		traceWarnStr = gormLogger.Green + "%s " + gormLogger.Yellow + "%s\n" + gormLogger.Reset + gormLogger.RedBold + "[%.3fms] " + gormLogger.Yellow + "[rows:%v]" + gormLogger.Magenta + " %s" + gormLogger.Reset
		traceErrStr = gormLogger.RedBold + "%s " + gormLogger.MagentaBold + "%s\n" + gormLogger.Reset + gormLogger.Yellow + "[%.3fms] " + gormLogger.BlueBold + "[rows:%v]" + gormLogger.Reset + " %s"
	}

	return &logger{
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type logger struct {
	//logx.Logger
	gormLogger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *logger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Info {
		logx.WithContext(ctx).Infof(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Warn {
		logx.WithContext(ctx).Slowf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Error {
		logx.WithContext(ctx).Errorf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormLogger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gormLogger.ErrRecordNotFound)):
		sql, rows := fc()
		if rows == -1 {
			logx.WithContext(ctx).Errorf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logx.WithContext(ctx).Errorf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			logx.WithContext(ctx).Slowf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logx.WithContext(ctx).Slowf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == gormLogger.Info:
		sql, rows := fc()
		if rows == -1 {
			logx.WithContext(ctx).Infof(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logx.WithContext(ctx).Infof(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
