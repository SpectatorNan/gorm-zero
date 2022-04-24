
func (m *default{{.upperStartCamelObject}}Model) Insert(ctx context.Context, data *{{.upperStartCamelObject}}) error {
	{{if .withCache}}{{.keys}}
    err := m.ExecCtx(ctx, func(conn *gorm.DB) *gorm.DB {
		return conn.Save(&data)
	}, {{.keyValues}}){{else}}err:=m.conn.WithContext(ctx).Save(&data).Error{{end}}
	return err
}
