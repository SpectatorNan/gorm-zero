
func (m *default{{.upperStartCamelObject}}Model) FindOne(ctx context.Context, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) (*{{.upperStartCamelObject}}, error) {
	{{if .withCache}}{{.cacheKey}}
	var resp {{.upperStartCamelObject}}
	err := m.QueryCtx(ctx, &resp, {{.cacheKeyVariable}}, func(conn *gorm.DB, v interface{}) error {
    		return conn.Model(&{{.upperStartCamelObject}}{}).Where("{{.originalPrimaryKey}} = ?", {{.lowerStartCamelPrimaryKey}}).First(&resp).Error
    	})
	switch err {
	case nil:
		return &resp, nil
	case gormc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}{{else}}var resp {{.upperStartCamelObject}}
	err := m.conn.WithContext(ctx).Model(&{{.upperStartCamelObject}}{}).Where("{{.originalPrimaryKey}} = ?", {{.lowerStartCamelPrimaryKey}}).Take(&resp).Error
	switch err {
	case nil:
		return &resp, nil
	case gormc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}{{end}}
}
func (m *default{{.upperStartCamelObject}}Model) FindPageList(ctx context.Context, page *pagex.ListReq, orderBy pagex.OrderBy,
	orderKeys map[string]string, whereClause func(db *gorm.DB) *gorm.DB) ([]{{.upperStartCamelObject}}, int64, error) {
	{{if .withCache}}formatDB := func(conn *gorm.DB) (*gorm.DB, *gorm.DB) {
    		db := conn.Model(&{{.upperStartCamelObject}}{})
    		if whereClause != nil {
    			db = whereClause(db)
    		}
    		return db, nil
    	}
    	res, total, err := pagex.FindPageList[{{.upperStartCamelObject}}](ctx, m, page, orderBy, orderKeys, formatDB)
    	return res, total, err{{else}}conn := m.conn
                                      	formatDB := func() (*gorm.DB, *gorm.DB) {
                                      		db := conn.Model(&Users{})
                                      		if whereClause != nil {
                                      			db = whereClause(db)
                                      		}
                                      		return db, nil
                                      	}

                                      	res, total, err := pagex.FindPageListWithCount[Users](ctx, page, orderBy, orderKeys, formatDB)
                                      	return res, total, err{{end}}
}