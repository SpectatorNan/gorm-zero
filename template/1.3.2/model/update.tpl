
func (m *default{{.upperStartCamelObject}}Model) Update(data *{{.upperStartCamelObject}}) error {
	{{if .withCache}}{{.keys}}
    err := m.Exec(func(conn *gorm.DB) *gorm.DB {
    		return conn.Save(data)
    	}, {{.keyValues}}){{else}}
    err:=m.ExecNoCache(func(conn *gorm.DB) *gorm.DB {
             		return conn.Save(data)
             	}){{end}}
	return err
}
