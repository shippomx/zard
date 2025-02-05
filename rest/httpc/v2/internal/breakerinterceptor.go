package internal

import (
	"fmt"
	"net/http"

	"github.com/shippomx/zard/core/breaker"
)

func BreakerInterceptor(r *http.Request, ex ExtendInfo) (*http.Request, ResponseHandler) {
	var promise breaker.Promise

	if ex.Promise != nil {
		promise = ex.Promise
	}

	return r, func(resp *http.Response, err error) {
		if promise == nil {
			return
		}
		if err != nil {
			promise.Reject(err.Error())
		} else if !isOkResponse(resp.StatusCode) {
			promise.Reject(fmt.Sprintf("bad response code: %d", resp.StatusCode))
		} else {
			promise.Accept()
		}
	}
}

func WithBreaker(breaker breaker.Promise) breaker.Promise {
	return breaker
}
