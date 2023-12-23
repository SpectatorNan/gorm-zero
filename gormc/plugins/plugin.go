package plugins

import "gorm.io/gorm"

func InitPlugins(db *gorm.DB) error {
	if err := db.Use(&TracingPlugin{}); err != nil {
		return err
	}
	return nil
}
