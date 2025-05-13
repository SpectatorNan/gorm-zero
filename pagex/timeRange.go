package pagex

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

// TimeRange
// example: db.Scopes(gormx.TimeRange(startTime, endTime))
func TimeRange(startTime, endTime *time.Time) func(db *gorm.DB) *gorm.DB {
	return TimeRangeByTable("", startTime, endTime)
}

// TimeRangeByTable
// example: db.Scopes(gormx.TimeRangeByTable("tableName", startTime, endTime))
func TimeRangeByTable(tableName string, startTime, endTime *time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if startTime != nil {
			sql := "created_at >= ?"
			if len(tableName) > 0 {
				sql = tableName + "." + sql
			}
			db = db.Where(sql, *startTime)
		}
		if endTime != nil {
			sql := "created_at <= ?"
			if len(tableName) > 0 {
				sql = tableName + "." + sql
			}
			db = db.Where(sql, *endTime)
		}
		return db
	}
}

func TimeRangeByColumn(column string, startTime, endTime *time.Time) func(db *gorm.DB) *gorm.DB {
	return TimeRangeByTableNameColumn("", column, startTime, endTime)
}

func TimeRangeByTableNameColumn(tableName string, column string, startTime, endTime *time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if startTime != nil {
			sql := fmt.Sprintf("%s >= ?", column)
			if len(tableName) > 0 {
				sql = tableName + "." + sql
			}
			db = db.Where(sql, *startTime)
		}
		if endTime != nil {
			sql := fmt.Sprintf("%s <= ?", column)
			if len(tableName) > 0 {
				sql = tableName + "." + sql
			}
			db = db.Where(sql, *endTime)
		}
		return db
	}
}

func Paginate(page *ListReq) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == nil {
			return db
		}
		return db.Offset(page.Offset()).Limit(page.Limit())
	}
}

func AvgRaw(column, alias string) string {
	return fmt.Sprintf("IFNULL(avg(%s),0) as %s", column, alias)
}

func CaseWhenNull(column, alias string, defaultValue interface{}) string {
	var fmtStr = "case WHEN avg( %s )  IS NULL THEN %v ELSE avg( %s ) END %s"
	if _, ok := defaultValue.(int); ok {
		fmtStr = "case WHEN avg( %s )  IS NULL THEN %d ELSE avg( %s ) END %s"
	} else if _, ok := defaultValue.(float64); ok {
		fmtStr = "case WHEN avg( %s )  IS NULL THEN %.2f ELSE avg( %s ) END %s"
	}

	return fmt.Sprintf(fmtStr, column, defaultValue, column, alias)
}
