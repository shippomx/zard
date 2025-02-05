package serverinterceptors

import (
	"context"
	"fmt"

	"github.com/shippomx/zard/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UserIDInterceptor(ctx context.Context, req any,
	_ *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// metadata format lowercase
		fmt.Println(md)
		uids := md.Get("x-gate-user-id")
		if len(uids) > 0 {
			newctx := context.WithValue(ctx, logx.UserIDContextKey, uids[0])
			resp, err := handler(newctx, req)
			return resp, err
		}
	}
	resp, err := handler(ctx, req)
	return resp, err
}
