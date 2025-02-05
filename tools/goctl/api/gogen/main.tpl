package main

import (
	{{.importPackages}}
)

var configFile = flag.String("f", "etc/{{.serviceName}}.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config

	if err := c.InitNacosEnv(); err != nil {
		logx.Warnf("%s", err.Error())
		conf.MustLoad(*configFile, &c)
	} else {
		logx.Must(c.SyncNacosConf())
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(&c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
