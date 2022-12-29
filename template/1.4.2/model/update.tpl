
func (m *default{{.upperStartCamelObject}}Model) Update(ctx context.Context, tx *gorm.DB, data *{{.upperStartCamelObject}}) error {
    {{if .withCache}}old, err := m.FindOne(ctx, data.{{.upperStartCamelPrimaryKey}})
    if err != nil && err != ErrNotFound {
        return err
    }
    err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
        db := conn
        if tx != nil {
            db = tx
        }
        return db.Save(data).Error
    }, m.getCacheKeys(old)...){{else}}db := m.conn
        if tx != nil {
            db = tx
        }
        err:= db.WithContext(ctx).Save(data).Error{{end}}
    return err
}
{{if .withCache}}
func (m *default{{.upperStartCamelObject}}Model) getCacheKeys(data *{{.upperStartCamelObject}}) []string {
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
