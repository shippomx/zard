package gogen

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/shippomx/zard/tools/goctl/api/spec"
	"github.com/shippomx/zard/tools/goctl/config"
	"github.com/shippomx/zard/tools/goctl/util/format"
	"github.com/shippomx/zard/tools/goctl/util/pathx"
	"github.com/shippomx/zard/tools/goctl/vars"
)

//go:embed main.tpl
var mainTemplate string

func genMain(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	name := strings.ToLower(api.Service.Name)
	filename, err := format.FileNamingFormat(cfg.NamingFormat, name)
	if err != nil {
		return err
	}

	configName := filename
	if strings.HasSuffix(filename, "-api") {
		filename = strings.ReplaceAll(filename, "-api", "")
	}

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          "",
		filename:        filename + ".go",
		templateName:    "mainTemplate",
		category:        category,
		templateFile:    mainTemplateFile,
		builtinTemplate: mainTemplate,
		data: map[string]string{
			"importPackages": genMainImports(rootPkg),
			"serviceName":    configName,
		},
	})
}

func genMainImports(parentPkg string) string {
	var imports []string
	imports = append(imports, `"flag"`)
	imports = append(imports, `"fmt"`+"\n")
	imports = append(imports, fmt.Sprintf("\"%s/core/conf\"", vars.ProjectOpenSourceURL))
	imports = append(imports, fmt.Sprintf("\"%s/rest\"", vars.ProjectOpenSourceURL))
	imports = append(imports, fmt.Sprintf("\"%s/core/logx\"", vars.ProjectOpenSourceURL))
	imports = append(imports, "\n")
	imports = append(imports, fmt.Sprintf("\"%s\"", pathx.JoinPackages(parentPkg, configDir)))
	imports = append(imports, fmt.Sprintf("\"%s\"", pathx.JoinPackages(parentPkg, handlerDir)))
	imports = append(imports, fmt.Sprintf("\"%s\"", pathx.JoinPackages(parentPkg, contextDir)))
	return strings.Join(imports, "\n\t")
}
