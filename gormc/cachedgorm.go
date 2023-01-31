package gormc

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/mathx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// see doc/sql-cache.md
const cacheSafeGapBetweenIndexAndPrimary = time.Second * 5

// spanName is used to identify the span name for the SQL execution.
const spanName = "sql"

// make the expiry unstable to avoid lots of cached items expire at the same time
// make the unstable expiry to be [0.95, 1.05] * seconds
const expiryDeviation = 0.05

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
		db                 *gorm.DB
		cache              cache.Cache
		unstableExpiryTime mathx.Unstable
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
		db:                 db,
		cache:              c,
		unstableExpiryTime: mathx.NewUnstable(expiryDeviation),
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

// ExecNoCache runs exec with given sql statement, without affecting cache.
func (cc CachedConn) ExecNoCache(exec ExecCtxFn) error {
	return cc.ExecNoCacheCtx(context.Background(), exec)
}

// ExecNoCacheCtx runs exec with given sql statement, without affecting cache.
func (cc CachedConn) ExecNoCacheCtx(ctx context.Context, execCtx ExecCtxFn) (err error) {
	ctx, span := startSpan(ctx, "ExecNoCache")
	defer func() {
		endSpan(span, err)
	}()
	return execCtx(cc.db.WithContext(ctx))
}

// QueryRowIndex unmarshals into v with given key.
func (cc CachedConn) QueryRowIndex(v interface{}, key string, keyer func(primary interface{}) string,
	indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error {
	return cc.QueryRowIndexCtx(context.Background(), v, key, keyer, indexQuery, primaryQuery)
}

// QueryRowIndexCtx unmarshals into v with given key.
func (cc CachedConn) QueryRowIndexCtx(ctx context.Context, v interface{}, key string, keyer func(primary interface{}) string, indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) (err error) {
	ctx, span := startSpan(ctx, "QueryRowIndex")
	defer func() {
		endSpan(span, err)
	}()

	var primaryKey interface{}
	var found bool

	if err = cc.cache.TakeWithExpireCtx(ctx, &primaryKey, key, func(val interface{}, expire time.Duration) error {
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

func (cc CachedConn) QueryCtx(ctx context.Context, v interface{}, key string, query QueryCtxFn) (err error) {
	ctx, span := startSpan(ctx, "Query")
	defer func() {
		endSpan(span, err)
	}()
	return cc.cache.TakeCtx(ctx, v, key, func(v interface{}) error {
		return query(cc.db.WithContext(ctx), v)
	})
}

func (cc CachedConn) QueryNoCacheCtx(ctx context.Context, v interface{}, query QueryCtxFn) (err error) {
	ctx, span := startSpan(ctx, "QueryNoCache")
	defer func() {
		endSpan(span, err)
	}()
	return query(cc.db.WithContext(ctx), v)
}

// QueryWithExpireCtx unmarshals into v with given key, set expire duration and query func.
func (cc CachedConn) QueryWithExpireCtx(ctx context.Context, v interface{}, key string, expire time.Duration, query QueryCtxFn) (err error) {
	ctx, span := startSpan(ctx, "QueryWithExpire")
	defer func() {
		endSpan(span, err)
	}()
	err = query(cc.db.WithContext(ctx), v)
	if err != nil {
		return err
	}
	return cc.cache.SetWithExpireCtx(ctx, key, v, cc.aroundDuration(expire))
	//return cc.cache.TakeWithSetExpireCtx(ctx, v, key, expire, func(val interface{}) error {
	//	return query(cc.db.WithContext(ctx), v)
	//})
}
func (cc CachedConn) aroundDuration(duration time.Duration) time.Duration {
	return cc.unstableExpiryTime.AroundDuration(duration)
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

var sqlAttributeKey = attribute.Key("sql.method")

func startSpan(ctx context.Context, method string) (context.Context, oteltrace.Span) {
	tracer := otel.Tracer(trace.TraceName)
	start, span := tracer.Start(ctx,
		spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
	)
	span.SetAttributes(sqlAttributeKey.String(method))

	return start, span
}

func endSpan(span oteltrace.Span, err error) {
	defer span.End()

	if err == nil || err == ErrNotFound {
		span.SetStatus(codes.Ok, "")
		return
	}

	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}
