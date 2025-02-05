package internal

import (
	"net/http"

	"github.com/shippomx/zard/core/breaker"
)

type ExtendInfo struct {
	breaker.Promise
	EnableMetricURL bool
}

type (
	ResponseHandler func(resp *http.Response, err error)
	Interceptor     func(r *http.Request, e ExtendInfo) (*http.Request, ResponseHandler)
)

func isOkResponse(code int) bool {
	return code < http.StatusBadRequest
}
