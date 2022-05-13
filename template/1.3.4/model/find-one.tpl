
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
