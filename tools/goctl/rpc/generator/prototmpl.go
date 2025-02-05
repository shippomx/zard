package generator

import (
	"embed"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/shippomx/zard/tools/goctl/util"
	"github.com/shippomx/zard/tools/goctl/util/pathx"
	"github.com/shippomx/zard/tools/goctl/util/stringx"
)

//go:embed rpc.tpl
var rpcTemplateText string

//go:embed validate.tpl
var rpcValidateTemplateText string

// ProtoTmpl returns a sample of a proto file
func ProtoTmpl(out string, validate, gateway bool) error {
	protoFilename := filepath.Base(out)
	serviceName := stringx.From(strings.TrimSuffix(protoFilename, filepath.Ext(protoFilename)))
	text, err := pathx.LoadTemplate(category, rpcTemplateFile, rpcTemplateText)
	if err != nil {
		return err
	}

	dir := filepath.Dir(out)
	err = pathx.MkdirIfNotExist(dir)
	if err != nil {
		return err
	}
	validateStr := ""
	if validate {
		validateStr = "true"
	}
	gatewayStr := ""
	if gateway {
		gatewayStr = "true"
	}
	err = util.With("t").Parse(text).SaveTo(map[string]string{
		"package":     serviceName.Untitle(),
		"serviceName": serviceName.Title(),
		"validate":    validateStr,
		"gateway":     gatewayStr,
	}, out, false)
	return err
}

func ValidateProtoTmpl(out string) error {
	dir := filepath.Dir(out)
	validateDir := filepath.Join(dir, "validate")
	validateFile := filepath.Join(validateDir, "validate.proto")
	err := pathx.MkdirIfNotExist(path.Join(dir, "validate"))
	if err != nil {
		return err
	}
	text, err := pathx.LoadTemplate(category, rpcValidateTemplateFile, rpcValidateTemplateText)
	if err != nil {
		return err
	}
	err = util.With("t").Parse(text).SaveTo(map[string]string{}, validateFile, false)
	return err
}

// https://pkg.go.dev/embed

//go:embed third_party/**/*.proto
//go:embed third_party/**/**/*.proto

var ThirdPartyTemplateFile embed.FS

func CopyThirdPartyFromEmbed(out string) error {
	dir := filepath.Dir(out)

	err := fs.WalkDir(ThirdPartyTemplateFile, "third_party", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".proto" {
			destPath := filepath.Join(dir, path)
			err := copyFile(destPath, path, ThirdPartyTemplateFile)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func copyFile(dst, src string, fs embed.FS) error {
	srcFile, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	//尝试创建dst的目录
	dir := filepath.Dir(dst)
	err = pathx.MkdirIfNotExist(dir)
	if err != nil {
		return err
	}
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
