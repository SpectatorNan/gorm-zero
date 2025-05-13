package logger

import (
	"context"
	"errors"
	gormLogger "gorm.io/gorm/logger"
	"testing"
	"time"
)

func TestZeroLog(t *testing.T) {
	log := NewZeroLog(gormLogger.Config{
		SlowThreshold:             10 * time.Millisecond,
		LogLevel:                  gormLogger.Info,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})
	traceFc := func() (sql string, rowsAffected int64) {
		return "select * from test", 1
	}
	cases := []struct {
		name string
		exec func()
	}{
		{
			name: "normal log",
			exec: func() {
				log.Info(context.Background(), "normal test log")
				log.Trace(context.Background(), time.Now(), traceFc, nil)
			},
		},
		{
			name: "warn log",
			exec: func() {
				log.Warn(context.Background(), "warn test log")
				log.Trace(context.Background(), time.Now(), traceFc, nil)
			},
		},
		{
			name: "slow log",
			exec: func() {
				log.Trace(context.Background(), time.Now().Add(-11*time.Millisecond), traceFc, nil)
			},
		},
		{
			name: "error log",
			exec: func() {
				log.Error(context.Background(), "error test log")
				err := errors.New("error test log")
				log.Trace(context.Background(), time.Now(), traceFc, err)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.exec()
		})
	}

}
