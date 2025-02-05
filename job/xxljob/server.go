package xxljob

import (
	"context"
	"fmt"

	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/core/proc"
	"github.com/shippomx/zard/core/service"
	ztrace "github.com/shippomx/zard/core/trace"
	"github.com/shippomx/zard/internal/health"
	"github.com/shippomx/zard/xxl-job-executor-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	probeNamePrefix = "xxljob"
)

type Server struct {
	service.ServiceConf
	conf          Conf
	executor      xxl.Executor
	healthManager health.Probe
}

type TaskRequest struct {
	JobID                 int64  `json:"jobId"`
	ExecutorHandler       string `json:"executorHandler"`
	ExecutorParams        string `json:"executorParams"`
	ExecutorBlockStrategy string `json:"executorBlockStrategy"`
	ExecutorTimeout       int64  `json:"executorTimeout"`
	LogID                 int64  `json:"logId"`
	LogDateTime           int64  `json:"logDateTime"`
	GlueType              string `json:"glueType"`
	GlueSource            string `json:"glueSource"`
	GlueUpdatetime        int64  `json:"glueUpdatetime"`
	BroadcastIndex        int64  `json:"broadcastIndex"`
	BroadcastTotal        int64  `json:"broadcastTotal"`
}

type TaskFunc func(ctx context.Context, param *TaskRequest) string

func NewServer(c Conf) *Server {
	logAdapter := &logxAdapter{}
	options := []xxl.Option{
		xxl.ServerAddr(c.ServerAddr),
		xxl.AccessToken(c.AccessToken),
		xxl.ExecutorPort(c.ExecutorPort),
		xxl.RegistryKey(c.Name),
		xxl.SetLogger(logAdapter),
	}

	if c.ExecutorIP != "" {
		options = append(options, xxl.ExecutorIp(c.ExecutorIP))
	}

	executor := xxl.NewExecutor(options...)
	executor.Use(TracingMiddleware)
	executor.Init()
	healthManager := health.NewHealthManager(fmt.Sprintf("%s-%s:%s", probeNamePrefix, c.ExecutorIP, c.ExecutorPort))
	health.AddProbe(healthManager)

	return &Server{
		ServiceConf:   c.ServiceConf,
		conf:          c,
		executor:      executor,
		healthManager: healthManager,
	}
}

func (s *Server) Start() {
	if err := s.SetUp(); err != nil {
		logx.Errorf("XXL-JOB setup error: %v", err)
		return
	}

	waitForCalled := proc.AddShutdownListener(func() {
		s.healthManager.MarkNotReady()
		s.executor.Stop()
	})

	defer func() {
		waitForCalled()
	}()

	s.healthManager.MarkReady()

	logx.Infof("Starting XXL-JOB server at %s:%s...", s.conf.ExecutorIP, s.conf.ExecutorPort)
	err := s.executor.Run()
	if err != nil {
		logx.Errorf("XXL-JOB server run error: %v", err)
	}
}

func (s *Server) Stop() {
	logx.Info("Stopping XXL-JOB server...")
	logx.Close()
}

func (s *Server) RegisterTask(pattern string, task TaskFunc) {
	s.executor.RegTask(pattern, s.adaptTask(task))
}

func (s *Server) adaptTask(task TaskFunc) xxl.TaskFunc {
	return func(cxt context.Context, param *xxl.RunReq) string {
		taskRequest := &TaskRequest{
			JobID:                 param.JobID,
			ExecutorHandler:       param.ExecutorHandler,
			ExecutorParams:        param.ExecutorParams,
			ExecutorBlockStrategy: param.ExecutorBlockStrategy,
			ExecutorTimeout:       param.ExecutorTimeout,
			LogID:                 param.LogID,
			LogDateTime:           param.LogDateTime,
			GlueType:              param.GlueType,
			GlueSource:            param.GlueSource,
			GlueUpdatetime:        param.GlueUpdatetime,
			BroadcastIndex:        param.BroadcastIndex,
			BroadcastTotal:        param.BroadcastTotal,
		}

		return task(cxt, taskRequest)
	}
}

// TracingMiddleware is a middleware that injects OpenTelemetry tracing into the XXL-JOB execution.
func TracingMiddleware(next xxl.TaskFunc) xxl.TaskFunc {
	return func(ctx context.Context, param *xxl.RunReq) string {
		tracer := otel.Tracer(ztrace.TraceName)
		propagator := otel.GetTextMapPropagator()

		ctx = propagator.Extract(ctx, propagation.MapCarrier{})
		spanName := param.ExecutorHandler
		spanCtx, span := tracer.Start(
			ctx,
			spanName,
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			oteltrace.WithAttributes(
				attribute.Int64("jobId", param.JobID),
				attribute.String("executorHandler", param.ExecutorHandler),
				attribute.Int64("logId", param.LogID),
			),
		)
		defer span.End()

		logx.WithContext(spanCtx).Debugf("Starting task: %s", param.ExecutorHandler)
		result := next(spanCtx, param)
		logx.WithContext(spanCtx).Debugf("Completed task: %s", param.ExecutorHandler)

		return result
	}
}
