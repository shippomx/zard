package clientinterceptors

import (
	"context"

	"google.golang.org/grpc"
)

func ValidateInterceptor(ctx context.Context, method string, req, reply any,
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if v, ok := req.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	err := invoker(ctx, method, req, reply, cc, opts...)
	if v, ok := reply.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return err
}
