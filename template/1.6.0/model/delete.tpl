
func (m *default{{.upperStartCamelObject}}Model) Delete(ctx context.Context, tx *gorm.DB, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) error {
	{{if .withCache}}data, err:=m.FindOne(ctx, {{.lowerStartCamelPrimaryKey}})
	if err!=nil{
        if err == ErrNotFound {
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
	}, m.getCacheKeys(data)...){{else}} db := m.conn
        if tx != nil {
            db = tx
        }
        err:= db.WithContext(ctx).Delete(&{{.upperStartCamelObject}}{}, {{.lowerStartCamelPrimaryKey}}).Error
	{{end}}
	return err
}

func (m *default{{.upperStartCamelObject}}Model) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
    {{if .withCache}}return m.TransactCtx(ctx, fn){{else}} return m.conn.WithContext(ctx).Transaction(fn){{end}}
}