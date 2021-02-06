package interceptor

import (
	"context"

	std "github.com/wednesdaysunny/onerpc/eco/inter"
	"google.golang.org/grpc"
)

func GetSentryServerInterceptors() ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	unarys := []grpc.UnaryServerInterceptor{
		GetSentryUnaryServerInterceptor(),
	}
	streams := []grpc.StreamServerInterceptor{
		GetSentryStreamServerInterceptor(),
	}
	return unarys, streams
}

func GetSentryUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer std.RecoverRepanicWithSentry(ctx, req)
		return handler(ctx, req)
	}
}
func GetSentryStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(src interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		defer std.RecoverRepanicWithSentry(context.Background(), nil)
		return handler(src, ss)
	}
}
