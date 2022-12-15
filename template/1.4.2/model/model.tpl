package {{.pkg}}
{{if .withCache}}
import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"gorm.io/gorm"
	{{ if or (.gormCreatedAt) (.gormUpdatedAt) }} "time" {{ end }}
)
{{else}}
import (
	"gorm.io/gorm"
	{{ if or (.gormCreatedAt) (.gormUpdatedAt) }} "time" {{ end }}
)
{{end}}
var _ {{.upperStartCamelObject}}Model = (*custom{{.upperStartCamelObject}}Model)(nil)

type (
	// {{.upperStartCamelObject}}Model is an interface to be customized, add more methods here,
	// and implement the added methods in custom{{.upperStartCamelObject}}Model.
	{{.upperStartCamelObject}}Model interface {
		{{.lowerStartCamelObject}}Model
		custom{{.upperStartCamelObject}}LogicModel
	}

	custom{{.upperStartCamelObject}}Model struct {
		*default{{.upperStartCamelObject}}Model
	}

	custom{{.upperStartCamelObject}}LogicModel interface {

    	}
)
{{ if or (.gormCreatedAt) (.gormUpdatedAt) }}
// BeforeCreate hook create time
func (s *{{.upperStartCamelObject}}) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	{{ if .gormCreatedAt }}s.CreatedAt = now{{ end }}
	{{ if .gormUpdatedAt}}s.UpdatedAt = now{{ end }}
	return nil
}
{{ end }}
{{ if .gormUpdatedAt}}
// BeforeUpdate hook update time
func (s *{{.upperStartCamelObject}}) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}
{{ end }}
// New{{.upperStartCamelObject}}Model returns a model for the database table.
func New{{.upperStartCamelObject}}Model(conn *gorm.DB{{if .withCache}}, c cache.CacheConf{{end}}) {{.upperStartCamelObject}}Model {
	return &custom{{.upperStartCamelObject}}Model{
		default{{.upperStartCamelObject}}Model: new{{.upperStartCamelObject}}Model(conn{{if .withCache}}, c{{end}}),
	}
}

func (m *default{{.upperStartCamelObject}}Model) customCacheKeys(data *{{.upperStartCamelObject}}) []string {
    if data == nil {
        return []string{}
    }
	return []string{}
}
