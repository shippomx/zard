import (
	"context"
	"fmt"
	"time"
	"database/sql"

    {{if .decimal}}"github.com/shopspring/decimal"{{end}}
	gormcsql "github.com/shippomx/zard/gorm/gormc/sql"
	"github.com/shippomx/zard/core/stores/cache"
	"gorm.io/gorm"
)

// avoid unused err.
var _ = time.Second
