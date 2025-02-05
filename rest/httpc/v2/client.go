package httpc

import (
	"context"
	"net/http"

	"github.com/shippomx/zard/core/breaker"
	"github.com/shippomx/zard/rest/httpc/v2/internal"
)

type Client interface {
	Do(ctx context.Context, method, url string, data any) (*http.Response, error)
	DoRequest(r *http.Request) (*http.Response, error)
	Close()
}

type HTTPClient struct {
	client          *http.Client
	brk             breaker.Breaker
	EnableMetricURL bool
	interceptors    []internal.Interceptor
	OnClose         func()
}

type Option func(r *HTTPClient)

func NewServieClient(config HTTPClientConfig, opts ...Option) (Client, error) {
	var b breaker.Breaker
	if config.Breaker {
		b = breaker.NewBreaker()
	}
	client := &HTTPClient{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:          config.MaxIdleConns,
				MaxConnsPerHost:       config.MaxConnsPerHost,
				TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
				DisableKeepAlives:     config.DisableKeepAlives,
				DisableCompression:    config.DisableCompression,
				IdleConnTimeout:       config.IdleConnTimeout,
				ResponseHeaderTimeout: config.ResponseHeaderTimeout,
				ExpectContinueTimeout: config.ExpectContinueTimeout,
				MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
				ForceAttemptHTTP2:     config.ForceAttemptHTTP2,
			},
		},
		brk:             b,
		EnableMetricURL: config.EnableMetricURL,
		interceptors:    []internal.Interceptor{},
	}

	for _, opt := range opts {
		opt(client)
	}
	if config.Breaker {
		client.interceptors = append(client.interceptors, internal.BreakerInterceptor)
	}
	if config.Metric {
		client.interceptors = append(client.interceptors, internal.MetricsInterceptor)
	}
	if config.Logger {
		client.interceptors = append(client.interceptors, internal.LogInterceptor)
	}
	if config.Tracing {
		client.interceptors = append(client.interceptors, internal.TracingInterceptor)
	}
	return client, nil
}

func (c *HTTPClient) GetHTTPTransport() *http.Transport {
	if t, ok := c.client.Transport.(*http.Transport); ok {
		return t
	}
	return &http.Transport{}
}

type nacosInterface interface {
	RegisterTransport(t *http.Transport)
	Close()
}

func WithNacosDiscovery(t http.RoundTripper) Option {
	return func(c *HTTPClient) {
		if t, ok := t.(nacosInterface); ok {
			t.RegisterTransport(c.GetHTTPTransport())
			c.OnClose = t.Close
		}
		c.GetHTTPTransport().RegisterProtocol("nacos", t)
		c.GetHTTPTransport().RegisterProtocol("nacoss", t)
	}
}

func (c *HTTPClient) DoRequest(r *http.Request) (*http.Response, error) {
	respHandlers := make([]internal.ResponseHandler, len(c.interceptors))
	var p breaker.Promise
	var err error
	if c.brk != nil && c.brk.Name() != "" {
		p, err = c.brk.Allow()
		if err != nil {
			return nil, err
		}
	}
	for i, in := range c.interceptors {
		newr, h := in(r, internal.ExtendInfo{Promise: p, EnableMetricURL: c.EnableMetricURL})
		respHandlers[i] = h
		r = newr
	}
	resp, err := c.client.Do(r)
	for _, h := range respHandlers {
		h(resp, err)
	}
	return resp, err
}

func (c *HTTPClient) Do(ctx context.Context, method, url string, data any) (*http.Response, error) {
	req, err := buildRequest(ctx, method, url, data)
	if err != nil {
		return nil, err
	}
	return c.DoRequest(req)
}

func (c *HTTPClient) Close() {
	c.client.CloseIdleConnections()
}
