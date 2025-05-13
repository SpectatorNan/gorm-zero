import (
	"context"
	"errors"
	"fmt"
	{{if .time}}"time"{{end}}
	{{if .containsDbSql}}"database/sql"{{end}}
	"github.com/SpectatorNan/gorm-zero/gormc"
    "github.com/SpectatorNan/gorm-zero/gormc/batchx"
	"github.com/SpectatorNan/gorm-zero/gormc/pagex"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"gorm.io/gorm"

	{{if .third}}{{.third}}{{end}}
)
