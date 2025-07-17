FindOne(ctx context.Context, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) (*{{.upperStartCamelObject}}, error)
FindPageList(ctx context.Context, page *pagex.ListReq, orderBys []pagex.OrderBy,
	orderKeys map[string]string, whereClause func(db *gorm.DB) *gorm.DB) ([]{{.upperStartCamelObject}}, int64, error)