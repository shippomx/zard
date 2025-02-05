package clientinterceptors

import (
	"context"
	"strconv"

	"github.com/shippomx/zard/core/metric"
	"github.com/shippomx/zard/core/timex"
	"github.com/shippomx/zard/core/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const clientNamespace = "rpc_client"

var (
	metricClientReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "rpc client requests duration(ms).",
		Labels:    []string{"gz_version", "method"},
		Buckets:   []float64{1, 2, 5, 10, 25, 50, 100, 250, 500, 1000, 2000, 5000},
	})

	metricClientReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "rpc client requests code count.",
		Labels:    []string{"gz_version", "method", "code"},
	})
)

// PrometheusInterceptor is an interceptor that reports to prometheus server.
func PrometheusInterceptor(ctx context.Context, method string, req, reply any,
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
) error {
	startTime := timex.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	metricClientReqDur.Observe(timex.Since(startTime).Milliseconds(), utils.BuildVersion, method)
	metricClientReqCodeTotal.Inc(utils.BuildVersion, method, strconv.Itoa(int(status.Code(err))))
	return err
}
