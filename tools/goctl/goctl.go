package main

import (
	"github.com/shippomx/zard/core/load"
	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/tools/goctl/cmd"
)

func main() {
	logx.Disable()
	load.Disable()
	cmd.Execute()
}
