
func (m *default{{.upperStartCamelObject}}Model) Delete(ctx context.Context, tx *gorm.DB, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) error {
	{{if .withCache}}data, err:=m.FindOne(ctx, {{.lowerStartCamelPrimaryKey}})
	if err!=nil{
        if errors.Is(err, ErrNotFound) {
                return nil
        }
		return err
	}
	 err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
        if tx != nil {
            db = tx
        }
        return db.Delete(&{{.upperStartCamelObject}}{}, {{.lowerStartCamelPrimaryKey}}).Error
	}, m.GetCacheKeys(data)...){{else}} db := m.conn
        if tx != nil {
            db = tx
        }
        err:= db.WithContext(ctx).Delete(&{{.upperStartCamelObject}}{}, {{.lowerStartCamelPrimaryKey}}).Error
	{{end}}
	return err
}

func (m *default{{.upperStartCamelObject}}Model) BatchDelete(ctx context.Context, tx *gorm.DB, datas []{{.upperStartCamelObject}}) error {
	{{if .withCache}}err := batchx.BatchExecCtx(ctx, m, datas, func(conn *gorm.DB) error {
	        db := conn
                if tx != nil {
                    db = tx
                }
                return db.Delete(&datas).Error
        	})
    {{else}}db := m.conn
                    if tx != nil {
                        db = tx
                    }
    err:= db.WithContext(ctx).Delete(&datas).Error
    {{end}}
	return err
}

// deprecated. recommend add a transaction in service context instead of using this
func (m *default{{.upperStartCamelObject}}Model) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
    {{if .withCache}}return m.TransactCtx(ctx, fn){{else}} return m.conn.WithContext(ctx).Transaction(fn){{end}}
}