package gogen

import (
	_ "embed"
)

const (
	ciFileName = ".golangci.yml"
)

//go:embed golangci.tpl
var ciTemplate string

func genLint(dir string, rootPkg string) error {
	return genFile(fileGenConfig{
		dir:             dir,
		rootpkg:         rootPkg,
		subdir:          "",
		filename:        ciFileName,
		templateName:    "golangci-lint",
		category:        category,
		templateFile:    ciTemplateFile,
		builtinTemplate: ciTemplate,
		data: map[string]string{
			"serviceName": rootPkg,
		},
	})
}
