package httpc

import "time"

type HTTPClientConfig struct {
	TLSHandshakeTimeout   time.Duration `json:",optional"`
	DisableKeepAlives     bool          `json:",default=false"`
	DisableCompression    bool          `json:",default=false"`
	MaxIdleConns          int           `json:",optional"`
	MaxIdleConnsPerHost   int           `json:",optional"`
	MaxConnsPerHost       int           `json:",optional"`
	ExpectContinueTimeout time.Duration `json:",optional"`
	IdleConnTimeout       time.Duration `json:",optional"`
	ResponseHeaderTimeout time.Duration `json:",optional"`

	Metric            bool `json:",default=true"`
	EnableMetricURL   bool `json:",default=false"`
	Tracing           bool `json:",default=true"`
	Breaker           bool `json:",default=true"`
	Logger            bool `json:",default=true"`
	ForceAttemptHTTP2 bool `json:",default=false"`
}
