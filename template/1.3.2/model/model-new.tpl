
func New{{.upperStartCamelObject}}Model(conn *gorm.DB{{if .withCache}}, c cache.CacheConf{{end}}) {{.upperStartCamelObject}}Model {
	return &default{{.upperStartCamelObject}}Model{
		CachedConn: gormc.NewConn(conn, c),
	}
}
