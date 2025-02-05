package new

import (
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/shippomx/zard/tools/goctl/api/gogen"
	conf "github.com/shippomx/zard/tools/goctl/config"
	"github.com/shippomx/zard/tools/goctl/util"
	"github.com/shippomx/zard/tools/goctl/util/pathx"
	"github.com/spf13/cobra"
)

//go:embed api.tpl
var apiTemplate string

const letterStep = 32

var (
	// VarStringHome describes the goctl home.
	VarStringHome string
	// VarStringRemote describes the remote git repository.
	VarStringRemote string
	// VarStringBranch describes the git branch.
	VarStringBranch string
	// VarStringStyle describes the style of output files.
	VarStringStyle string
	// VarHandlerStyle describes the style of handler name
	VarHandlerStyle string
)

// CreateServiceCommand fast create service
func CreateServiceCommand(_ *cobra.Command, args []string) error {
	dirName := args[0]

	if len(VarStringStyle) == 0 {
		VarStringStyle = conf.DefaultFormat
	}
	if strings.Contains(dirName, "-") {
		return errors.New("api new command service name not support strikethrough, because this will used by function name")
	}

	abs, err := filepath.Abs(dirName)
	if err != nil {
		return err
	}

	err = pathx.MkdirIfNotExist(abs)
	if err != nil {
		return err
	}

	dirName = filepath.Base(filepath.Clean(abs))
	filename := dirName + ".api"
	apiFilePath := filepath.Join(abs, filename)
	fp, err := os.Create(apiFilePath)
	if err != nil {
		return err
	}

	defer fp.Close()

	if len(VarStringRemote) > 0 {
		repo, _ := util.CloneIntoGitHome(VarStringRemote, VarStringBranch)
		if len(repo) > 0 {
			VarStringHome = repo
		}
	}

	if len(VarStringHome) > 0 {
		pathx.RegisterGoctlHome(VarStringHome)
	}

	text, err := pathx.LoadTemplate(category, apiTemplateFile, apiTemplate)
	if err != nil {
		return err
	}

	handlerName := dirName
	if VarHandlerStyle == conf.CamelHandlerFormat {
		handlerName = toCamel(dirName)
	}
	t := template.Must(template.New("template").Parse(text))
	if err := t.Execute(fp, map[string]string{
		"name":    dirName,
		"handler": strings.Title(handlerName),
	}); err != nil {
		return err
	}
	err = gogen.DoGenProject(apiFilePath, abs, VarStringStyle)
	return err
}

func toCamel(s string) string {
	res := strings.Builder{}
	words := strings.Split(s, "_")
	for _, w := range words {

		if len(w) > 0 && w[0] >= 'a' && w[0] <= 'z' {
			res.WriteString(fmt.Sprintf("%v%v", string(w[0]-letterStep), w[1:]))
		} else {
			res.WriteString(w)
		}
		// 单词首字母非小写字母的情况不需要处理
	}
	return res.String()
}
