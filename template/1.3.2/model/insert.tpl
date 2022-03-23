
func (m *default{{.upperStartCamelObject}}Model) Insert(data *{{.upperStartCamelObject}}) error {
	{{if .withCache}}{{if .containsIndexCache}}{{.keys}}
    err := m.Exec(func(conn *gorm.DB) *gorm.DB {
                       		return conn.Save(data)
                       	}, {{.keyValues}}){{else}}
	err:=m.ExecNoCache(func(conn *gorm.DB) *gorm.DB {
                                return conn.Save(data)
                              }){{end}}{{else}}
    err:=m.ExecNoCache(func(conn *gorm.DB) *gorm.DB {
                               		return conn.Save(data)
                               	}){{end}}
	return err
}
