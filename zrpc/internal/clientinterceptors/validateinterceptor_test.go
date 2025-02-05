package clientinterceptors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type testReq struct{}

type testReply struct{}

func (t *testReq) Validate() error {
	return errors.New("req validate error")
}

func (t *testReply) Validate() error {
	return errors.New("rep validate error")
}

func TestValidateInterceptor(t *testing.T) {
	ctx := context.Background()
	method := "testMethod"
	cc := &grpc.ClientConn{}
	invoker := func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		return nil
	}

	t.Run("with Validate method", func(t *testing.T) {
		req := &testReq{}
		reply := &testReply{}

		err := ValidateInterceptor(ctx, method, req, reply, cc, invoker)
		assert.Contains(t, err.Error(), "req validate error")
	})

	t.Run("no Validate method", func(t *testing.T) {
		req := struct{}{}
		reply := struct{}{}

		err := ValidateInterceptor(ctx, method, req, reply, cc, invoker)

		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})
}
