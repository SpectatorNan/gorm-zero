package queryx

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
)

func LikeByBigIntColumn(db *gorm.DB, column, val string) *gorm.DB {
	if len(column) > 0 {
		return db.Where(fmt.Sprintf("CAST(%s as char) like ?", column), "%"+val+"%")
	}
	return db
}

func OptionWhereString(db *gorm.DB, whereStr string, fuzz string) *gorm.DB {
	if len(fuzz) > 0 {
		if strings.Contains(whereStr, " like ") {
			return db.Where(whereStr, "%"+fuzz+"%")
		} else {
			return db.Where(whereStr, fuzz)
		}
	}
	return db
}
func OptionWhereBool(db *gorm.DB, whereStr string, b *bool) *gorm.DB {
	if b != nil {
		return db.Where(whereStr, b)
	}
	return db
}
func OptionWhereId(db *gorm.DB, whereStr string, id int64) *gorm.DB {
	if id > 0 {
		return db.Where(whereStr, id)
	}
	return db
}
func OptionWhereInt(db *gorm.DB, whereStr string, val int64) *gorm.DB {
	if val != 0 {
		return db.Where(whereStr, val)
	}
	return db
}
func OptionWhereInts(db *gorm.DB, whereStr string, val []int64) *gorm.DB {
	if len(val) > 0 {
		return db.Where(whereStr, val)
	}
	return db
}
