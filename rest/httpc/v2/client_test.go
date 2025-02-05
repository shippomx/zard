package httpc_test

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/shippomx/zard/rest/httpc/v2"
	"github.com/shippomx/zard/rest/httpc/v2/nds"
	"github.com/stretchr/testify/assert"
)

func TestDefaultClient(t *testing.T) {
	if os.Getenv("NACOS_SERVICE_NAME") == "" {
		t.Skip("nacos env not set")
	}
	_url, err := url.Parse("nacos://local-test")
	assert.NoError(t, err, "")
	res, err := httpc.DoRequest(&http.Request{Method: "GET", URL: _url})
	if err == nil {
		defer res.Body.Close()
	}
	assert.NoError(t, err, "")
}

func TestNacosServiceClient(t *testing.T) {
	if os.Getenv("NACOS_SERVICE_NAME") == "" {
		t.Skip("nacos env not set")
	}
	conf := nds.NacosDiscoveryConfig{
		IPAddr:      "127.0.0.1",
		GroupName:   "DEFAULT_GROUP",
		Port:        8848,
		Clusters:    []string{"DEFAULT"},
		Timeout:     10,
		Username:    "nacos",
		Password:    "nacos",
		LogLevel:    "info",
		NamespaceID: "public",
	}
	nt := nds.NacosTransport{}
	err := nt.Register(&conf)
	assert.NoError(t, err, "")
	client, err := httpc.NewServieClient(httpc.HTTPClientConfig{
		MaxConnsPerHost: 10,
	}, httpc.WithNacosDiscovery(&nt))
	assert.NoError(t, err, "")
	_url, err := url.Parse("nacos://local-test:8000")
	assert.NoError(t, err, "")
	for i := 0; i < 10; i++ {
		go func() {
			res, err := client.DoRequest(&http.Request{Method: "GET", URL: _url})
			if err == nil {
				defer res.Body.Close()
			}
			assert.NoError(t, err, "")
		}()
	}
}

func TestNewClientService(t *testing.T) {
	mn := nds.GetMockNacosTransport(t)

	client, err := httpc.NewServieClient(httpc.HTTPClientConfig{
		MaxConnsPerHost: 10,
	}, httpc.WithNacosDiscovery(mn))
	assert.NoError(t, err, "")

	_url, err := url.Parse("nacos://test:8000")
	assert.NoError(t, err, "")

	res, err := client.DoRequest(&http.Request{Method: "GET", URL: _url})
	if err == nil {
		defer res.Body.Close()
	}
	assert.NoError(t, err, "")

	res, err = client.Do(context.Background(), "GET", _url.String(), nil)
	if err == nil {
		defer res.Body.Close()
	}
	assert.NoError(t, err, "")
	var testStruc struct {
		Name string
		Age  int
	}
	res, err = client.Do(context.Background(), "GET", _url.String(), &testStruc)
	if err == nil {
		defer res.Body.Close()
	}
	assert.NoError(t, err, "")
}
