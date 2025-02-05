package serverinterceptors

import (
	"context"
	"strconv"

	"github.com/shippomx/zard/core/metric"
	"github.com/shippomx/zard/core/timex"
	"github.com/shippomx/zard/core/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const serverNamespace = "rpc_server"

var (
	metricServerReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "rpc server requests duration(ms).",
		Labels:    []string{"gz_version", "method"},
		Buckets:   []float64{1, 2, 5, 10, 25, 50, 100, 250, 500, 1000, 2000, 5000},
	})

	metricServerReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "rpc server requests code count.",
		Labels:    []string{"gz_version", "method", "code"},
	})
)

// UnaryPrometheusInterceptor reports the statistics to the prometheus server.
func UnaryPrometheusInterceptor(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {
	startTime := timex.Now()
	resp, err := handler(ctx, req)
	metricServerReqDur.Observe(timex.Since(startTime).Milliseconds(), utils.BuildVersion, info.FullMethod)
	metricServerReqCodeTotal.Inc(utils.BuildVersion, info.FullMethod, strconv.Itoa(int(status.Code(err))))
	return resp, err
}
