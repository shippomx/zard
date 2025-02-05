package serverinterceptors

import (
	"context"
	"testing"

	"github.com/shippomx/zard/core/logx"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func init() {
	logx.Disable()
}

func TestStreamCrashInterceptor(t *testing.T) {
	err := StreamRecoverInterceptor(nil, nil, nil, func(
		svr any, stream grpc.ServerStream) error {
		panic("mock panic")
	})
	assert.NotNil(t, err)
}

func TestUnaryCrashInterceptor(t *testing.T) {
	_, err := UnaryRecoverInterceptor(context.Background(), nil, nil,
		func(ctx context.Context, req any) (any, error) {
			panic("mock panic")
		})
	assert.NotNil(t, err)
}
