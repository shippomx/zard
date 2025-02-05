package gateway

import (
	"github.com/shippomx/zard/rest"
	"github.com/shippomx/zard/zrpc"
)

type GatewayConf struct {
	// nolint:revive // for backward compatibility
	RpcServer      zrpc.RpcServerConf
	Rest           rest.RestConf
	DisableTraces  bool `json:",default=false"`
	DisableMetrics bool `json:",default=false"`
}
