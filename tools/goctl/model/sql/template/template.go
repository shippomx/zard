package template

import (
	"embed"
	"fmt"
	"os"

	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/tools/goctl/util"
)

//go:embed gorm-zero/* tpl/*
var f embed.FS

const ENV_GOCTL_SQL_TEMPLATE = "GOCTL_SQL_TEMPLATE"

func MustEmbed(file string) string {
	tplPath := "gorm-zero"
	switch val := os.Getenv(ENV_GOCTL_SQL_TEMPLATE); val {
	case "sqlx":
		tplPath = "tpl/"
	case "gorm-zero":
		fallthrough
	default:
		tplPath = "gorm-zero/"
	}
	content, err := f.ReadFile(tplPath + file)
	logx.Must(err)
	return string(content)
}

var (
	// Vars defines a template for var block in model
	Vars,

	// Types defines a template for types in model.
	Types,

	// Tag defines a tag template text
	Tag,

	// TableName defines a template that generate the tableName method.
	TableName,

	// New defines the template for creating model instance.
	New,

	// ModelCustom defines a template for extension
	ModelCustom,

	// ModelGen defines a template for model
	ModelGen,

	// Insert defines a template for insert code in model
	Insert,

	// InsertMethod defines an interface method template for insert code in model
	InsertMethod,

	// Update defines a template for generating update codes
	Update,

	// UpdateMethod defines an interface method template for generating update codes
	UpdateMethod,

	// Imports defines a import template for model in cache case
	Imports,

	// ImportsNoCache defines a import template for model in normal case
	ImportsNoCache,

	// FindOne defines find row by id.
	FindOne,

	// FindOneByField defines find row by field.
	FindOneByField,

	// FindOneByFieldExtraMethod defines find row by field with extras.
	FindOneByFieldExtraMethod,

	// FindOneMethod defines find row method.
	FindOneMethod,

	// FindOneByFieldMethod defines find row by field method.
	FindOneByFieldMethod,

	// Field defines a filed template for types
	Field,

	// Error defines an error template
	Error,

	// Delete defines a delete template
	Delete,

	// DeleteMethod defines a delete template for interface method
	DeleteMethod string
)

func Init() {
	Vars = MustEmbed("var.tpl")
	Types = MustEmbed("types.tpl")
	Tag = MustEmbed("tag.tpl")
	TableName = MustEmbed("table-name.tpl")
	New = MustEmbed("model-new.tpl")
	ModelCustom = MustEmbed("model.tpl")
	ModelGen = fmt.Sprintf(`%s

package {{.pkg}}
{{.imports}}
{{.vars}}
{{.types}}
{{.new}}
{{.delete}}
{{.find}}
{{.insert}}
{{.update}}
{{.extraMethod}}
{{.tableName}}
`, util.DoNotEditHead)
	Insert = MustEmbed("insert.tpl")
	InsertMethod = MustEmbed("interface-insert.tpl")
	Update = MustEmbed("update.tpl")
	UpdateMethod = MustEmbed("interface-update.tpl")
	Imports = MustEmbed("import.tpl")
	ImportsNoCache = MustEmbed("import-no-cache.tpl")
	FindOne = MustEmbed("find-one.tpl")
	FindOneByField = MustEmbed("find-one-by-field.tpl")
	FindOneByFieldExtraMethod = MustEmbed("find-one-by-field-extra-method.tpl")
	FindOneMethod = MustEmbed("interface-find-one.tpl")
	FindOneByFieldMethod = MustEmbed("interface-find-one-by-field.tpl")
	Field = MustEmbed("field.tpl")
	Error = MustEmbed("err.tpl")
	Delete = MustEmbed("delete.tpl")
	DeleteMethod = MustEmbed("interface-delete.tpl")
}
