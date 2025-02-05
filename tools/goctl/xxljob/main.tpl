package main

import (
	"flag"
	"{{.projectPackage}}/internal/config"
	"{{.projectPackage}}/internal/svc"
	"{{.projectPackage}}/internal/tasks"

	"github.com/shippomx/zard/core/conf"
	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/job/xxljob"
)

var configFile = flag.String("f", "etc/xxljob.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	err := c.InitNacosEnv()
	if err != nil {
		logx.Warnf(err.Error())
		// 如果env不存在,则加载本地文件.
		conf.MustLoad(*configFile, &c)
	} else {
		logx.Must(c.MustNacosConf())
	}

	ctx := svc.NewServiceContext(c)

	server := xxljob.NewServer(c.Conf)
	defer server.Stop()
	tasks.RegisterTasks(server, ctx)
	server.Start()
}
