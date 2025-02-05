package admin

import (
	"net/http"
)

var SchemeFunc map[string]http.RoundTripper

var (
	internalClient interface{}
	internalFunc   func()
)

func RegisterClient(i interface{}, fn func()) {
	internalClient = i
	internalFunc = fn
}

// 对于init 保证httpdefaultclient init 后可以通过此回调 重新替换transport初始化.
func AddService(fn func(client interface{}, fn func())) {
	fn(internalClient, internalFunc)
}
