package internal

import (
	"errors"
	"net/http"
	"testing"

	"github.com/shippomx/zard/core/utils"
)

func TestMetricsInterceptor(t *testing.T) {
	tests := []struct {
		name         string
		resp         *http.Response
		err          error
		ex           ExtendInfo
		expectedCode int
	}{
		{
			name:         "nil response and error",
			resp:         nil,
			err:          errors.New("test error"),
			ex:           ExtendInfo{},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "non-nil response and no error",
			resp: &http.Response{
				StatusCode: http.StatusOK,
			},
			err: nil,
			ex:  ExtendInfo{},
		},
		{
			name: "non-nil response and error",
			resp: &http.Response{
				StatusCode: http.StatusInternalServerError,
			},
			err: errors.New("test error"),
			ex:  ExtendInfo{},
		},
		{
			name: "empty ExtendInfo",
			resp: &http.Response{
				StatusCode: http.StatusOK,
			},
			err: nil,
			ex:  ExtendInfo{},
		},
		{
			name: "non-empty ExtendInfo and MetricURL set to true",
			resp: &http.Response{
				StatusCode: http.StatusOK,
			},
			err: nil,
			ex: ExtendInfo{
				EnableMetricURL: true,
			},
		},
		{
			name: "non-empty ExtendInfo and MetricURL set to false",
			resp: &http.Response{
				StatusCode: http.StatusOK,
			},
			err: nil,
			ex: ExtendInfo{
				EnableMetricURL: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			r, _ := http.NewRequest("GET", "http://example.com", nil)
			_, handler := MetricsInterceptor(r, tt.ex)
			handler(tt.resp, tt.err)
		})
	}
}

func TestMetric(_ *testing.T) {
	var path string
	MetricClientReqDur.ObserveFloat(float64(1), utils.BuildVersion, "0", path)
	MetricClientReqCodeTotal.Inc(utils.BuildVersion, "0", path, "200")
}
