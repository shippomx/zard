package internal

import (
	"net/http"

	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/core/timex"
	"go.opentelemetry.io/otel/propagation"
)

func LogInterceptor(r *http.Request, enableLogger bool) (*http.Request, ResponseHandler) {
	start := timex.Now()
	return r, func(resp *http.Response, err error) {
		duration := timex.Since(start)
		if err != nil {
			logger := logx.WithContext(r.Context()).WithDuration(duration)
			logger.Errorf("[HTTP] %s %s - %v", r.Method, r.URL, err)
			return
		}

		var tc propagation.TraceContext
		ctx := tc.Extract(r.Context(), propagation.HeaderCarrier(resp.Header))
		logger := logx.WithContext(ctx).WithDuration(duration)
		if isOkResponse(resp.StatusCode) {
			if enableLogger {
				logger.Infof("[HTTP] %d - %s %s", resp.StatusCode, r.Method, r.URL)
			}
		} else {
			logger.Errorf("[HTTP] %d - %s %s", resp.StatusCode, r.Method, r.URL)
		}
	}
}

func isOkResponse(code int) bool {
	return code < http.StatusBadRequest
}
