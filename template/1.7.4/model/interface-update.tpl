Update(ctx context.Context, tx *gorm.DB, data *{{.upperStartCamelObject}}) error
BatchUpdate(ctx context.Context, tx *gorm.DB, olds, news []{{.upperStartCamelObject}}) error
BatchDelete(ctx context.Context, tx *gorm.DB, datas []{{.upperStartCamelObject}}) error
