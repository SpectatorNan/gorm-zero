
type (
	{{.upperStartCamelObject}}Model interface{
		{{.method}}
	}

	default{{.upperStartCamelObject}}Model struct {
		{{if .withCache}}gormc.CachedConn{{else}}gormc.CachedConn{{end}}
	}

	{{.upperStartCamelObject}} struct {
		{{.fields}}
	}
)
