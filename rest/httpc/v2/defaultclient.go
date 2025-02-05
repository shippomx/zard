package httpc

import (
	"context"
	"net/http"

	"github.com/shippomx/zard/core/breaker"
	"github.com/shippomx/zard/core/proc"
	"github.com/shippomx/zard/rest/httpc/v2/admin"
	"github.com/shippomx/zard/rest/httpc/v2/internal"
)

var dc Client

// 用于设置包的defaultclient 的 DisableBreaker.
func DisableBreaker() {
	if hc, ok := dc.(*HTTPClient); ok {
		hc.brk = nil
	}
}

// 用于设置包的defaultclient 的 EnableMetricURL.
func EnableMetricURL() {
	if hc, ok := dc.(*HTTPClient); ok {
		hc.EnableMetricURL = true
	}
}

// 用于设置包的defaultclient 的 Middleware.
// 可变参数为了 扩展性.
func EnableDefaultMiddleware(metirc, logger, brk, trace bool, _ ...bool) {
	if hc, ok := dc.(*HTTPClient); ok {
		hc.interceptors = []internal.Interceptor{}
		if metirc {
			hc.interceptors = append(hc.interceptors, internal.MetricsInterceptor)
		}
		if logger {
			hc.interceptors = append(hc.interceptors, internal.LogInterceptor)
		}
		if brk {
			hc.interceptors = append(hc.interceptors, internal.BreakerInterceptor)
		}
		if trace {
			hc.interceptors = append(hc.interceptors, internal.TracingInterceptor)
		}
	}
}

func init() {
	p := breaker.NewBreaker(breaker.WithName("httpc"))
	dc = &HTTPClient{
		&http.Client{
			Transport: transport(),
		},
		p,
		false,
		[]internal.Interceptor{},
		func() {},
	}

	admin.RegisterClient(dc, func() {
		// re register if nacos no inited
		dc.(*HTTPClient).client.Transport = transport()
	})
	proc.AddShutdownListener(func() {
		dc.Close()
	})
}

func DoRequest(r *http.Request) (*http.Response, error) {
	return dc.DoRequest(r)
}

func Do(ctx context.Context, method, url string, data any) (*http.Response, error) {
	return dc.Do(ctx, method, url, data)
}

func Close() {
	dc.Close()
}

func transport() http.RoundTripper {
	t := http.Transport{}
	for schemaName, f := range admin.SchemeFunc {
		t.RegisterProtocol(schemaName, f)
	}
	return &t
}
