package gormc

import (
	"context"
	"database/sql"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"go.opentelemetry.io/otel"
	tracesdk "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"time"
)

// see doc/sql-cache.md
const cacheSafeGapBetweenIndexAndPrimary = time.Second * 5

// spanName is used to identify the span name for the SQL execution.
const spanName = "sql"

// TraceName represents the tracing name.
const TraceName = "gorm-zero"

var (
	// ErrNotFound is an alias of gorm.ErrRecordNotFound.
	ErrNotFound = gorm.ErrRecordNotFound

	// can't use one SingleFlight per conn, because multiple conns may share the same cache key.
	singleFlights = syncx.NewSingleFlight()
	stats         = cache.NewStat("gorm")
)

type (

	// ExecCtxFn defines the sql exec method.
	ExecCtxFn func(conn *gorm.DB) error
	// IndexQueryCtxFn defines the query method that based on unique indexes.
	IndexQueryCtxFn func(conn *gorm.DB, v interface{}) (interface{}, error)
	// PrimaryQueryCtxFn defines the query method that based on primary keys.
	PrimaryQueryCtxFn func(conn *gorm.DB, v, primary interface{}) error
	// QueryCtxFn defines the query method.
	QueryCtxFn func(conn *gorm.DB, v interface{}) error

	CachedConn struct {
		db    *gorm.DB
		cache cache.Cache
	}
)

// NewConn returns a CachedConn with a redis cluster cache.
func NewConn(db *gorm.DB, c cache.CacheConf, opts ...cache.Option) CachedConn {
	cc := cache.New(c, singleFlights, stats, ErrNotFound, opts...)
	return NewConnWithCache(db, cc)
}

// NewConnWithCache returns a CachedConn with a custom cache.
func NewConnWithCache(db *gorm.DB, c cache.Cache) CachedConn {
	return CachedConn{
		db:    db,
		cache: c,
	}
}

// NewNodeConn returns a CachedConn with a redis node cache.
func NewNodeConn(db *gorm.DB, rds *redis.Redis, opts ...cache.Option) CachedConn {
	cc := cache.NewNode(rds, singleFlights, stats, ErrNotFound, opts...)
	return NewConnWithCache(db, cc)
}

// DelCache deletes cache with keys.
func (cc CachedConn) DelCache(keys ...string) error {
	return cc.cache.DelCtx(context.Background(), keys...)
}

// DelCacheCtx deletes cache with keys.
func (cc CachedConn) DelCacheCtx(ctx context.Context, keys ...string) error {
	return cc.cache.DelCtx(ctx, keys...)
}

// GetCache unmarshals cache with given key into v.
func (cc CachedConn) GetCache(key string, v interface{}) error {
	return cc.cache.GetCtx(context.Background(), key, v)
}

// GetCacheCtx unmarshals cache with given key into v.
func (cc CachedConn) GetCacheCtx(ctx context.Context, key string, v interface{}) error {
	return cc.cache.GetCtx(ctx, key, v)
}

// Exec runs given exec on given keys, and returns execution result.
func (cc CachedConn) Exec(exec ExecCtxFn, keys ...string) error {
	return cc.ExecCtx(context.Background(), exec, keys...)
}

// ExecCtx runs given exec on given keys, and returns execution result.
func (cc CachedConn) ExecCtx(ctx context.Context, execCtx ExecCtxFn, keys ...string) error {
	err := execCtx(cc.db.WithContext(ctx))
	if err != nil {
		return err
	}
	if err := cc.DelCacheCtx(ctx, keys...); err != nil {
		return err
	}
	return nil
}

// ExecCtx runs given exec on given keys, and returns execution result.
func (cc CachedConn) ExecFnErrCtx(ctx context.Context, exec ExecCtxFn, keys ...string) error {
	err := exec(cc.db.WithContext(ctx))
	if err != nil {
		return err
	}
	if err := cc.DelCacheCtx(ctx, keys...); err != nil {
		return err
	}
	return nil
}

// ExecNoCache runs exec with given sql statement, without affecting cache.
func (cc CachedConn) ExecNoCache(exec ExecCtxFn) error {
	return cc.ExecNoCacheCtx(context.Background(), exec)
}

// ExecNoCacheCtx runs exec with given sql statement, without affecting cache.
func (cc CachedConn) ExecNoCacheCtx(ctx context.Context, execCtx ExecCtxFn) error {
	ctx, span := startSpan(ctx)
	defer span.End()
	return execCtx(cc.db.WithContext(ctx))
}

// QueryRowIndex unmarshals into v with given key.
func (cc CachedConn) QueryRowIndex(v interface{}, key string, keyer func(primary interface{}) string,
	indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error {
	return cc.QueryRowIndexCtx(context.Background(), v, key, keyer, indexQuery, primaryQuery)
}

// QueryRowIndexCtx unmarshals into v with given key.
func (cc CachedConn) QueryRowIndexCtx(ctx context.Context, v interface{}, key string, keyer func(primary interface{}) string, indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error {
	ctx, span := startSpan(ctx)
	defer span.End()

	var primaryKey interface{}
	var found bool

	if err := cc.cache.TakeWithExpireCtx(ctx, &primaryKey, key, func(val interface{}, expire time.Duration) (err error) {
		primaryKey, err = indexQuery(cc.db.WithContext(ctx), v)
		if err != nil {
			return err
		}
		found = true
		return cc.cache.SetWithExpireCtx(ctx, keyer(primaryKey), v, expire+cacheSafeGapBetweenIndexAndPrimary)
	}); err != nil {
		return err
	}
	if found {
		return nil
	}
	return cc.cache.TakeCtx(ctx, v, keyer(primaryKey), func(v interface{}) error {
		return primaryQuery(cc.db.WithContext(ctx), v, primaryKey)
	})
}

func (cc CachedConn) QueryCtx(ctx context.Context, v interface{}, key string, query QueryCtxFn) error {
	ctx, span := startSpan(ctx)
	defer span.End()
	return cc.cache.TakeCtx(ctx, v, key, func(v interface{}) error {
		return query(cc.db.WithContext(ctx), v)
	})
}

func (cc CachedConn) QueryNoCacheCtx(ctx context.Context, v interface{}, fn QueryCtxFn) error {
	ctx, span := startSpan(ctx)
	defer span.End()
	return fn(cc.db.WithContext(ctx), v)
}

// SetCache sets v into cache with given key.
func (cc CachedConn) SetCache(key string, v interface{}) error {
	return cc.cache.SetCtx(context.Background(), key, v)
}

// SetCacheCtx sets v into cache with given key.
func (cc CachedConn) SetCacheCtx(ctx context.Context, key string, val interface{}) error {
	return cc.cache.SetCtx(ctx, key, val)
}

// SetCacheWithExpireCtx sets v into cache with given key.
func (cc CachedConn) SetCacheWithExpireCtx(ctx context.Context, key string, val interface{}, expire time.Duration) error {
	return cc.cache.SetWithExpireCtx(ctx, key, val, expire)
}

// Transact runs given fn in transaction mode.
func (cc CachedConn) Transact(fn func(db *gorm.DB) error, opts ...*sql.TxOptions) error {
	return cc.TransactCtx(context.Background(), fn, opts...)
}

// TransactCtx runs given fn in transaction mode.
func (cc CachedConn) TransactCtx(ctx context.Context, fn func(db *gorm.DB) error, opts ...*sql.TxOptions) error {
	return cc.db.WithContext(ctx).Transaction(fn, opts...)
}

func startSpan(ctx context.Context) (context.Context, tracesdk.Span) {
	tracer := otel.GetTracerProvider().Tracer(TraceName)
	return tracer.Start(ctx, spanName)
}
