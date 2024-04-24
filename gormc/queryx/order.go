package queryx

import (
	"github.com/SpectatorNan/gorm-zero/gormc/pagex"
	"gorm.io/gorm"
)

func Order(conn *gorm.DB, orderBy *pagex.OrderBy, orderKeys map[string]string) *gorm.DB {
	if orderBy == nil {
		return conn
	}
	db := conn
	if orderStr, ok := orderKeys[orderBy.OrderKey]; ok {
		if orderBy.Sort == pagex.Desc() {
			db = db.Order(orderStr + " desc")
		} else {
			db = db.Order(orderStr + " asc")
		}
	}
	return db
}
