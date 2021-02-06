package interceptor

import (
	"context"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	std "github.com/wednesdaysunny/onerpc/eco/inter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
)

//MDReaderWriter metadata Reader and Writer
type MDReaderWriter struct {
	metadata.MD
}

// ForeachKey implements ForeachKey of opentracing.TextMapReader
func (c MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range c.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// Set implements Set() of opentracing.TextMapWriter
func (c MDReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	c.MD[key] = append(c.MD[key], val)
}

// NewJaegerTracer NewJaegerTracer for current service
func NewJaegerTracer(serviceName string) {
	jcfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
			// 将span发往jaeger-collector的服务地址
			//CollectorEndpoint: "http://localhost:14268/api/traces",
			CollectorEndpoint: "http://jaeger-collector.istio-system:14268/api/traces",
		},
	}

	tracer, _, err := jcfg.NewTracer()
	if err != nil {
		std.LogErrorLn("NewJaegerTracer start ", err, serviceName)
	}

	opentracing.SetGlobalTracer(tracer)
	std.LogInfoLn("NewJaegerTracer start....", serviceName)
}

// ClientInterceptor grpc client wrapper
func ClientInterceptor() grpc.UnaryClientInterceptor {
	if !opentracing.IsGlobalTracerRegistered() {
		return nil
	}
	return func(ctx context.Context, method string,
		req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		var parentCtx opentracing.SpanContext
		parentSpan := opentracing.SpanFromContext(ctx)
		if parentSpan != nil {
			parentCtx = parentSpan.Context()
		}

		span := opentracing.GlobalTracer().StartSpan(
			method,
			opentracing.ChildOf(parentCtx),
			opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
			ext.SpanKindRPCClient,
		)
		defer span.Finish()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		mdWriter := MDReaderWriter{md}
		err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, mdWriter)
		if err != nil {
			span.LogFields(log.String("inject-error", err.Error()))
		}

		newCtx := metadata.NewOutgoingContext(ctx, md)
		err = invoker(newCtx, method, req, reply, cc, opts...)
		if err != nil {
			span.LogFields(log.String("call-error", err.Error()))
		}
		return err
	}
}

func IsGlobalTracerRegistered(serviceName string) bool {
	if !opentracing.IsGlobalTracerRegistered() {
		NewJaegerTracer(serviceName)
		if !opentracing.IsGlobalTracerRegistered() {
			return false
		}
	}
	return true
}

// ServerInterceptor grpc server wrapper
func ServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	if !IsGlobalTracerRegistered(serviceName) {
		return nil
	}

	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		spanContext, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, MDReaderWriter{md})
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			grpclog.Errorf("extract from metadata err: %v", err)
		} else {
			span := opentracing.GlobalTracer().StartSpan(
				info.FullMethod,
				ext.RPCServerOption(spanContext),
				opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
				ext.SpanKindRPCServer,
			)
			defer span.Finish()

			ctx = opentracing.ContextWithSpan(ctx, span)
		}

		return handler(ctx, req)
	}
}
