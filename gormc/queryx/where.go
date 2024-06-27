package queryx

import (
	"gorm.io/gorm"
	"strings"
)

func WhereOptionalInt64(db *gorm.DB, whereStr string, val *int64) *gorm.DB {
	if val != nil {
		return db.Where(whereStr, val)
	}
	return db
}

func WhereString(db *gorm.DB, whereStr string, fuzz string) *gorm.DB {
	if len(fuzz) > 0 {
		if strings.Contains(whereStr, " like ") {
			return db.Where(whereStr, "%"+fuzz+"%")
		} else {
			return db.Where(whereStr, fuzz)
		}
	}
	return db
}

func WhereOptionalBool(db *gorm.DB, whereStr string, b *bool) *gorm.DB {
	if b != nil {
		return db.Where(whereStr, b)
	}
	return db
}

func WhereUint64(db *gorm.DB, whereStr string, val uint64) *gorm.DB {
	if val != 0 {
		return db.Where(whereStr, val)
	}
	return db
}

func WhereInt64s(db *gorm.DB, whereStr string, val []int64) *gorm.DB {
	if len(val) > 1 {
		return db.Where(whereStr, val)
	} else if len(val) == 1 {
		return db.Where(whereStr, val[0])
	}
	return db
}

func WhereInt64(db *gorm.DB, whereStr string, val int64) *gorm.DB {
	if val != 0 {
		return db.Where(whereStr, val)
	}
	return db
}
