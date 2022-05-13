
func (m *default{{.upperStartCamelObject}}Model) Delete(ctx context.Context, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) error {
	{{if .withCache}}{{if .containsIndexCache}}data, err:=m.FindOne(ctx, {{.lowerStartCamelPrimaryKey}})
	if err!=nil{
		return err
	}

{{end}}	{{.keys}}
	 err {{if .containsIndexCache}}={{else}}:={{end}} m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Delete(&{{.upperStartCamelObject}}{}, {{.lowerStartCamelPrimaryKey}}).Error
	}, {{.keyValues}}){{else}} err:= m.conn.WithContext(ctx).Delete(&{{.upperStartCamelObject}}{}, {{.lowerStartCamelPrimaryKey}}).Error
	{{end}}
	return err
}