FindOne(ctx context.Context, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) (*{{.upperStartCamelObject}}, error)
FindPageList(ctx context.Context, page *pagex.ListReq, orderBy []pagex.OrderBy,
	whereClause func(db *gorm.DB) *gorm.DB) ([]{{.upperStartCamelObject}}, int64, error)