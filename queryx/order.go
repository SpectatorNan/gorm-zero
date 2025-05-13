package queryx

import (
	pagex2 "github.com/SpectatorNan/gorm-zero/pagex"
	"gorm.io/gorm"
)

func Order(conn *gorm.DB, orderBy *pagex2.OrderBy, orderKeys map[string]string) *gorm.DB {
	if orderBy == nil {
		return conn
	}
	db := conn
	if orderStr, ok := orderKeys[orderBy.OrderKey]; ok {
		if orderBy.Sort == pagex2.Desc() {
			db = db.Order(orderStr + " desc")
		} else {
			db = db.Order(orderStr + " asc")
		}
	}
	return db
}
