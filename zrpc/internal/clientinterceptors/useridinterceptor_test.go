package clientinterceptors

import (
	"context"
	"testing"

	"github.com/shippomx/zard/core/logx"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUserIDInterceptor(t *testing.T) {
	// Test case 1: With X-Gate-User-Id metadata
	ctx := context.Background()
	// md := metadata.Pairs()
	req := "request"
	reply := "reply"
	cc := &grpc.ClientConn{}
	invoker := func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		uid := ctx.Value(logx.UserIDContextKey).(string)
		assert.Equal(t, "123", uid)
		return nil
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{})
	ctx = metadata.AppendToOutgoingContext(ctx, "X-Gate-User-Id", "123")
	opts := []grpc.CallOption{}
	err := UserIDInterceptor(ctx, "method", req, reply, cc, invoker, opts...)
	assert.NoError(t, err)

	// Test case 2: Without X-Gate-User-Id metadata
	invoker = func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		_, ok := ctx.Value(logx.UserIDContextKey).(string)
		assert.False(t, ok)
		return nil
	}
	err = UserIDInterceptor(context.Background(), "method", req, reply, cc, invoker, opts...)
	assert.NoError(t, err)
}
