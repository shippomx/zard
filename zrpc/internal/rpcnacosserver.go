package internal

import (
	"github.com/shippomx/zard/core/nacos"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
)

// NewRpcPubServer returns a Server.
func NewRpcNacosServer(nacosConf nacos.Config, name, listenOn string, middlewares ServerMiddlewaresConf,
	opts ...ServerOption) (Server, error) {
	registerNacos := func() error {
		// register service to nacos
		sc := []constant.ServerConfig{
			*constant.NewServerConfig(nacosConf.Ip, nacosConf.Port),
		}

		cc := &constant.ClientConfig{
			NamespaceId:         "public",
			Username:            nacosConf.Username,
			Password:            nacosConf.Password,
			TimeoutMs:           5000,
			NotLoadCacheAtStart: true,
			LogDir:              "/tmp/nacos/log",
			CacheDir:            "/tmp/nacos/cache",
			LogLevel:            "debug",
		}

		opts := nacos.NewNacosConfig(name, listenOn, sc, cc)
		err := nacos.RegisterService(opts)
		if err != nil {
			return err
		}

		return nil
	}

	server := keepAliveNacosServer{
		registerNacos: registerNacos,
		Server:        NewRpcServer(listenOn, middlewares, opts...),
	}

	return server, nil
}

type keepAliveNacosServer struct {
	registerNacos func() error
	Server
}

func (s keepAliveNacosServer) Start(fn RegisterFn) error {
	if err := s.registerNacos(); err != nil {
		return err
	}

	return s.Server.Start(fn)
}
