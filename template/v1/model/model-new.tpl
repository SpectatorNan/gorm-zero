func ({{.upperStartCamelObject}}) TableName() string {
    return {{.table}}
}

func new{{.upperStartCamelObject}}Model(db *gorm.DB{{if .withCache}}, c cache.CacheConf{{end}}) *default{{.upperStartCamelObject}}Model {
	return &default{{.upperStartCamelObject}}Model{
		{{if .withCache}}CachedConn: gormc.NewConn(db, c){{else}}conn: db{{end}},
		table: {{.table}},
	}
}
