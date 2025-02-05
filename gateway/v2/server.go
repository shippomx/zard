package gateway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/shippomx/zard/core/logx"
	gozerometric "github.com/shippomx/zard/core/metric"
	"github.com/shippomx/zard/core/proc"
	"github.com/shippomx/zard/core/timex"
	ztrace "github.com/shippomx/zard/core/trace"
	"github.com/shippomx/zard/core/utils"
	"go.opentelemetry.io/otel"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type (
	Server struct {
		HttpServer *http.Server
		ctx        context.Context
	}
	registerFn func(ctx context.Context, gwmux *runtime.ServeMux, conn *grpc.ClientConn) error
)

const serverNamespace = "gateway_http_server"

var (
	metricServerReqDur = gozerometric.NewHistogramVec(&gozerometric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration",
		Help:      "gateway http server requests duration(ms).",
		Labels:    []string{"gz_version", "path", "method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 750, 1000},
	})

	metricServerReqTotal = gozerometric.NewCounterVec(&gozerometric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "total",
		Help:      "gateway http server requests count.",
		Labels:    []string{"gz_version", "path", "method"},
	})
)

func MustNewServer(c GatewayConf, reg registerFn, opts ...runtime.ServeMuxOption) *Server {
	options := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	ctx := context.Background()

	timeout := time.Duration(c.RpcServer.Timeout) * time.Millisecond
	if timeout != 0 {
		options = append(options, grpc.WithTimeout(timeout))
	}

	conn, err := grpc.DialContext(ctx, c.RpcServer.ListenOn, options...)
	if err != nil {
		logx.Must(fmt.Errorf("failed to dial grpc server: %w", err))
	}

	serveMuxOpts := append([]runtime.ServeMuxOption{}, opts...)

	if !c.DisableTraces {
		serveMuxOpts = append(serveMuxOpts, InjectTracingToHTTP(c.RpcServer.Name),
			InjectMetadataToRPC(),
			InjectRoutingErrorHandler(c.RpcServer.Name),
			runtime.WithErrorHandler(WrapDefaultHTTPErrorHandler),
		)
	}

	if !c.DisableMetrics {
		serveMuxOpts = append(serveMuxOpts, InjectMetricsToHTTP())
	}

	gwmux := runtime.NewServeMux(serveMuxOpts...)
	err = reg(ctx, gwmux, conn)
	if err != nil {
		logx.Must(fmt.Errorf("failed to register gateway handlers: %w", err))
	}

	hostPort := net.JoinHostPort(c.Rest.Host, strconv.Itoa(c.Rest.Port))
	gwserver := &http.Server{
		Addr:    hostPort,
		Handler: gwmux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	return &Server{
		HttpServer: gwserver,
	}
}

func (s *Server) Start() {
	proc.AddShutdownListener(
		func() {
			s.Stop()
		})
	s.HttpServer.ListenAndServe()
}

func (s *Server) Stop() {
	err := s.HttpServer.Shutdown(s.ctx)
	if err != nil {
		logx.Error(err)
	}
	logx.Close()
}

type StartServer interface {
	Start()
}

func Start(options ...StartServer) {
	wg := &sync.WaitGroup{}
	wg.Add(len(options))
	for _, option := range options {
		go func(option StartServer) {
			option.Start()
			wg.Done()
		}(option)
	}
	wg.Wait()
}

func WrapDefaultHTTPErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	propagator := otel.GetTextMapPropagator()
	spanCtx := propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))
	span := oteltrace.SpanFromContext(spanCtx)
	defer span.End()
	span.SetStatus(otelcodes.Error, err.Error())
	runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
}

func InjectTracingToHTTP(name string) runtime.ServeMuxOption {
	return runtime.WithMiddlewares(
		// tracing middleware
		func(hf runtime.HandlerFunc) runtime.HandlerFunc {
			tracer := otel.Tracer(ztrace.TraceName)
			propagator := otel.GetTextMapPropagator()
			return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
				spanName := r.URL.Path
				ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

				spanCtx, span := tracer.Start(
					ctx,
					spanName,
					oteltrace.WithSpanKind(oteltrace.SpanKindServer),
					oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(
						"grpc-gateway-"+name, spanName, r)...),
				)
				defer span.End()
				propagator.Inject(spanCtx, propagation.HeaderCarrier(r.Header))
				propagator.Inject(spanCtx, propagation.HeaderCarrier(w.Header()))
				hf(w, r, pathParams)
			}
		},
	)
}

func InjectMetricsToHTTP() runtime.ServeMuxOption {
	return runtime.WithMiddlewares(
		// metric middleware
		func(hf runtime.HandlerFunc) runtime.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
				startTime := timex.Now()
				defer func() {
					metricServerReqDur.Observe(timex.Since(startTime).Milliseconds(), utils.BuildVersion, r.URL.Path, r.Method)
					metricServerReqTotal.Inc(utils.BuildVersion, r.URL.Path, r.Method)
				}()
				hf(w, r, pathParams)
			}
		},
	)
}

func InjectMetadataToRPC() runtime.ServeMuxOption {
	return runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
		md := metadata.MD{}
		propagator := otel.GetTextMapPropagator()
		omd, oexists := metadata.FromOutgoingContext(ctx)
		if oexists {
			md = omd
		}
		imd, iexists := metadata.FromIncomingContext(ctx)
		if iexists {
			md = imd
		}
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))
		ztrace.Inject(ctx, propagator, &md)
		return md
	})
}

func InjectRoutingErrorHandler(name string) runtime.ServeMuxOption {
	return runtime.WithRoutingErrorHandler(func(ctx context.Context, sm *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, httpStatus int) {
		propagator := otel.GetTextMapPropagator()
		tracer := otel.Tracer(ztrace.TraceName)
		spanCtx := propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(
			spanCtx,
			r.URL.Path,
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(
				"grpc-gateway-"+name, r.URL.Path, r)...),
		)
		span.SetStatus(otelcodes.Error, "Unexpected routing erro")
		defer span.End()

		sterr := status.Error(codes.Internal, "Unexpected routing error")
		switch httpStatus {
		case http.StatusBadRequest:
			sterr = status.Error(codes.InvalidArgument, http.StatusText(httpStatus))
		case http.StatusMethodNotAllowed:
			sterr = status.Error(codes.Unimplemented, http.StatusText(httpStatus))
		case http.StatusNotFound:
			sterr = status.Error(codes.NotFound, http.StatusText(httpStatus))
		}
		WrapDefaultHTTPErrorHandler(ctx, sm, m, w, r, sterr)
	})
}
