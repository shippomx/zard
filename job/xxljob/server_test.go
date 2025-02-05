package xxljob

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/core/service"
	"github.com/shippomx/zard/xxl-job-executor-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestNewServer(t *testing.T) {
	conf := Conf{
		ServiceConf:  service.ServiceConf{Name: "test-service"},
		ServerAddr:   "http://localhost:8080",
		AccessToken:  "test-token",
		ExecutorIP:   "127.0.0.1",
		ExecutorPort: "9999",
	}

	server := NewServer(conf)

	assert.NotNil(t, server)
	assert.Equal(t, conf, server.conf)
	assert.NotNil(t, server.executor)
}

func TestServer_Start(t *testing.T) {
	mockExecutor := new(MockExecutor)

	server := &Server{
		conf:          Conf{ExecutorIP: "127.0.0.1", ExecutorPort: "9999"},
		executor:      mockExecutor,
		healthManager: mockProbe{},
	}

	// 创建一个通道用于同步
	done := make(chan bool)
	mockExecutor.On("Run").Return(nil).Run(func(_ mock.Arguments) {
		done <- true
	})

	go server.Start()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for Run to be called")
	}

	// 验证Run方法被调用
	mockExecutor.AssertCalled(t, "Run")

	server.Stop()
}

func TestServer_Stop(t *testing.T) {
	_ = logx.SetUp(logx.LogConf{Mode: "console", Encoding: "plain"})

	mockExecutor := &MockExecutor{}
	server := &Server{
		executor: mockExecutor,
	}

	// Create a custom writer that implements logx.Writer
	logOutput := &TestLogWriter{}
	logx.SetWriter(logOutput)

	server.Stop()

	assert.Contains(t, logOutput.String(), "Stopping XXL-JOB server...")
}

func TestServer_RegisterTask(t *testing.T) {
	mockExecutor := new(MockExecutor)
	server := &Server{
		executor: mockExecutor,
	}
	testTask := func(ctx context.Context, param *TaskRequest) string {
		_ = ctx
		_ = param
		return "Task executed"
	}

	mockExecutor.On("RegTask", "testPattern", mock.AnythingOfType("xxl.TaskFunc")).Return()

	server.RegisterTask("testPattern", testTask)

	mockExecutor.AssertExpectations(t)
	calls := mockExecutor.Calls
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, "testPattern", calls[0].Arguments[0])
	assert.IsType(t, (xxl.TaskFunc)(nil), calls[0].Arguments[1])
}

func TestServer_AdaptTask(t *testing.T) {
	server := &Server{}

	originalTask := func(ctx context.Context, param *TaskRequest) string {
		_ = ctx
		result := map[string]interface{}{
			"success": true,
			"message": "Task executed",
			"data": map[string]interface{}{
				"jobId":           param.JobID,
				"executorHandler": param.ExecutorHandler,
			},
		}
		jsonResult, _ := json.Marshal(result)
		return string(jsonResult)
	}

	adaptedTask := server.adaptTask(originalTask)

	testRunReq := &xxl.RunReq{
		JobID:           123,
		ExecutorHandler: "testHandler",
		ExecutorParams:  "testParams",
	}

	result := adaptedTask(context.Background(), testRunReq)

	var resultMap map[string]interface{}
	err := json.Unmarshal([]byte(result), &resultMap)
	assert.NoError(t, err)

	assert.True(t, resultMap["success"].(bool))
	assert.Equal(t, "Task executed", resultMap["message"])
	assert.Equal(t, float64(123), resultMap["data"].(map[string]interface{})["jobId"])
	assert.Equal(t, "testHandler", resultMap["data"].(map[string]interface{})["executorHandler"])
}

func TestTracingMiddleware(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	mockTask := func(_ context.Context, _ *xxl.RunReq) string {
		return "Task completed"
	}

	testReq := &xxl.RunReq{
		JobID:           123,
		ExecutorHandler: "testHandler",
		LogID:           456,
	}
	wrappedTask := TracingMiddleware(mockTask)

	result := wrappedTask(context.Background(), testReq)
	assert.Equal(t, "Task completed", result)

	spans := sr.Ended()
	assert.Equal(t, 1, len(spans))

	span := spans[0]
	assert.Equal(t, "testHandler", span.Name())

	attrs := span.Attributes()
	assert.Contains(t, attrs, attribute.Int64("jobId", 123))
	assert.Contains(t, attrs, attribute.String("executorHandler", "testHandler"))
	assert.Contains(t, attrs, attribute.Int64("logId", 456))
}

func TestXXLJobServer_ExecutorMethods(t *testing.T) {
	mockExecutor := &MockExecutor{}

	server := &Server{
		executor: mockExecutor,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	server.executor.RunTask(w, r)
	assert.True(t, mockExecutor.RunTaskCalled)

	server.executor.KillTask(w, r)
	assert.True(t, mockExecutor.KillTaskCalled)

	server.executor.TaskLog(w, r)
	assert.True(t, mockExecutor.TaskLogCalled)

	server.executor.Beat(w, r)
	assert.True(t, mockExecutor.BeatCalled)

	server.executor.IdleBeat(w, r)
	assert.True(t, mockExecutor.IdleBeatCalled)
}

// MockExecutor is a mock implementation of xxl.Executor for testing.
type MockExecutor struct {
	InitCalled     bool
	RunCalled      bool
	StopCalled     bool
	RunTaskCalled  bool
	KillTaskCalled bool
	TaskLogCalled  bool
	BeatCalled     bool
	IdleBeatCalled bool
	mock.Mock
}

func (m *MockExecutor) Init(_ ...xxl.Option) { m.InitCalled = true }
func (m *MockExecutor) Run() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockExecutor) Stop() { m.StopCalled = true }
func (m *MockExecutor) RegTask(pattern string, task xxl.TaskFunc) {
	m.Called(pattern, task)
}
func (m *MockExecutor) RunTask(_ http.ResponseWriter, _ *http.Request)  { m.RunTaskCalled = true }
func (m *MockExecutor) KillTask(_ http.ResponseWriter, _ *http.Request) { m.KillTaskCalled = true }
func (m *MockExecutor) TaskLog(_ http.ResponseWriter, _ *http.Request)  { m.TaskLogCalled = true }
func (m *MockExecutor) Beat(_ http.ResponseWriter, _ *http.Request)     { m.BeatCalled = true }
func (m *MockExecutor) IdleBeat(_ http.ResponseWriter, _ *http.Request) { m.IdleBeatCalled = true }
func (m *MockExecutor) LogHandler(_ xxl.LogHandler)                     {}
func (m *MockExecutor) Use(_ ...xxl.Middleware)                         {}

type TestLogWriter struct {
	content string
}

func (w *TestLogWriter) Alert(v interface{}) {
	w.content += "ALERT: " + v.(string) + "\n"
}

func (w *TestLogWriter) Close() error {
	return nil
}

func (w *TestLogWriter) Error(v interface{}, _ ...logx.LogField) {
	w.content += "ERROR: " + v.(string) + "\n"
}

func (w *TestLogWriter) Warn(v interface{}, _ ...logx.LogField) {
	w.content += "WARN: " + v.(string) + "\n"
}

func (w *TestLogWriter) Info(v interface{}, _ ...logx.LogField) {
	w.content += "INFO: " + v.(string) + "\n"
}

func (w *TestLogWriter) Debug(v interface{}, _ ...logx.LogField) {
	w.content += "DEBUG: " + v.(string) + "\n"
}

func (w *TestLogWriter) Severe(v interface{}) {
	w.content += "SEVERE: " + v.(string) + "\n"
}

func (w *TestLogWriter) Slow(v interface{}, _ ...logx.LogField) {
	w.content += "SLOW: " + v.(string) + "\n"
}

func (w *TestLogWriter) Stack(v interface{}) {
	w.content += "STACK: " + v.(string) + "\n"
}

func (w *TestLogWriter) Stat(v interface{}, _ ...logx.LogField) {
	w.content += "STAT: " + v.(string) + "\n"
}

func (w *TestLogWriter) Write(data []byte) (n int, err error) {
	w.content += string(data)
	return len(data), nil
}

func (w *TestLogWriter) String() string {
	return w.content
}

type mockProbe struct{}

func (m mockProbe) MarkReady() {}

func (m mockProbe) MarkNotReady() {}

func (m mockProbe) IsReady() bool { return false }

func (m mockProbe) Name() string { return "" }
