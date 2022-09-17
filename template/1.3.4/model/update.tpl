
func (m *default{{.upperStartCamelObject}}Model) Update(ctx context.Context, data *{{.upperStartCamelObject}}) error {
	{{if .withCache}}
    err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Save(data).Error
	}, m.getCacheKeys(data)...){{else}}err:=m.conn.WithContext(ctx).Save(data).Error{{end}}
	return err
}
{{if .withCache}}
func (m *default{{.upperStartCamelObject}}Model) getCacheKeys(data *{{.upperStartCamelObject}}) []string {
	{{.keys}}
	cacheKeys := []string{
		{{.keyValues}},
	}
	cacheKeys = append(cacheKeys, m.customCacheKeys(data)...)
	return cacheKeys
}
{{end}}