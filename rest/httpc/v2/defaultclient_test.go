package httpc

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisableBreaker(_ *testing.T) {
	DisableBreaker()
}

func TestEnableEnableMetricURL(_ *testing.T) {
	EnableMetricURL()
}

func TestEnableDefaultMiddle(_ *testing.T) {
	EnableDefaultMiddleware(true, true, true, true)
}

// nolint: all //only check panic
func TestDefaultClient_NoPanic(t *testing.T) {
	resp, err := Do(context.Background(), "GET", "https://err-localhost", nil)
	if err == nil {
		defer resp.Body.Close()
	}
	assert.Error(t, err)
	resp, err = DoRequest(&http.Request{Method: "GET", URL: &url.URL{}})
	if err == nil {
		defer resp.Body.Close()
	}
	assert.Error(t, err)
	Close()
}
