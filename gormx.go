package gorm_zero

import (
	"database/sql"
	"github.com/tal-tech/go-zero/core/stores/cache"
	"github.com/tal-tech/go-zero/core/stores/redis"
	"github.com/tal-tech/go-zero/core/syncx"
	"gorm.io/gorm"
	"time"
)

// see doc/sql-cache.md
const cacheSafeGapBetweenIndexAndPrimary = time.Second * 5

var (
	// ErrNotFound is an alias of gorm.ErrRecordNotFound.
	ErrNotFound = gorm.ErrRecordNotFound

	// can't use one SingleFlight per conn, because multiple conns may share the same cache key.
	exclusiveCalls = syncx.NewSingleFlight()
	stats          = cache.NewStat("gorm")
)

type (
	CachedConn struct {
		db    gorm.DB
		cache cache.Cache
	}
)

// ~/Documents/go/pkg/mod/github.com/tal-tech/go-zero@v1.2.4/core/stores/sqlc/cachedsql.go

// NewNodeConn returns a CachedConn with a redis node cache.
func NewNodeConn(db gorm.DB, rds *redis.Redis, opts ...cache.Option) CachedConn {
	return CachedConn{
		db:    db,
		cache: cache.NewNode(rds, exclusiveCalls, stats, sql.ErrNoRows, opts...),
	}
}

// NewConn returns a CachedConn with a redis cluster cache.
func NewConn(db gorm.DB, c cache.CacheConf, opts ...cache.Option) CachedConn {
	return CachedConn{
		db:    db,
		cache: cache.New(c, exclusiveCalls, stats, sql.ErrNoRows, opts...),
	}
}