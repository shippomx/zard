package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/shippomx/zard/tools/goctl/rpc/generator"
	"github.com/shippomx/zard/tools/goctl/util"
	"github.com/shippomx/zard/tools/goctl/util/console"
	"github.com/shippomx/zard/tools/goctl/util/pathx"
	"github.com/spf13/cobra"
)

var (
	// VarStringOutput describes the output.
	VarStringOutput string
	// VarStringHome describes the goctl home.
	VarStringHome string
	// VarStringRemote describes the remote git repository.
	VarStringRemote string
	// VarStringBranch describes the git branch.
	VarStringBranch string
	// VarStringSliceGoOut describes the go output.
	VarStringSliceGoOut []string
	// VarStringSliceGoGRPCOut describes the grpc output.
	VarStringSliceGoGRPCOut []string
	// VarStringSlicePlugin describes the protoc plugin.
	VarStringSlicePlugin []string
	// VarStringSliceProtoPath describes the proto path.
	VarStringSliceProtoPath []string
	// VarStringSliceGoOpt describes the go options.
	VarStringSliceGoOpt []string
	// VarStringSliceGoGRPCOpt describes the grpc options.
	VarStringSliceGoGRPCOpt []string
	// VarStringStyle describes the style of output files.
	VarStringStyle string
	// VarStringZRPCOut describes the zRPC output.
	VarStringZRPCOut string
	// VarBoolIdea describes whether idea or not
	VarBoolIdea bool
	// VarBoolVerbose describes whether verbose.
	VarBoolVerbose bool
	// VarBoolMultiple describes whether support generating multiple rpc services or not.
	VarBoolMultiple bool
	// VarBoolClient describes whether to generate rpc client
	VarBoolClient bool
	// VarStringSlicevalidateOut describes the go output.
	VarStringSlicevalidateOut []string
	// VarBoolProtocValidate describes whether validate or not
	VarBoolProtocValidate bool
	// VarStringSliceGrpcGatewayOut describes the grpc gateway output.
	VarStringSliceGrpcGatewayOut []string
	// VarBoolGateway  describes whether generate gateway
	VarBoolGrpcGateway bool
	// VarBoolClientOnly describes whether generate client or not
	VarBoolClientOnly bool
)

// RPCNew is to generate rpc greet service, this greet service can speed
// up your understanding of the zrpc service structure
func RPCNew(_ *cobra.Command, args []string) error {
	rpcname := args[0]
	ext := filepath.Ext(rpcname)
	if len(ext) > 0 {
		return fmt.Errorf("unexpected ext: %s", ext)
	}
	style := VarStringStyle
	home := VarStringHome
	remote := VarStringRemote
	branch := VarStringBranch
	verbose := VarBoolVerbose
	validate := VarBoolProtocValidate
	gateway := VarBoolGrpcGateway
	if len(remote) > 0 {
		repo, _ := util.CloneIntoGitHome(remote, branch)
		if len(repo) > 0 {
			home = repo
		}
	}
	if len(home) > 0 {
		pathx.RegisterGoctlHome(home)
	}

	protoName := rpcname + ".proto"
	filename := filepath.Join(".", rpcname, protoName)
	src, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	err = generator.ProtoTmpl(src, validate, gateway)
	if err != nil {
		return err
	}
	validateArg := ""
	if validate {
		validateArg = fmt.Sprintf("--validate_out=lang=go:%s", filepath.Dir(src))
	}
	gatewayArg := ""
	if gateway {
		gatewayArg = fmt.Sprintf("--grpc-gateway_out=%s", filepath.Dir(src))
	}
	tpArgs := ""
	if gateway || validate {
		err = generator.CopyThirdPartyFromEmbed(src)
		if err != nil {
			return err
		}
		tpArgs = fmt.Sprintf("-I%s/third_party ", filepath.Dir(src))
	}
	var ctx generator.ZRpcContext
	ctx.Src = src
	ctx.GoOutput = filepath.Dir(src)
	ctx.GrpcOutput = filepath.Dir(src)
	ctx.IsGooglePlugin = true
	ctx.Output = filepath.Dir(src)
	ctx.ProtocCmd = fmt.Sprintf("protoc -I=%s %s %s --go_out=%s --go-grpc_out=%s %s %s", filepath.Dir(src), filepath.Base(src), tpArgs, filepath.Dir(src), filepath.Dir(src), validateArg, gatewayArg)
	ctx.IsGenClient = VarBoolClient
	ctx.IsGrpcGateway = VarBoolGrpcGateway
	grpcOptList := VarStringSliceGoGRPCOpt
	if len(grpcOptList) > 0 {
		ctx.ProtocCmd += " --go-grpc_opt=" + strings.Join(grpcOptList, ",")
	}

	goOptList := VarStringSliceGoOpt
	if len(goOptList) > 0 {
		ctx.ProtocCmd += " --go_opt=" + strings.Join(goOptList, ",")
	}

	g := generator.NewGenerator(style, verbose)
	return g.Generate(&ctx)
}

// RPCTemplate is the entry for generate rpc template
func RPCTemplate(latest bool) error {
	if !latest {
		console.Warning("deprecated: goctl rpc template -o is deprecated and will be removed in the future, use goctl rpc -o instead")
	}
	protoFile := VarStringOutput
	validate := VarBoolProtocValidate
	gateway := VarBoolGrpcGateway
	home := VarStringHome
	remote := VarStringRemote
	branch := VarStringBranch
	if len(remote) > 0 {
		repo, _ := util.CloneIntoGitHome(remote, branch)
		if len(repo) > 0 {
			home = repo
		}
	}
	if len(home) > 0 {
		pathx.RegisterGoctlHome(home)
	}

	if len(protoFile) == 0 {
		return errors.New("missing -o")
	}
	err := generator.ProtoTmpl(protoFile, validate, gateway)
	if err != nil {
		return err
	}
	if validate || gateway {
		err = generator.CopyThirdPartyFromEmbed(protoFile)
	}
	return err
}
