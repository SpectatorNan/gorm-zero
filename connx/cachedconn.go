package connx

import (
	"context"
	"github.com/SpectatorNan/gorm-zero/v2/conn"
	"github.com/SpectatorNan/gorm-zero/v2/trace"
	"github.com/zeromicro/go-zero/core/mathx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"gorm.io/gen"
	"gorm.io/gorm"
	"time"
)

type (
	CachedConn[T any] struct {
		Conn               conn.Conn[T]
		cache              cache.Cache
		unstableExpiryTime mathx.Unstable
		cacheKeyProvider   ICacheKey[T] // 显式依赖缓存键提供者
		span               trace.Span
	}

	Option struct {
		cacheOpts []cache.Option
		genOpts   []gen.DOOption
	}
)

// missing query cache with expire set
func NewCachedConn[T any](db *gorm.DB, c cache.CacheConf, cacheKeyProvider ICacheKey[T], opts ...Option) CachedConn[T] {
	cOpts := make([]cache.Option, 0)
	gOpts := make([]gen.DOOption, 0)
	for _, opt := range opts {
		cOpts = append(cOpts, opt.cacheOpts...)
		gOpts = append(gOpts, opt.genOpts...)
	}
	cc := cache.New(c, singleFlights, stats, ErrNotFound, cOpts...)
	return NewCachedConnWithCache[T](db, cc, cacheKeyProvider, gOpts...)
}

func NewCachedConnWithCache[T any](db *gorm.DB, c cache.Cache, cacheKeyProvider ICacheKey[T], opts ...gen.DOOption) CachedConn[T] {
	condb := conn.NewConn[T](db, opts...)

	return CachedConn[T]{
		Conn:               condb,
		cache:              c,
		unstableExpiryTime: mathx.NewUnstable(expiryDeviation),
		cacheKeyProvider:   cacheKeyProvider, // 使用传入的缓存键提供者
		span:               trace.SpanFrom(traceName, spanName),
	}
}

// 添加 GetCacheKeys 方法到 TCachedConn
func (cc CachedConn[T]) GetCacheKeys(data *T) []string {
	return cc.cacheKeyProvider.GetCacheKeys(data)
}
func (cc CachedConn[T]) QueryCtx(ctx context.Context, key string, result any, query DoFn[T]) error {

	err := cc.span.With(ctx, "QueryCtx", func(ctx context.Context) error {
		return cc.cache.TakeCtx(ctx, &result, key, func(val any) error {
			err := query(cc.Conn.WithContext(ctx))
			if err != nil {
				return err
			}

			return nil
		})
	})
	if err != nil {
		return err
	}
	return nil
}

func (cc CachedConn[T]) QueryNoCache(ctx context.Context, query DoFn[T]) error {
	err := cc.span.With(ctx, "QueryNocCache", func(ctx context.Context) error {
		err := query(cc.Conn.WithContext(ctx))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (cc CachedConn[T]) DoCtx(ctx context.Context, doFn DoFn[T], keys ...string) error {
	return cc.span.With(ctx, "DoCtx", func(ctx context.Context) error {
		err := doFn(cc.Conn.WithContext(ctx))
		if err != nil {
			return err
		}
		err = cc.DelCache(ctx, keys...)
		if err != nil {
			return err
		}
		return nil
	})
}

// 通用批量处理
func (cc CachedConn[T]) doBatchProcess(
	ctx context.Context,
	datas []*T,
	processFn func(repo conn.Conn[T], data *T) error,
	opts ...DoBatchOption,
) error {
	options := &DoBatchOptions{
		BatchSize: defaultBatchSize,
		Tx:        nil,
	}
	for _, opt := range opts {
		opt(options)
	}
	return cc.span.With(ctx, "doBatchProcess", func(ctx context.Context) error {
		keys := cc.getCacheKeysByMultiData(datas)
		if len(keys) > 0 {
			opts = append(opts, WithCleanKeys(keys))
		}

		return cc.DoBatch(ctx, func(repo conn.Repository[T]) error {
			if options.Tx == nil {
				return cc.Conn.Transaction(func(tx *conn.ConnTx) error {
					txConn := conn.NewConnFromTx[T](tx)
					for i := 0; i < len(datas); i += options.BatchSize {
						end := i + options.BatchSize
						if end > len(datas) {
							end = len(datas)
						}
						batch := datas[i:end]
						for _, data := range batch {
							if err := processFn(txConn, data); err != nil {
								return err
							}
						}
					}
					return nil
				})
			} else {
				//txConn := conn.NewConn[T](options.Tx, cc.Conn.opts...)
				txConn := conn.NewConnFromTx[T](options.Tx) // 使用传入的事务
				for i := 0; i < len(datas); i += options.BatchSize {
					end := i + options.BatchSize
					if end > len(datas) {
						end = len(datas)
					}
					batch := datas[i:end]
					for _, data := range batch {
						if err := processFn(txConn, data); err != nil {
							return err
						}
					}
				}
				return nil
			}
		}, opts...)
	})
}

func (cc CachedConn[T]) doBatchProcessMulti(
	ctx context.Context,
	datas []*T,
	processFn func(repo conn.Conn[T], data []*T) error,
	opts ...DoBatchOption,
) error {
	options := &DoBatchOptions{
		BatchSize: defaultBatchSize,
		Tx:        nil,
	}
	for _, opt := range opts {
		opt(options)
	}
	return cc.span.With(ctx, "doBatchProcessMulti", func(ctx context.Context) error {
		keys := cc.getCacheKeysByMultiData(datas)
		if len(keys) > 0 {
			opts = append(opts, WithCleanKeys(keys))
		}

		return cc.DoBatch(ctx, func(repo conn.Repository[T]) error {
			if options.Tx == nil {
				return cc.Conn.Transaction(func(tx *conn.ConnTx) error {
					txConn := conn.NewConnFromTx[T](tx)
					for i := 0; i < len(datas); i += options.BatchSize {
						end := i + options.BatchSize
						if end > len(datas) {
							end = len(datas)
						}
						batch := datas[i:end]
						if err := processFn(txConn, batch); err != nil {
							return err
						}
					}
					return nil
				})
			} else {
				txConn := conn.NewConnFromTx[T](options.Tx)
				for i := 0; i < len(datas); i += options.BatchSize {
					end := i + options.BatchSize
					if end > len(datas) {
						end = len(datas)
					}
					batch := datas[i:end]
					if err := processFn(txConn, batch); err != nil {
						return err
					}
				}
				return nil
			}
		}, opts...)
	})
}

// 批量更新
func (cc CachedConn[T]) DoBatchUpdate(
	ctx context.Context,
	datas []*T,
	updateFn func(repo conn.Conn[T], data *T) error,
	opts ...DoBatchOption,
) error {
	return cc.doBatchProcess(ctx, datas, updateFn, opts...)
}

// 批量删除
func (cc CachedConn[T]) DoBatchDelete(
	ctx context.Context,
	datas []*T,
	deleteFn func(repo conn.Conn[T], data *T) error,
	opts ...DoBatchOption,
) error {
	return cc.doBatchProcess(ctx, datas, deleteFn, opts...)
}

// 批量创建（使用数据库原生批量插入）
func (cc CachedConn[T]) DoBatchCreate(ctx context.Context, datas []*T, opts ...DoBatchOption) error {
	options := &DoBatchOptions{
		BatchSize: defaultBatchSize,
		Tx:        nil,
	}
	for _, opt := range opts {
		opt(options)
	}
	return cc.span.With(ctx, "DoBatchCreate", func(ctx context.Context) error {
		keys := cc.getCacheKeysByMultiData(datas)
		if len(keys) > 0 {
			opts = append(opts, WithCleanKeys(keys))
		}

		return cc.DoBatch(ctx, func(repo conn.Repository[T]) error {
			return repo.CreateInBatches(datas, options.BatchSize)
		}, opts...)
	})
}

// 批量创建（使用自定义创建函数）
func (cc CachedConn[T]) DoBatchCreateCustom(
	ctx context.Context,
	datas []*T,
	createFn func(repo conn.Conn[T], data []*T) error,
	opts ...DoBatchOption,
) error {
	return cc.doBatchProcessMulti(ctx, datas, createFn, opts...)
}

func (cc CachedConn[T]) DoBatch(ctx context.Context, doFn DoFn[T], opts ...DoBatchOption) error {
	options := &DoBatchOptions{
		BatchSize: defaultBatchSize,
		Tx:        nil,
	}
	for _, opt := range opts {
		opt(options)
	}
	return cc.span.With(ctx, "DoBatch", func(ctx context.Context) error {
		conn := cc.Conn.WithContext(ctx)
		if options.Tx != nil {
			conn.ReplaceDB(options.Tx.UnderlyingDB())
		}

		err := doFn(conn)
		if err != nil {
			return err
		}
		if len(options.CleanKeys) > 0 {
			err = cc.DelCache(ctx, options.CleanKeys...)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// cache
func (cc CachedConn[T]) GetCache(ctx context.Context, key string, v interface{}) error {
	return cc.span.With(ctx, "GetCache", func(ctx context.Context) error {
		return cc.cache.GetCtx(ctx, key, v)
	})
}

func (cc CachedConn[T]) SetCache(ctx context.Context, key string, value interface{}) error {
	return cc.span.With(ctx, "SetCache", func(ctx context.Context) error {
		return cc.cache.SetCtx(ctx, key, value)
	})
}

func (cc CachedConn[T]) SetCacheWithExpire(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return cc.span.With(ctx, "SetCacheWithExpire", func(ctx context.Context) error {
		return cc.cache.SetWithExpireCtx(ctx, key, value, ttl)
	})
}

func (cc CachedConn[T]) DelCache(ctx context.Context, keys ...string) error {
	return cc.span.With(ctx, "DelCache", func(ctx context.Context) error {
		return cc.cache.DelCtx(ctx, keys...)
	})
}

func (cc CachedConn[T]) getCacheKeysByMultiData(data []*T) []string {
	if len(data) == 0 {
		return []string{}
	}
	var keys []string
	for _, v := range data {
		if v == nil {
			continue
		}
		keys = append(keys, cc.GetCacheKeys(v)...)
	}
	keys = cc.uniqueKeys(keys)
	return keys
}
func (cc CachedConn[T]) uniqueKeys(keys []string) []string {
	keySet := make(map[string]struct{})
	for _, key := range keys {
		keySet[key] = struct{}{}
	}

	uniKeys := make([]string, 0, len(keySet))
	for key := range keySet {
		uniKeys = append(uniKeys, key)
	}

	return uniKeys
}

// QueryRowIndexCtx unmarshals into v with given key using index-based caching strategy.
// This is the core method that implements the two-level caching strategy:
// 1. First level: cache index key -> primary key mapping
// 2. Second level: cache primary key -> actual data mapping
func (cc CachedConn[T]) QueryRowIndexCtx(ctx context.Context, v *T, key string,
	keyer func(primary any) string, indexQuery IndexQueryFn[T],
	primaryQuery PrimaryQueryFn[T]) error {

	return cc.span.With(ctx, "QueryRowIndex", func(ctx context.Context) error {
		var primaryKey any
		var found bool

		// First, try to get the primary key from index cache
		if err := cc.cache.TakeWithExpireCtx(ctx, &primaryKey, key,
			func(val any, expire time.Duration) (err error) {
				// Index cache miss, execute index query to get primary key
				primaryKey, err = indexQuery(ctx, cc.Conn.WithContext(ctx))
				if err != nil {
					return err
				}

				found = true
				// Cache the actual data with primary key and add safety gap
				return cc.cache.SetWithExpireCtx(ctx, keyer(primaryKey), v,
					expire+cacheSafeGapBetweenIndexAndPrimary)
			}); err != nil {
			return err
		}

		// If we found data during index query, we're done
		if found {
			return nil
		}

		// Otherwise, get data from primary key cache
		return cc.cache.TakeCtx(ctx, v, keyer(primaryKey), func(v any) error {
			return primaryQuery(ctx, cc.Conn.WithContext(ctx), primaryKey)
		})
	})
}
