package main

import (
	"flag"

	"github.com/shippomx/zard/core/conf"
	"github.com/shippomx/zard/gateway"
)

var configFile = flag.String("f", "etc/gateway.yaml", "config file")

func main() {
	flag.Parse()

	var c gateway.GatewayConf
	conf.MustLoad(*configFile, &c)
	gw := gateway.MustNewServer(c)
	defer gw.Stop()
	gw.Start()
}
