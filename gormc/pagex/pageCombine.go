package pagex

import (
	"context"
	"github.com/SpectatorNan/gorm-zero/gormc"
	"gorm.io/gorm"
)

var tableSortDesc = "descend"
var tableSortAsc = "ascend"

func SetTableSortAsc(key string) {
	tableSortAsc = key
}
func SetTableSortDesc(key string) {
	tableSortDesc = key
}

type GormcCacheConn interface {
	QueryNoCacheCtx(ctx context.Context, v interface{}, fn gormc.QueryCtxFn) error
	ExecNoCacheCtx(ctx context.Context, execCtx gormc.ExecCtxFn) error
}

func FindPageList[T any](ctx context.Context, cc GormcCacheConn, page *ListReq, orderBy OrderBy, orderKeys map[string]string, fn func(conn *gorm.DB) *gorm.DB) ([]T, int64, error) {
	var res []T
	var count int64
	err := cc.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		return fn(conn).Count(&count).Error
	})
	if err != nil {
		return nil, 0, err
	}
	err = cc.QueryNoCacheCtx(ctx, &res, func(conn *gorm.DB, v interface{}) error {
		db := fn(conn).Scopes(Paginate(page))
		if orderStr, ok := orderKeys[orderBy.OrderKey]; ok {
			if orderBy.Sort == tableSortDesc {
				db = db.Order(orderStr + " desc")
			} else {
				db = db.Order(orderStr + " asc")
			}
		}
		return db.Find(v).Error
	})
	if err != nil {
		return nil, 0, err
	}
	return res, count, nil
}
