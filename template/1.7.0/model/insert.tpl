{{if .withCache}}
func (m *default{{.upperStartCamelObject}}Model) GetCacheKeys(data *{{.upperStartCamelObject}}) []string {
    if data == nil {
        return []string{}
    }
    {{.keys}}
    cacheKeys := []string{
        {{.keyValues}},
    }
    cacheKeys = append(cacheKeys, m.customCacheKeys(data)...)
    return cacheKeys
}
{{end}}

func (m *default{{.upperStartCamelObject}}Model) Insert(ctx context.Context, tx *gorm.DB, data *{{.upperStartCamelObject}}) error {
	{{if .withCache}}
    err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
        if tx != nil {
            db = tx
        }
        return db.Save(&data).Error
	}, m.GetCacheKeys(data)...){{else}}db := m.conn
        if tx != nil {
            db = tx
        }
        err:= db.WithContext(ctx).Save(&data).Error{{end}}
	return err
}
func (m *default{{.upperStartCamelObject}}Model) BatchInsert(ctx context.Context, tx *gorm.DB, news []{{.upperStartCamelObject}}) error {
	{{if .withCache}}
    err := batchx.BatchExecCtx(ctx, m, news, func(conn *gorm.DB) error {
    {{else}}
    err := m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
    {{end}}db := conn
        if tx != nil {
            db = tx
        }
        return db.Create(&news).Error
	})
	return err
}
