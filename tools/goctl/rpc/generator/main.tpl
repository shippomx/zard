package main

import ({{ if .gateway }}
	"context"
{{end}}
	{{.imports}}
	"flag"
	"fmt"

	"github.com/shippomx/zard/core/conf"
	"github.com/shippomx/zard/core/service"
	"github.com/shippomx/zard/zrpc"
	"github.com/shippomx/zard/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/{{.serviceName}}.yaml", "the config file")

func main() {
	flag.Parse()
	var c config.Config

	err := c.InitNacosEnv()
	if err != nil {
		logx.Warnf(err.Error())
		conf.MustLoad(*configFile, &c)
	} else {
		err = c.MustNacosConf()
		if err != nil {
			logx.Must(err)
		}
	}
	
	ctx := svc.NewServiceContext(&c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
{{range .serviceNames}}       {{.Pkg}}.Register{{.GRPCService}}Server(grpcServer, {{.ServerPkg}}.New{{.Service}}Server(ctx))
{{end}}
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop(){{ if .gateway }}
	gatewayConf := gateway.GatewayConf{
		RpcServer: c.RpcServerConf,
		Rest:      c.RestConf,
	}

	gw := gateway.MustNewServer(gatewayConf, func(ctx context.Context, gwmux *runtime.ServeMux, conn *grpc.ClientConn) error {
		{{range .serviceNames}}		if err = {{.Pkg}}.Register{{.GRPCService}}Handler(ctx, gwmux, conn);err != nil {
			logx.Must(err)
		}{{end}}
		return nil
	})
	defer gw.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	fmt.Printf("Starting grpc gateway at %s:%d...\n", c.RestConf.Host, c.RestConf.Port)
	gateway.Start(gw, s)
{{else}}
	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
{{end}}}
