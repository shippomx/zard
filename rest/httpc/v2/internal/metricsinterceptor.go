package internal

import (
	"net/http"
	"strconv"
	"time"

	"github.com/shippomx/zard/core/metric"
	"github.com/shippomx/zard/core/timex"
	"github.com/shippomx/zard/core/utils"
)

const clientNamespace = "http_client"

var (
	MetricClientReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "duration",
		Help:      "http client requests duration(ms).",
		Labels:    []string{"gz_version", "method", "url"},
		Buckets:   []float64{0.25, 0.5, 1, 2, 5, 10, 25, 50, 100, 250, 500, 1000, 2000, 5000, 10000, 15000},
	})

	MetricClientReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "total",
		Help:      "http client requests code count.",
		Labels:    []string{"gz_version", "method", "url", "code"},
	})
)

func MetricsInterceptor(r *http.Request, ex ExtendInfo) (*http.Request, ResponseHandler) {
	startTime := timex.Now()
	return r, func(resp *http.Response, err error) {
		var code int
		var path string

		// error or resp is nil, set code=500
		if err != nil || resp == nil {
			code = http.StatusInternalServerError
		} else {
			code = resp.StatusCode
		}

		if ex.EnableMetricURL {
			path = r.URL.Host + "/" + r.URL.Path
		}
		MetricClientReqDur.ObserveFloat(float64(timex.Since(startTime))/float64(time.Millisecond), utils.BuildVersion, r.Method, path)
		MetricClientReqCodeTotal.Inc(utils.BuildVersion, r.Method, path, strconv.Itoa(code))
	}
}
