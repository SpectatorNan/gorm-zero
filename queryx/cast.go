package queryx

import (
	"fmt"
	"gorm.io/gorm"
)

func CastInt64ToChar(db *gorm.DB, column string, operator, val string) *gorm.DB {
	return db.Select(fmt.Sprintf("CAST(%s as char) %s ?", column, operator), val)
}
