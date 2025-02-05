package generator

import _ "embed"

const (
	makefileName = "Makefile"
)

//go:embed makefile.tpl
var makefileTemplate string

func genMakefile(dir string) error {
	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          "",
		filename:        makefileName,
		templateName:    "makefile",
		category:        category,
		templateFile:    makefileTemplateFile,
		builtinTemplate: makefileTemplate,
	})
}
