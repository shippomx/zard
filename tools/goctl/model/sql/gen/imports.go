package gen

import (
	"github.com/shippomx/zard/tools/goctl/model/sql/template"
	"github.com/shippomx/zard/tools/goctl/util"
	"github.com/shippomx/zard/tools/goctl/util/pathx"
)

func genImports(table Table, withCache, timeImport bool, decimalImport bool) (string, error) {
	if withCache {
		text, err := pathx.LoadTemplate(category, importsTemplateFile, template.Imports)
		if err != nil {
			return "", err
		}

		buffer, err := util.With("import").Parse(text).Execute(map[string]any{
			"time":       timeImport,
			"containsPQ": table.ContainsPQ,
			"decimal":    decimalImport,
			"data":       table,
		})
		if err != nil {
			return "", err
		}

		return buffer.String(), nil
	}

	text, err := pathx.LoadTemplate(category, importsWithNoCacheTemplateFile, template.ImportsNoCache)
	if err != nil {
		return "", err
	}

	buffer, err := util.With("import").Parse(text).Execute(map[string]any{
		"time":       timeImport,
		"containsPQ": table.ContainsPQ,
		"decimal":    decimalImport,
		"data":       table,
	})
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
