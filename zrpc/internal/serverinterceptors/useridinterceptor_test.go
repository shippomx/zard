package serverinterceptors

import (
	"context"
	"testing"

	"github.com/shippomx/zard/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUserIDInterceptor(t *testing.T) {
	// Test case 1: Metadata exists and user ID is present
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("X-Gate-User-Id", "123"))
	req := struct{}{}
	info := &grpc.UnaryServerInfo{}
	handler := func(ctx context.Context, _ any) (any, error) {
		userID, ok := ctx.Value(logx.UserIDContextKey).(string)
		if !ok {
			t.Errorf("User ID not set in context")
		}
		if userID != "123" {
			t.Errorf("Expected user ID '123', got '%s'", userID)
		}
		return struct{}{}, nil
	}
	resp, err := UserIDInterceptor(ctx, req, info, handler)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp == nil {
		t.Error("Expected non-nil response")
	}

	// Test case 2: Metadata exists but user ID is not present
	handler = func(ctx context.Context, _ any) (any, error) {
		userID, ok := ctx.Value(logx.UserIDContextKey).(string)
		if !ok {
			t.Errorf("User ID not set in context")
		}
		if userID != "" {
			t.Errorf("Expected user ID '123', got '%s'", userID)
		}
		return struct{}{}, nil
	}
	ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("X-Gate-User-Id", ""))
	resp, err = UserIDInterceptor(ctx, req, info, handler)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp == nil {
		t.Error("Expected non-nil response")
	}

	// Test case 3: Metadata does not exist
	handler = func(ctx context.Context, _ any) (any, error) {
		_, ok := ctx.Value(logx.UserIDContextKey).(string)
		if !ok {
			return struct{}{}, nil
		}
		t.Errorf("User ID set in context")
		return struct{}{}, nil
	}
	ctx = context.Background()
	resp, err = UserIDInterceptor(ctx, req, info, handler)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp == nil {
		t.Error("Expected non-nil response")
	}
}
