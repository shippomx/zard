package xxljob

import (
	_ "embed"
	"fmt"

	"github.com/shippomx/zard/tools/goctl/util/pathx"
)

const (
	category                 = "xxljob"
	etcTemplateFile          = "etc.tpl"
	configTemplateFile       = "config.tpl"
	svcTemplateFile          = "servicecontext.tpl"
	taskRegistryTemplateFile = "taskregistry.tpl"
	emailTaskTemplateFile    = "emailtask.tpl"
	mainTemplateFile         = "main.tpl"
)

//go:embed etc.tpl
var etcTemplate string

//go:embed config.tpl
var configTemplate string

//go:embed servicecontext.tpl
var svcTemplate string

//go:embed taskregistry.tpl
var taskRegistryTemplate string

//go:embed emailtask.tpl
var emailTaskTemplate string

//go:embed main.tpl
var mainTemplate string

var templates = map[string]string{
	etcTemplateFile:          etcTemplate,
	configTemplateFile:       configTemplate,
	svcTemplateFile:          svcTemplate,
	taskRegistryTemplateFile: taskRegistryTemplate,
	emailTaskTemplateFile:    emailTaskTemplate,
	mainTemplateFile:         mainTemplate,
}

// GenTemplates generates xxljob template files
func GenTemplates() error {
	return pathx.InitTemplates(category, templates)
}

// RevertTemplate reverts the given template file to the default value
func RevertTemplate(name string) error {
	content, ok := templates[name]
	if !ok {
		return fmt.Errorf("%s: no such file name", name)
	}
	return pathx.CreateTemplate(category, name, content)
}

// Clean deletes all template files
func Clean() error {
	return pathx.Clean(category)
}

// Update is used to update the template files, it will delete the existing old templates at first,
// and then create the latest template files
func Update() error {
	err := Clean()
	if err != nil {
		return err
	}

	return pathx.InitTemplates(category, templates)
}

// Category returns a const string value for xxljob template category
func Category() string {
	return category
}
