package http

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shippomx/zard/core/logx"
	errorx "github.com/shippomx/zard/rest/errors"
	"github.com/shippomx/zard/rest/test"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMain(m *testing.M) {
	logx.Disable()
	m.Run()
}

func TestJsonBaseResponse(t *testing.T) {
	executor := test.NewExecutor[any, testWriterResult](comparisonOption)
	executor.Add([]test.Data[any, testWriterResult]{
		{
			Name:  "code-msg-pointer",
			Input: errorx.New(1, "test"),
			Want: testWriterResult{
				code:        200,
				writeString: `{"code":1,"message":"test"}`,
			},
		},
		{
			Name:  "code-msg-struct",
			Input: errorx.CodeMsg{Code: 1, Message: "test"},
			Want: testWriterResult{
				code:        200,
				writeString: `{"code":1,"message":"test"}`,
			},
		},
		{
			Name:  "status.Status",
			Input: status.New(codes.OK, "ok"),
			Want: testWriterResult{
				code:        200,
				writeString: `{"code":0,"message":"ok"}`,
			},
		},
		{
			Name:  "error",
			Input: errors.New("test"),
			Want: testWriterResult{
				code:        200,
				writeString: `{"code":-1,"message":"test"}`,
			},
		},
		{
			Name:  "struct",
			Input: message{Name: "anyone"},
			Want: testWriterResult{
				code:        200,
				writeString: `{"code":0,"data":{"name":"anyone"},"message":"ok"}`,
			},
		},
	}...)
	executor.RunE(t, func(a any) (testWriterResult, error) {
		w := &tracedResponseWriter{headers: make(map[string][]string)}
		JsonBaseResponse(w, a)

		result, err := w.result()
		if err != nil {
			t.Fatalf("获取结果失败: %v", err)
		}

		// 解析返回的JSON字符串
		var response map[string]interface{}
		err = json.Unmarshal([]byte(result.writeString), &response)
		if err != nil {
			t.Fatalf("解析JSON字符串失败: %v", err)
		}

		delete(response, "timestamp")
		respByte, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("序列化JSON字符串失败: %v", err)
		}

		return testWriterResult{
			code:        200,
			writeString: string(respByte),
		}, nil
	})
}

func TestJsonBaseResponseCtx(t *testing.T) {
	executor := test.NewExecutor[any, testWriterResult](comparisonOption)
	executor.Add([]test.Data[any, testWriterResult]{
		{
			Name:  "code-msg-pointer",
			Input: errorx.New(1, "test"),
			Want: testWriterResult{
				code:        200,
				writeString: `{"code":1,"message":"test"}`,
			},
		},
		{
			Name:  "code-msg-struct",
			Input: errorx.CodeMsg{Code: 1, Message: "test"},
			Want: testWriterResult{
				code:        200,
				writeString: `{"code":1,"message":"test"}`,
			},
		},
		{
			Name:  "status.Status",
			Input: status.New(codes.OK, "ok"),
			Want: testWriterResult{
				code:        200,
				writeString: `{"code":0,"message":"ok"}`,
			},
		},
		{
			Name:  "error",
			Input: errors.New("test"),
			Want: testWriterResult{
				code:        200,
				writeString: `{"code":-1,"message":"test"}`,
			},
		},
		{
			Name:  "struct",
			Input: message{Name: "anyone"},
			Want: testWriterResult{
				code:        200,
				writeString: `{"code":0,"data":{"name":"anyone"},"message":"ok"}`,
			},
		},
	}...)
	executor.RunE(t, func(a any) (testWriterResult, error) {
		w := &tracedResponseWriter{headers: make(map[string][]string)}
		JsonBaseResponseCtx(context.TODO(), w, a)
		result, err := w.result()
		if err != nil {
			t.Fatalf("获取结果失败: %v", err)
		}

		// 解析返回的JSON字符串
		var response map[string]interface{}
		err = json.Unmarshal([]byte(result.writeString), &response)
		if err != nil {
			t.Fatalf("解析JSON字符串失败: %v", err)
		}

		delete(response, "timestamp")
		respByte, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("序列化JSON字符串失败: %v", err)
		}

		return testWriterResult{
			code:        200,
			writeString: string(respByte),
		}, nil
	})
}

func TestXmlBaseResponse(t *testing.T) {
	executor := test.NewExecutor[any, testWriterResult](comparisonOption)
	executor.Add([]test.Data[any, testWriterResult]{
		{
			Name:  "code-msg",
			Input: errorx.New(1, "test"),
			Want: testWriterResult{
				code:        200,
				writeString: `<xml version="1.0" encoding="UTF-8"><code>1</code><message>test</message></xml>`,
			},
		},
		{
			Name:  "status.Status",
			Input: status.New(codes.OK, "ok"),
			Want: testWriterResult{
				code:        200,
				writeString: `<xml version="1.0" encoding="UTF-8"><code>0</code><message>ok</message></xml>`,
			},
		},
		{
			Name:  "error",
			Input: errors.New("test"),
			Want: testWriterResult{
				code:        200,
				writeString: `<xml version="1.0" encoding="UTF-8"><code>-1</code><message>test</message></xml>`,
			},
		},
		//{
		//	Name:  "struct",
		//	Input: message{Name: "anyone"},
		//	Want: testWriterResult{
		//		code:        200,
		//		writeString: `<xml version="1.0" encoding="UTF-8"><code>0</code><message>ok</message><data><name>anyone</name></data></xml>`,
		//	},
		//},
	}...)
	executor.RunE(t, func(a any) (testWriterResult, error) {
		w := &tracedResponseWriter{headers: make(map[string][]string)}
		XmlBaseResponse(w, a)
		result, err := w.result()
		if err != nil {
			t.Fatalf("获取结果失败: %v", err)
		}

		// 解析返回的XML字符串
		var response baseXmlResponse[any]
		err = xml.Unmarshal([]byte(result.writeString), &response)
		if err != nil {
			t.Fatalf("解析XML字符串失败: %v", err)
		}

		// 删除timestamp字段
		response.BaseResponse.Timestamp = 0

		// 重新序列化XML
		respByte, err := xml.Marshal(response)
		if err != nil {
			t.Fatalf("序列化XML字符串失败: %v", err)
		}

		return testWriterResult{
			code:        200,
			writeString: string(respByte),
		}, nil
	})
}

func TestXmlBaseResponseCtx(t *testing.T) {
	executor := test.NewExecutor[any, testWriterResult](comparisonOption)
	executor.Add([]test.Data[any, testWriterResult]{
		{
			Name:  "code-msg",
			Input: errorx.New(1, "test"),
			Want: testWriterResult{
				code:        200,
				writeString: `<xml version="1.0" encoding="UTF-8"><code>1</code><message>test</message></xml>`,
			},
		},
		{
			Name:  "status.Status",
			Input: status.New(codes.OK, "ok"),
			Want: testWriterResult{
				code:        200,
				writeString: `<xml version="1.0" encoding="UTF-8"><code>0</code><message>ok</message></xml>`,
			},
		},
		{
			Name:  "error",
			Input: errors.New("test"),
			Want: testWriterResult{
				code:        200,
				writeString: `<xml version="1.0" encoding="UTF-8"><code>-1</code><message>test</message></xml>`,
			},
		},
		//{
		//	Name:  "struct",
		//	Input: message{Name: "anyone"},
		//	Want: testWriterResult{
		//		code:        200,
		//		writeString: `<xml version="1.0" encoding="UTF-8"><code>0</code><message>ok</message><data><name>anyone</name></data></xml>`,
		//	},
		//},
	}...)
	executor.RunE(t, func(a any) (testWriterResult, error) {
		w := &tracedResponseWriter{headers: make(map[string][]string)}
		XmlBaseResponseCtx(context.TODO(), w, a)
		result, err := w.result()
		if err != nil {
			t.Fatalf("获取结果失败: %v", err)
		}

		// 解析返回的XML字符串
		var response baseXmlResponse[any]
		err = xml.Unmarshal([]byte(result.writeString), &response)
		if err != nil {
			t.Fatalf("解析XML字符串失败: %v", err)
		}

		// 删除timestamp字段
		response.BaseResponse.Timestamp = 0

		// 重新序列化XML
		respByte, err := xml.Marshal(response)
		if err != nil {
			t.Fatalf("序列化XML字符串失败: %v", err)
		}

		return testWriterResult{
			code:        200,
			writeString: string(respByte),
		}, nil
	})
}

var comparisonOption = test.WithComparison[any, testWriterResult](func(t *testing.T, expected, actual testWriterResult) {
	assert.Equal(t, expected.code, actual.code)
	assert.Equal(t, expected.writeString, actual.writeString)
})

func TestJsonErrResponseCtx(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		ctx := context.Background()
		w := httptest.NewRecorder()
		JsonErrResponseCtx(ctx, w, nil)
		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}
		w = httptest.NewRecorder()
		var response BaseResponse[any]
		JsonErrResponseCtx(ctx, w, nil)
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %s - %v", w.Body.Bytes(), err)
		}
	})

	t.Run("non-nil error", func(t *testing.T) {
		ctx := context.Background()
		w := httptest.NewRecorder()
		err := errors.New("test error")
		JsonErrResponseCtx(ctx, w, err)
		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}
		w = httptest.NewRecorder()
		err = errorx.New(334, err.Error())
		var response BaseResponse[any]
		JsonErrResponseCtx(ctx, w, err)
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if response.Code != 334 {
			t.Errorf("expected code -1, got %d", response.Code)
		}
		if response.Message != "test error" {
			t.Errorf("expected message %s, got %s", err.Error(), response.Message)
		}
		if response.Timestamp == 0 {
			t.Errorf("expected timestamp > 0, got %d", response.Timestamp)
		}
		if response.Label != nil {
			t.Errorf("%s expected label nil, got %v", w.Body.Bytes(), response.Label)
		}
		if strings.Contains(w.Body.String(), "label") {
			t.Errorf("%s expected not contains label", w.Body.String())
		}
		w = httptest.NewRecorder()
		err = errors.New("test")
		JsonErrResponseCtx(ctx, w, err)
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if response.Code != -1 {
			t.Errorf("expected code 0, got %d", response.Code)
		}
		w = httptest.NewRecorder()
		err = errorx.New(334, "test", errorx.WithLabel("label"))
		JsonErrResponseCtx(ctx, w, err)
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if response.Label == nil {
			t.Errorf("expected label not nil, got %v", response.Label)
		}
	})
}
