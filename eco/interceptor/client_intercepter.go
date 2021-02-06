package interceptor

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/wednesdaysunny/onerpc/eco/inter/toolkit/strutil"

	occ "github.com/wednesdaysunny/onerpc/eco/inter/common"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	oc "github.com/wednesdaysunny/onerpc/eco/inter"
	oconf "github.com/wednesdaysunny/onerpc/eco/inter/conf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	stop := time.Now()
	l := stop.Sub(start)
	logField := oc.LogFields{
		"type":          "grpcaccess",
		"remote_ip":     occ.PbMetaGet(occ.Md_CLIENTIP, ctx),
		"host":          occ.PbMetaGet(occ.Md_Host, ctx),
		"uri":           occ.PbMetaGet(occ.Md_Uri, ctx),
		"grpc_method":   info.FullMethod,
		"app":           oconf.ConfSvcName(),
		"request_data":  strutil.FromObject(req),
		"method":        occ.PbMetaGet(occ.Md_Method, ctx),
		"path":          occ.PbMetaGet(occ.Md_Path, ctx),
		"route":         occ.PbMetaGet(occ.Md_Route, ctx),
		"user_agent":    occ.PbMetaGet(occ.Md_UserAgent, ctx),
		"x_request_id":  occ.PbMetaGet(occ.Md_RequestId, ctx),
		"latency":       l.Nanoseconds() / 1000000,
		"latency_human": l.String(),
		"app_header":    occ.PbMetaGet(occ.Md_APPHEADER, ctx),
		"version":       occ.PbMetaGet(occ.Md_Version, ctx),
		"device_id":     occ.PbMetaGet(occ.Md_DEVICEID, ctx),
		"device_type":   occ.PbMetaGet(occ.Md_DEVICETYPE, ctx),
		"user_id":       occ.PbMetaGet(occ.Md_USERID, ctx),
		"bundle_id":     occ.PbMetaGet(occ.Md_Bundle_Id, ctx),
		"app_type":      occ.PbMetaGet(occ.Md_App_Type, ctx),
		"client_type":   occ.PbMetaGet(occ.Md_Client_Type, ctx),
	}
	if err != nil {
		logField["is_error"] = true
		logField["err_message"] = err.Error()
	}
	oc.LogUserActivity(logField, "grpcaccess")

	return resp, err
}

func RecoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
			oc.LogRecover(e)
			err = status.Errorf(codes.Internal, "Panic err: %v", e)
		}
	}()

	return handler(ctx, req)
}

func RecoverInterceptorV2() grpc.UnaryServerInterceptor {
	return grpc_recovery.UnaryServerInterceptor(
		grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			oc.LogRecover(p)
			return oc.ErrInternal
		}),
	)
}

func ClientTimeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := ShrinkDeadline(ctx, timeout)
		defer cancel()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func ServerTimeoutInterceptor(t int64) grpc.UnaryServerInterceptor {
	if t <= 0 {
		t = 5
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx, cancel := ShrinkDeadline(ctx, time.Duration(t)*time.Second)
		defer cancel()
		return handler(ctx, req)
	}
}

func ShrinkDeadline(ctx context.Context, timeout time.Duration) (context.Context, func()) {
	if deadline, ok := ctx.Deadline(); ok {
		leftTime := time.Until(deadline)
		if leftTime > timeout {
			timeout = leftTime
		}
	}
	return context.WithDeadline(ctx, time.Now().Add(timeout))
}
