package serverinterceptors

import (
	"context"

	"google.golang.org/grpc"
	gcodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidateInterceptor(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if v, ok := req.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return nil, status.Error(gcodes.InvalidArgument, err.Error())
		}
	}
	resp, err := handler(ctx, req)
	return resp, err
}
