package xxljob

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/shippomx/zard/tools/goctl/internal/cobrax"
	"github.com/shippomx/zard/tools/goctl/util/ctx"
	"github.com/shippomx/zard/tools/goctl/util/pathx"
	"github.com/spf13/cobra"
)

var (
	varStringDir string

	Cmd = cobrax.NewCommand("xxljob", cobrax.WithRunE(generateXXLJob))
)

func init() {
	Cmd.PersistentFlags().StringVar(&varStringDir, "dir")
}

func generateXXLJob(*cobra.Command, []string) error {
	if err := pathx.MkdirIfNotExist(varStringDir); err != nil {
		return err
	}

	projectCtx, err := ctx.Prepare(varStringDir)
	if err != nil {
		return err
	}

	err = genEtc(varStringDir)
	if err != nil {
		return err
	}

	err = genConfig(varStringDir)
	if err != nil {
		return err
	}

	err = genServiceContext(varStringDir, projectCtx.Path)
	if err != nil {
		return err
	}

	err = genTaskRegistry(varStringDir, projectCtx.Path)
	if err != nil {
		return err
	}

	err = genEmailTask(varStringDir, projectCtx.Path)
	if err != nil {
		return err
	}

	return genMain(varStringDir, projectCtx.Path)
}

func genEtc(dir string) error {
	etcContent, err := pathx.LoadTemplate(category, etcTemplateFile, etcTemplate)
	if err != nil {
		return err
	}

	etcDir := filepath.Join(dir, "etc")
	if err := pathx.MkdirIfNotExist(etcDir); err != nil {
		return err
	}
	etcFile := filepath.Join(etcDir, "xxljob.yaml")
	return os.WriteFile(etcFile, []byte(etcContent), 0644)
}

func genConfig(dir string) error {
	configContent, err := pathx.LoadTemplate(category, configTemplateFile, configTemplate)
	if err != nil {
		return err
	}

	configDir := filepath.Join(dir, "internal", "config")
	if err := pathx.MkdirIfNotExist(configDir); err != nil {
		return err
	}
	configFile := filepath.Join(configDir, "config.go")
	return os.WriteFile(configFile, []byte(configContent), 0644)
}

func genServiceContext(dir, projectPackage string) error {
	content, err := pathx.LoadTemplate(category, svcTemplateFile, svcTemplate)
	if err != nil {
		return err
	}

	return genFile(dir, "internal/svc/servicecontext.go", content, projectPackage)
}

func genTaskRegistry(dir, projectPackage string) error {
	content, err := pathx.LoadTemplate(category, taskRegistryTemplateFile, taskRegistryTemplate)
	if err != nil {
		return err
	}

	return genFile(dir, "internal/tasks/taskregistry.go", content, projectPackage)
}

func genEmailTask(dir, projectPackage string) error {
	content, err := pathx.LoadTemplate(category, emailTaskTemplateFile, emailTaskTemplate)
	if err != nil {
		return err
	}

	return genFile(dir, "internal/tasks/email.go", content, projectPackage)
}

func genMain(dir, projectPackage string) error {
	content, err := pathx.LoadTemplate(category, mainTemplateFile, mainTemplate)
	if err != nil {
		return err
	}

	return genFile(dir, "main.go", content, projectPackage)
}

func genFile(dir, filename, content, projectPackage string) error {
	filePath := filepath.Join(dir, filename)
	if err := pathx.MkdirIfNotExist(filepath.Dir(filePath)); err != nil {
		return err
	}

	t, err := template.New("template").Parse(content)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]interface{}{
		"projectPackage": projectPackage,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, buf.Bytes(), 0644)
}
