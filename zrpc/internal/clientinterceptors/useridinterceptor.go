package clientinterceptors

import (
	"context"

	"github.com/shippomx/zard/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UserIDInterceptor(ctx context.Context, method string, req, reply any,
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
) error {
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		// metadata format lowercase
		uids := md.Get("x-gate-user-id")
		if len(uids) > 0 {
			newctx := context.WithValue(ctx, logx.UserIDContextKey, uids[0])
			err := invoker(newctx, method, req, reply, cc, opts...)
			return err
		}
	}
	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}
