
type (
	{{.lowerStartCamelObject}}Model interface{
		{{.method}}
	}

	default{{.upperStartCamelObject}}Model struct {
		{{if .withCache}}gormc.CachedConn{{else}}conn *gorm.DB{{end}}
		table string
	}

	{{.upperStartCamelObject}} struct {
		{{.fields}}
	}
)

var Q{{.upperStartCamelObject}} {{.upperStartCamelObject}}

func init() {
    gormcsql.InitField(&Q{{.upperStartCamelObject}})
}

