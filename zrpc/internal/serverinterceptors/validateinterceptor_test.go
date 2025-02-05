package serverinterceptors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	gcodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateInterceptor(t *testing.T) {
	ctx := context.Background()
	info := &grpc.UnaryServerInfo{}
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, nil
	}

	t.Run("Validate method exists and returns nil", func(t *testing.T) {
		req := &mockRequest{
			validateFunc: func() error {
				return nil
			},
		}
		resp, err := ValidateInterceptor(ctx, req, info, handler)
		assert.Nil(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Validate method exists and returns an error", func(t *testing.T) {
		expectedErr := errors.New("validation error")
		req := &mockRequest{
			validateFunc: func() error {
				return expectedErr
			},
		}
		resp, err := ValidateInterceptor(ctx, req, info, handler)
		assert.Equal(t, status.Error(gcodes.InvalidArgument, expectedErr.Error()), err)
		assert.Nil(t, resp)
	})

	t.Run("Validate method does not exist", func(t *testing.T) {
		req := &mockRequest{}
		expectedResp := "response"
		expectedErr := errors.New("handler error")
		handler = func(ctx context.Context, req any) (any, error) {
			return expectedResp, expectedErr
		}
		resp, err := ValidateInterceptor(ctx, req, info, handler)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, expectedErr, err)
	})
}

type mockRequest struct {
	validateFunc func() error
}

func (m *mockRequest) Validate() error {
	if m.validateFunc != nil {
		return m.validateFunc()
	}
	return nil
}
