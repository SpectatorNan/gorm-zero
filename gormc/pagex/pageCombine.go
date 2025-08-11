package pagex

import (
	"context"

	"github.com/SpectatorNan/gorm-zero/gormc"
	"gorm.io/gorm"
)

var tableSortDesc = "descend"
var tableSortAsc = "ascend"

func Asc() string {
	return tableSortAsc
}
func Desc() string {
	return tableSortDesc
}

func SetTableSortAsc(key string) {
	tableSortAsc = key
}
func SetTableSortDesc(key string) {
	tableSortDesc = key
}

type GormcCacheConn interface {
	QueryNoCacheCtx(ctx context.Context, fn gormc.QueryCtxFn) error
	ExecNoCacheCtx(ctx context.Context, execCtx gormc.ExecCtxFn) error
}

// FindPageList
// fn first return db, second return countDb, if count sql need special handler (example: distinct on column), you can return countDb
// if countDb is nil, default count is first db
func FindPageList[T any](ctx context.Context, cc GormcCacheConn, page *ListReq, orderBy OrderBy,
	orderKeys map[string]string, fn func(conn *gorm.DB) (*gorm.DB, *gorm.DB)) ([]T, int64, error) {
	var res []T
	var count int64

	err := cc.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		db, countDb := fn(conn)
		if countDb != nil {
			db = countDb
		}
		return db.Count(&count).Error
	})
	if err != nil {
		return nil, 0, err
	}

	err = cc.QueryNoCacheCtx(ctx, func(conn *gorm.DB) error {
		db, _ := fn(conn)
		db = db.Scopes(Paginate(page))
		db = ApplyOrderBys(db, []OrderBy{orderBy}, orderKeys)
		return db.Find(&res).Error
	})
	if err != nil {
		return nil, 0, err
	}
	return res, count, nil
}

func FindPageListMultiOrderBy[T any](ctx context.Context, cc GormcCacheConn, page *ListReq, orderBys []OrderBy,
	orderKeys map[string]string, fn func(conn *gorm.DB) (*gorm.DB, *gorm.DB)) ([]T, int64, error) {
	var res []T
	var count int64

	err := cc.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		db, countDb := fn(conn)
		if countDb != nil {
			db = countDb
		}
		return db.Count(&count).Error
	})
	if err != nil {
		return nil, 0, err
	}

	err = cc.QueryNoCacheCtx(ctx, func(conn *gorm.DB) error {
		db, _ := fn(conn)
		db = db.Scopes(Paginate(page))
		db = ApplyOrderBys(db, orderBys, orderKeys)
		return db.Find(&res).Error
	})
	if err != nil {
		return nil, 0, err
	}
	return res, count, nil
}

func ApplyOrderBys(db *gorm.DB, orderBys []OrderBy, orderKeys map[string]string) *gorm.DB {
	for _, orderBy := range orderBys {
		if orderStr, ok := orderKeys[orderBy.OrderKey]; ok {
			if orderBy.Sort == tableSortDesc {
				db = db.Order(orderStr + " desc")
			} else {
				db = db.Order(orderStr + " asc")
			}
		}
	}
	return db
}

func FindList[T any](ctx context.Context, cc GormcCacheConn, page *ListReq, orderBys []OrderBy,
	orderKeys map[string]string, fn func(conn *gorm.DB) *gorm.DB) ([]T, error) {
	var res []T

	err := cc.QueryNoCacheCtx(ctx, func(conn *gorm.DB) error {
		db := fn(conn)
		db = db.Scopes(Paginate(page))
		db = ApplyOrderBys(db, orderBys, orderKeys)
		return db.Find(&res).Error
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// FindPageListWithCount
// fn first return db, second return countDb, if count sql need special handler (example: distinct on column), you can return countDb
// if countDb is nil, default count is first db
func FindPageListWithCount[T any](ctx context.Context, page *ListReq, orderBy OrderBy,
	orderKeys map[string]string, fn func() (*gorm.DB, *gorm.DB)) ([]T, int64, error) {
	var res []T
	var count int64

	db, countDb := fn()
	if countDb == nil {
		countDb = db
	}
	db = db.Scopes(Paginate(page))
	db = ApplyOrderBys(db, []OrderBy{orderBy}, orderKeys)

	err := db.Find(&res).Error
	if err != nil {
		return nil, 0, err
	}

	err = countDb.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	return res, count, nil
}
func FindPageListWithCountMultiOrderBy[T any](ctx context.Context, page *ListReq, orderBys []OrderBy,
	orderKeys map[string]string, fn func() (*gorm.DB, *gorm.DB)) ([]T, int64, error) {
	var res []T
	var count int64

	db, countDb := fn()
	if countDb == nil {
		countDb = db
	}
	err := countDb.WithContext(ctx).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	db = db.Scopes(Paginate(page))
	db = ApplyOrderBys(db, orderBys, orderKeys)
	err = db.WithContext(ctx).Find(&res).Error
	if err != nil {
		return nil, 0, err
	}
	return res, count, nil
}
