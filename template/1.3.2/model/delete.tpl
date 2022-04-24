
func (m *default{{.upperStartCamelObject}}Model) Delete({{.lowerStartCamelPrimaryKey}} {{.dataType}}) error {
	{{if .withCache}}{{if .containsIndexCache}}data, err:=m.FindOne({{.lowerStartCamelPrimaryKey}})
	if err!=nil{
		return err
	}{{end}}

	{{.keys}}
     err {{if .containsIndexCache}}={{else}}:={{end}} m.Exec(func(conn *gorm.DB) *gorm.DB {
        return conn.Delete(&{{.upperStartCamelObject}}{}, {{.lowerStartCamelPrimaryKey}})
	}, {{.keyValues}}){{else}}
		err:=m.CachedConn.ExecNoCache(func(conn *gorm.DB) *gorm.DB {
		    return conn.Delete(&{{.upperStartCamelObject}}{}, {{.lowerStartCamelPrimaryKey}})
		}){{end}}
	return err
}