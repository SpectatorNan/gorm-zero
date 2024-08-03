Delete(ctx context.Context, tx *gorm.DB, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) error
Transaction(ctx context.Context, fn func(db *gorm.DB) error) error