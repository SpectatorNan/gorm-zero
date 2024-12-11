Delete(ctx context.Context, tx *gorm.DB, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) error
// deprecated. recommend add a transaction in service context instead of using this
Transaction(ctx context.Context, fn func(db *gorm.DB) error) error