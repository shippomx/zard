package internal

import (
	"context"
	"errors"
	"net/http"
	"testing"

	ztrace "github.com/shippomx/zard/core/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

func TestTracingInterceptor(t *testing.T) {
	tests := []struct {
		name       string
		req        *http.Request
		resp       *http.Response
		err        error
		wantStatus codes.Code
	}{
		{
			name: "successful tracing",
			req:  &http.Request{Host: "example.com"},
			resp: &http.Response{StatusCode: http.StatusOK},
		},
		{
			name: "tracing with error",
			req:  &http.Request{Host: "example.com"},
			resp: &http.Response{StatusCode: http.StatusInternalServerError},
			err:  errors.New("test error"),
		},
		{
			name: "tracing with invalid request",
			req:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tracer := otel.Tracer(ztrace.TraceName)
			_ = otel.GetTextMapPropagator()
			_, span := tracer.Start(ctx, "test-span")
			defer span.End()

			r, handler := TracingInterceptor(tt.req, ExtendInfo{})

			if r == nil {
				if tt.req != nil {
					t.Errorf("TracingInterceptor returned nil request")
				}
			}

			if handler != nil {
				handler(tt.resp, tt.err)
			}
		})
	}
}
