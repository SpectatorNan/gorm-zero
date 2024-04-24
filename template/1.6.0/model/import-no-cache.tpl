import (
	"context"
	"github.com/SpectatorNan/gorm-zero/gormc"
	{{if .containsDbSql}}"database/sql"{{end}}
	{{if .time}}"time"{{end}}

	"gorm.io/gorm"
)
