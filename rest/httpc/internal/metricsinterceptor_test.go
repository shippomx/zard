package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shippomx/zard/core/logx"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMetricsInterceptor(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	logx.Disable()

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	req, err := http.NewRequest(http.MethodGet, svr.URL, nil)
	assert.NotNil(t, req)
	assert.Nil(t, err)
	interceptor := MetricsInterceptor("test", nil)
	req, handler := interceptor(req, false)
	resp, err := http.DefaultClient.Do(req)
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	handler(resp, err)
}
