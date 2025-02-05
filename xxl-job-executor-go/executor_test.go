package xxl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecutorCallback_Errors(t *testing.T) {
	e := &executor{
		runList: &taskList{
			data: make(map[string]*Task),
		},
		log: &logger{},
	}

	// 场景 1：发送 POST 请求失败
	t.Run("post request error", func(t *testing.T) {
		e.opts.ServerAddr = "http://localhost:8080"
		http.HandleFunc("/api/callback", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		go func() {
			// nolint
			if err := http.ListenAndServe(":8080", nil); err != nil {
				t.Errorf("Failed to start HTTP server: %v", err)
			}
		}()

		// 等待 HTTP 服务器启动
		time.Sleep(100 * time.Millisecond)

		task := &Task{
			Id:    1,
			Name:  "task1",
			Param: &RunReq{},
		}

		e.callback(task, SuccessCode, "success")
		http.DefaultServeMux = new(http.ServeMux)
	})

	// 场景 2：读取响应体失败
	t.Run("read response body error", func(t *testing.T) {
		e.opts.ServerAddr = "http://localhost:8081"
		http.HandleFunc("/api/callback", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("response body"))
			w.(http.Flusher).Flush()
			_, _ = w.Write([]byte("another response body"))
		})

		go func() {
			// nolint
			if err := http.ListenAndServe(":8081", nil); err != nil {
				t.Errorf("Failed to start HTTP server: %v", err)
			}
		}()

		// 等待 HTTP 服务器启动
		time.Sleep(100 * time.Millisecond)

		task := &Task{
			Id:    1,
			Name:  "task1",
			Param: &RunReq{},
		}

		e.callback(task, SuccessCode, "success")
		http.DefaultServeMux = new(http.ServeMux)
	})
}

func TestExecutorRunTask(t *testing.T) {
	e := &executor{
		runList: &taskList{
			data: make(map[string]*Task),
		},
		log: &logger{},
		regList: &taskList{
			data: make(map[string]*Task),
		},
	}

	// 场景 1：参数解析错误
	t.Run("params error", func(t *testing.T) {
		req, err := json.Marshal(&RunReq{
			ExecutorHandler: "test",
			JobID:           1,
		})
		if err != nil {
			t.Errorf("Failed to marshal request: %v", err)
		}

		req = append(req, []byte(`{"invalid": "json"}`)...)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/run", bytes.NewReader(req))

		e.RunTask(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "params err")
	})

	// 场景 2：任务没有注册
	t.Run("task not registered", func(t *testing.T) {
		req, err := json.Marshal(&RunReq{
			ExecutorHandler: "test",
			JobID:           1,
		})
		if err != nil {
			t.Errorf("Failed to marshal request: %v", err)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/run", bytes.NewReader(req))

		e.RunTask(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Task not registered")
	})

	// 场景 3：覆盖之前调度
	t.Run("cover early", func(t *testing.T) {
		req, err := json.Marshal(&RunReq{
			ExecutorHandler:       "test",
			JobID:                 1,
			ExecutorBlockStrategy: coverEarly,
		})
		if err != nil {
			t.Errorf("Failed to marshal request: %v", err)
		}

		regTask := &Task{
			fn: func(_ context.Context, _ *RunReq) string {
				return "success"
			},
		}
		e.regList.Set("test", regTask)
		_, cancel := context.WithCancel(context.Background())
		oldTask := &Task{
			Id:     1,
			Name:   "test",
			Param:  &RunReq{},
			log:    e.log,
			Cancel: cancel,
			fn:     regTask.fn,
		}
		e.runList.Set(Int64ToStr(oldTask.Id), oldTask)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/run", bytes.NewReader(req))

		e.RunTask(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		e.runList.Del(Int64ToStr(oldTask.Id))
	})

	// 场景 4：任务已经在运行
	t.Run("task already running", func(t *testing.T) {
		req, err := json.Marshal(&RunReq{
			ExecutorHandler: "test",
			JobID:           2,
		})
		if err != nil {
			t.Errorf("Failed to marshal request: %v", err)
		}

		regTask := &Task{
			fn: func(_ context.Context, _ *RunReq) string {
				return "success"
			},
		}
		e.regList.Set("test", regTask)

		task := &Task{
			Id:    2,
			Name:  "test",
			Param: &RunReq{},
			log:   e.log,
			fn:    regTask.fn,
		}
		e.runList.Set(Int64ToStr(task.Id), task)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/run", bytes.NewReader(req))

		e.RunTask(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "There are tasks running")
		e.runList.Del(Int64ToStr(task.Id))
	})

	// 场景 5：任务执行成功
	t.Run("task executed successfully", func(t *testing.T) {
		req, err := json.Marshal(&RunReq{
			ExecutorHandler: "test",
			JobID:           1,
		})
		if err != nil {
			t.Errorf("Failed to marshal request: %v", err)
		}

		regTask := &Task{
			fn: func(_ context.Context, _ *RunReq) string {
				return "success"
			},
		}
		e.regList.Set("test", regTask)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/run", bytes.NewReader(req))

		e.RunTask(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestExecutorKillTask(t *testing.T) {
	e := &executor{
		runList: &taskList{
			data: make(map[string]*Task),
		},
		log: &logger{},
	}

	// 场景 1：任务没有运行
	t.Run("task not running", func(t *testing.T) {
		req, err := json.Marshal(&killReq{
			JobID: 1,
		})
		if err != nil {
			t.Errorf("Failed to marshal request: %v", err)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/kill", bytes.NewReader(req))

		e.KillTask(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Task kill err")
	})

	// 场景 2：任务已经运行
	t.Run("task running", func(t *testing.T) {
		req, err := json.Marshal(&killReq{
			JobID: 1,
		})
		if err != nil {
			t.Errorf("Failed to marshal request: %v", err)
		}
		_, cancel := context.WithCancel(context.Background())
		task := &Task{
			Id:     1,
			Name:   "test",
			Param:  &RunReq{},
			log:    e.log,
			Cancel: cancel,
			fn:     func(_ context.Context, _ *RunReq) string { return "success" },
		}
		e.runList.Set(Int64ToStr(task.Id), task)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/kill", bytes.NewReader(req))

		e.KillTask(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Nil(t, e.runList.Get(Int64ToStr(task.Id)))
	})
}

func TestExecutorBeat(t *testing.T) {
	e := &executor{
		log: &logger{},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/beat", nil)

	e.Beat(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "200")
}

func TestExecutorIdleBeat(t *testing.T) {
	e := &executor{
		runList: &taskList{
			data: make(map[string]*Task),
		},
		log: &logger{},
	}

	// 场景 1：参数解析错误
	t.Run("params error", func(t *testing.T) {
		req, err := json.Marshal(&idleBeatReq{
			JobID: 1,
		})
		if err != nil {
			t.Errorf("Failed to marshal request: %v", err)
		}

		req = append(req, []byte(`{"invalid": "json"}`)...)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/idleBeat", bytes.NewReader(req))

		e.IdleBeat(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "500")
	})

	// 场景 2：任务正在运行
	t.Run("task running", func(t *testing.T) {
		req, err := json.Marshal(&idleBeatReq{
			JobID: 1,
		})
		if err != nil {
			t.Errorf("Failed to marshal request: %v", err)
		}

		task := &Task{
			Id:    1,
			Name:  "test",
			Param: &RunReq{},
			log:   e.log,
			fn:    func(_ context.Context, _ *RunReq) string { return "success" },
		}
		e.runList.Set(Int64ToStr(task.Id), task)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/idleBeat", bytes.NewReader(req))

		e.IdleBeat(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Task is busy")
		e.runList.Del(Int64ToStr(task.Id))
	})

	// 场景 3：任务参数正确
	t.Run("params correct", func(t *testing.T) {
		req, err := json.Marshal(&idleBeatReq{
			JobID: 1,
		})
		if err != nil {
			t.Errorf("Failed to marshal request: %v", err)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/idleBeat", bytes.NewReader(req))

		e.IdleBeat(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "200")
	})
}

type testLogger struct {
	entries []string
	mu      sync.RWMutex
}

func (l *testLogger) Info(format string, a ...interface{}) {
	l.mu.Lock()
	l.entries = append(l.entries, fmt.Sprintf(format, a...))
	l.mu.Unlock()
}

func (l *testLogger) Error(format string, a ...interface{}) {
	l.mu.Lock()
	l.entries = append(l.entries, fmt.Sprintf(format, a...))
	l.mu.Unlock()
}

func (l *testLogger) GetEntries() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.entries
}

func TestExecutorRegistryRemove(t *testing.T) {
	// 场景 1：注册信息解析失败
	t.Run("registry info parse error", func(t *testing.T) {
		e := &executor{
			log: &testLogger{},
		}
		e.opts.RegistryKey = ""
		e.Stop()
		entries := e.log.(*testLogger).GetEntries()
		assert.Contains(t, entries[0], "执行器摘除失败:")
	})

	// 场景 2：post 请求成功
	t.Run("post request success", func(t *testing.T) {
		e := &executor{
			log: &testLogger{},
		}
		e.opts.RegistryKey = "test"
		e.opts.ServerAddr = "http://localhost:8082"
		http.HandleFunc("/api/registryRemove", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"Code": 200, "Message": "摘除成功"}`))
		})

		go func() {
			// nolint
			if err := http.ListenAndServe(":8082", nil); err != nil {
				t.Errorf("Failed to start HTTP server: %v", err)
			}
		}()

		// 等待 HTTP 服务器启动.
		time.Sleep(100 * time.Millisecond)

		e.Stop()
		entries := e.log.(*testLogger).GetEntries()
		assert.Contains(t, entries[0], "执行器摘除成功:")
	})
}

func TestTaskLog(t *testing.T) {
	e := &executor{
		log: &testLogger{},
	}

	t.Run("normal request", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/task/log", strings.NewReader(`{"key":"value"}`))
		if err != nil {
			t.Errorf("Failed to create request: %v", err)
		}
		w := httptest.NewRecorder()
		e.TaskLog(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
		entries := e.log.(*testLogger).GetEntries()
		assert.Contains(t, entries[0], "日志请求参数:")
	})
}

func TestExecutorInit(t *testing.T) {
	e := &executor{}
	opts := []Option{
		SetLogger(&testLogger{}),
		ExecutorIp("127.0.0.1"),
		ExecutorPort("8080"),
	}
	e.Init(opts...)
	assert.NotNil(t, e.log)
	assert.NotNil(t, e.regList)
	assert.NotNil(t, e.runList)
	assert.Equal(t, "127.0.0.1:8080", e.address)
}

func TestExecutorRun(t *testing.T) {
	e := &executor{
		log: &testLogger{},
	}
	e.opts.ExecutorPort = "8088"
	go func() {
		_ = e.Run()
	}()
	time.Sleep(100 * time.Millisecond)
	assert.NotNil(t, e.log.(*testLogger).GetEntries())
	assert.Contains(t, e.log.(*testLogger).GetEntries()[0], "Starting server at ")
	req, err := http.NewRequest("GET", "http://localhost:8088/run", nil)
	if err != nil {
		t.Errorf("Failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestExecutorRegTask(t *testing.T) {
	e := &executor{
		regList: &taskList{
			data: make(map[string]*Task),
		},
	}
	pattern := "test"
	task := func(_ context.Context, _ *RunReq) string {
		return "test"
	}
	e.RegTask(pattern, task)
	taskObj := e.regList.Get(pattern)
	assert.NotNil(t, taskObj)
	assert.NotNil(t, taskObj.fn)
}

func TestNewExecutor(t *testing.T) {
	t.Run("with options", func(t *testing.T) {
		e := NewExecutor(
			SetLogger(&testLogger{}),
			ExecutorIp("127.0.0.1"),
			ExecutorPort("8080"),
		)
		assert.NotNil(t, e)
		assert.Equal(t, "127.0.0.1", e.(*executor).opts.ExecutorIp)
		assert.Equal(t, "8080", e.(*executor).opts.ExecutorPort)
		assert.NotNil(t, e.(*executor).opts.l)
	})
}

func TestExecutorLogHandler(t *testing.T) {
	e := &executor{}
	handler := defaultLogHandler
	e.LogHandler(handler)
	assert.NotNil(t, e.logHandler)
}

func TestExecutorUse(t *testing.T) {
	e := &executor{}
	middleware := e.chain
	e.Use(middleware)
	assert.NotNil(t, e.middlewares)
}
