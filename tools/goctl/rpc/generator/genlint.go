package generator

import _ "embed"

const (
	ciFileName = ".golangci.yml"
)

//go:embed golangci.tpl
var ciTemplate string

func genLint(ctx DirContext) error {
	return genFile(fileGenConfig{
		dir:             ctx.GetMain().Filename,
		subdir:          "",
		filename:        ciFileName,
		templateName:    "golangci-lint",
		category:        category,
		templateFile:    ciTemplateFile,
		builtinTemplate: ciTemplate,
		data: map[string]string{
			"serviceName": ctx.GetMain().Base,
		},
	})
}
