package gateway

import (
	"context"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestConfig(t *testing.T) {
	gConfig := GatewayConf{}
	assert.Equal(t, gConfig.DisableTraces, false)
}

func TestMustNewServer(t *testing.T) {
	var c GatewayConf

	// Avoid popup alert on macOS for asking permissions
	c.RpcServer.ListenOn = "localhost:18888"
	c.Rest.Host = "localhost"
	c.Rest.Port = 18889
	// Create a mock register function
	mockReg := func(_ context.Context, _ *runtime.ServeMux, _ *grpc.ClientConn) error {
		return nil
	}
	server := MustNewServer(c, mockReg)
	assert.NotNil(t, server)
}
