package lint

import (
	_ "embed"
	"github.com/spf13/cobra"
)

const (
	category             = "lint"
	ciFileName           = ".golangci.yml"
	makefileName         = "Makefile"
	ciTemplateFile       = "golangci.tpl"
	makefileTemplateFile = "makefile.tpl"
)

//go:embed golangci.tpl
var ciTemplate string

//go:embed makefile.tpl
var makefileTemplate string

var VarStringDir string

func lint(_ *cobra.Command, _ []string) error {
	dir := VarStringDir
	if err := genLint(dir); err != nil {
		return err
	}

	if err := genMakefile(dir); err != nil {
		return err
	}
	return nil
}

func genLint(dir string) error {
	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          "",
		filename:        ciFileName,
		templateName:    "golangci-lint",
		category:        category,
		templateFile:    ciTemplateFile,
		builtinTemplate: ciTemplate,
	})
}

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
