import (
	"context"
	"database/sql"
	"time"

    {{if .decimal}}"github.com/shopspring/decimal"{{end}}
    gormcsql "github.com/shippomx/zard/gorm/gormc/sql"
	"gorm.io/gorm"
)

// avoid unused err.
var _ = time.Second
