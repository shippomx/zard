package nds

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNacosTransport(t *testing.T) {
	n := GetMockNacosTransport(t)
	n.nacosClient = NewMockINamingClient(t)
}

func TestNacosTransport_RegisterTransport(t *testing.T) {
	n := GetMockNacosTransport(t)
	n.RegisterTransport(&http.Transport{})
}

func TestNacosTransport_RoundTrip(t *testing.T) {
	n := GetMockNacosTransport(t)

	resp, err := n.RoundTrip(&http.Request{})
	if err == nil {
		defer resp.Body.Close()
	}
	assert.Error(t, err)
	resp, err = n.RoundTrip(&http.Request{URL: &url.URL{}})
	if err == nil {
		defer resp.Body.Close()
	}
	assert.Error(t, err)
	resp, err = n.RoundTrip(&http.Request{Method: http.MethodGet, URL: &url.URL{}})
	if err == nil {
		defer resp.Body.Close()
	}
	assert.Error(t, err)

	n.RegisterTransport(&http.Transport{})
	u, err := url.Parse("nacos://test:8000")
	assert.NoError(t, err)
	resp, err = n.RoundTrip(httptest.NewRequest(http.MethodGet, u.String(), http.NoBody))
	if err == nil {
		defer resp.Body.Close()
	}
	assert.NoError(t, err)
	defer resp.Body.Close()
}

func TestNacosTransport_Close(t *testing.T) {
	tests := []struct {
		name           string
		nacosTransport *NacosTransport
	}{
		{
			name: "initialized NacosTransport",
			nacosTransport: &NacosTransport{
				inited:         true,
				nacostransport: &http.Transport{},
				nacosClient:    &MockINamingClient{},
			},
		},
		{
			name: "uninitialized NacosTransport",
			nacosTransport: &NacosTransport{
				inited: false,
			},
		},
		{
			name: "nil nacostransport",
			nacosTransport: &NacosTransport{
				inited:      true,
				nacosClient: &MockINamingClient{},
			},
		},
		{
			name: "nil nacosClient",
			nacosTransport: &NacosTransport{
				inited:         true,
				nacostransport: &http.Transport{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.nacosTransport.Close()
			assert.False(t, tt.nacosTransport.inited)
			if tt.nacosTransport.nacostransport != nil {
				assert.NotNil(t, tt.nacosTransport.nacostransport.CloseIdleConnections)
			}
			if tt.nacosTransport.nacosClient != nil {
				assert.NotNil(t, tt.nacosTransport.nacosClient.CloseClient)
			}
		})
	}
}
