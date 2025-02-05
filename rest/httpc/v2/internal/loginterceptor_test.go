package internal

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shippomx/zard/core/logx"
)

func TestLogInterceptor(t *testing.T) {
	tests := []struct {
		name       string
		req        *http.Request
		resp       *http.Response
		err        error
		wantLog    string
		wantLogger logx.Logger
	}{
		{
			name:    "error",
			req:     httptest.NewRequest("GET", "http://example.com", nil),
			err:     errors.New("test error"),
			wantLog: `[HTTP]  [Client] GET http://example.com - test error`,
		},
		{
			name:    "successful response",
			req:     httptest.NewRequest("GET", "http://example.com", nil),
			resp:    &http.Response{StatusCode: 200},
			wantLog: `[HTTP] [Client] 200 - GET http://example.com`,
		},
		{
			name:    "non-ok response",
			req:     httptest.NewRequest("GET", "http://example.com", nil),
			resp:    &http.Response{StatusCode: 404},
			wantLog: `[HTTP] [Client] 404 - GET http://example.com`,
		},
		{
			name:    "nil response",
			req:     httptest.NewRequest("GET", "http://example.com", nil),
			wantLog: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			_, handler := LogInterceptor(tt.req, ExtendInfo{})
			handler(tt.resp, tt.err)
		})
	}
}
