package generator

import (
	_ "embed"
	"path/filepath"

	conf "github.com/shippomx/zard/tools/goctl/config"
	"github.com/shippomx/zard/tools/goctl/rpc/parser"
	"github.com/shippomx/zard/tools/goctl/util"
	"github.com/shippomx/zard/tools/goctl/util/format"
	"github.com/shippomx/zard/tools/goctl/util/pathx"
)

//go:embed config.tpl
var configTemplate string

// GenConfig generates the configuration structure definition file of the rpc service,
// which contains the zrpc.RpcServerConf configuration item by default.
// You can specify the naming style of the target file name through config.Config. For details,
// see https://github.com/shippomx/zard/tree/master/tools/goctl/config/config.go
func (g *Generator) GenConfig(ctx DirContext, _ parser.Proto, cfg *conf.Config, c *ZRpcContext) error {
	dir := ctx.GetConfig()
	configFilename, err := format.FileNamingFormat(cfg.NamingFormat, "config")
	if err != nil {
		return err
	}

	fileName := filepath.Join(dir.Filename, configFilename+".go")
	if pathx.FileExists(fileName) {
		return nil
	}

	text, err := pathx.LoadTemplate(category, configTemplateFileFile, configTemplate)
	if err != nil {
		return err
	}

	return util.With("config").Parse(text).SaveTo(map[string]any{
		"gateway": c.IsGrpcGateway,
	}, fileName, false)
}
