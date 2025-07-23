package connx

import (
	"context"
	"github.com/SpectatorNan/gorm-zero/v2/conn"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/syncx"
	"time"
)

var (
	// ErrNotFound is an alias of sqlx.ErrNotFound.
	ErrNotFound = sqlx.ErrNotFound

	// can't use one SingleFlight per conn, because multiple conns may share the same cache key.
	singleFlights = syncx.NewSingleFlight()
	stats         = cache.NewStat("gorm-zero")
)

const (
	// see doc/sql-cache.md
	cacheSafeGapBetweenIndexAndPrimary = time.Second * 5

	// traceName is used to identify the trace name for the SQL execution.
	traceName = "gorm-zero-connx"

	// spanName is used to identify the span name for the SQL execution.
	spanName = "sql"

	// make the expiry unstable to avoid lots of cached items expire at the same time
	// make the unstable expiry to be [0.95, 1.05] * seconds
	expiryDeviation = 0.05

	defaultBatchSize = 500
)

type (
	// IndexQueryCtxFn defines the query method that based on unique indexes.
	IndexQueryFn[T any] func(ctx context.Context, repo conn.Repository[T]) (any, error)
	// PrimaryQueryCtxFn defines the query method that based on primary keys.
	PrimaryQueryFn[T any] func(ctx context.Context, repo conn.Repository[T], primary any) error

	DoFn[T any] func(repo conn.Repository[T]) error

	DoBatchFn[T any] func(repo conn.Repository[T], batchSize int) error

	ICacheKey[T any] interface {
		GetCacheKeys(data *T) []string
	}

	DoBatchOptions struct {
		BatchSize int
		//Tx        *gorm.DB
		Tx        *conn.ConnTx
		CleanKeys []string
	}

	DoBatchOption func(*DoBatchOptions)
)

func WithBatchSize(size int) DoBatchOption {
	return func(opts *DoBatchOptions) {
		opts.BatchSize = size
	}
}

// func WithTx(tx *gorm.DB) DoBatchOption {
func WithTx(tx *conn.ConnTx) DoBatchOption {
	return func(opts *DoBatchOptions) {
		opts.Tx = tx
	}
}

func WithCleanKeys(keys []string) DoBatchOption {
	return func(opts *DoBatchOptions) {
		opts.CleanKeys = keys
	}
}
