
func (m *default{{.upperStartCamelObject}}Model) Update(ctx context.Context, tx *gorm.DB, data *{{.upperStartCamelObject}}) error {
    {{if .withCache}}old, err := m.FindOne(ctx, data.{{.upperStartCamelPrimaryKey}})
    if err != nil && errors.Is(err, ErrNotFound) {
        return err
    }
    clearKeys := append(m.GetCacheKeys(old), m.GetCacheKeys(data)...)
    err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
        db := conn
        if tx != nil {
            db = tx
        }
        return db.Save(data).Error
    }, clearKeys...){{else}}db := m.conn
        if tx != nil {
            db = tx
        }
        err:= db.WithContext(ctx).Save(data).Error{{end}}
    return err
}
func (m *default{{.upperStartCamelObject}}Model) BatchUpdate(ctx context.Context, tx *gorm.DB, olds, news []{{.upperStartCamelObject}}) error {
    {{if .withCache}}clearData := make([]{{.upperStartCamelObject}}, 0, len(olds)+len(news))
    clearData = append(clearData, olds...)
    clearData = append(clearData, news...)
    err := batchx.BatchExecCtx(ctx, m, clearData, func(conn *gorm.DB) error {
    {{else}}err := m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
    {{end}}db := conn
        if tx != nil {
            db = tx
        }
        return db.Save(&news).Error
    })
    return err
}
