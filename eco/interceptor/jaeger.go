package interceptor

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/uber/jaeger-client-go"
	"log"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"google.golang.org/grpc"
)

const (
	TraceFieldTagName = "trace_field"
)

var (
	Jtracer opentracing.Tracer
)

func InitJaeger(svcName string) {
	return
	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		log.Fatal(err)
		return
	}
	metricsFactory := prometheus.New()
	cfg.ServiceName = svcName
	cfg.Headers = &jaeger.HeadersConfig{
		TraceContextHeaderName: "x-request-id",
	}
	tracer, _, err := cfg.NewTracer(config.Metrics(metricsFactory))
	if err != nil {
		log.Fatal(err)
		return
	}
	Jtracer = tracer
	log.Println("InitJaeger succ")
}
func GetJaegerClientInterceptors() ([]grpc.UnaryClientInterceptor, []grpc.StreamClientInterceptor) {
	if Jtracer == nil {
		return nil, nil
	}
	opts := getTracingOptions()
	return []grpc.UnaryClientInterceptor{grpc_opentracing.UnaryClientInterceptor(opts...)},
		[]grpc.StreamClientInterceptor{grpc_opentracing.StreamClientInterceptor(opts...)}
}

func GetJaegerServerInterceptors() ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	if Jtracer == nil {
		return nil, nil
	}
	opts := getTracingOptions()
	return []grpc.UnaryServerInterceptor{
			/*grpc_ctxtags.UnaryServerInterceptor(
			grpc_ctxtags.WithFieldExtractor(
				grpc_ctxtags.TagBasedRequestFieldExtractor(TraceFieldTagName))),*/
			grpc_opentracing.UnaryServerInterceptor(opts...)},
		[]grpc.StreamServerInterceptor{
			/*grpc_ctxtags.StreamServerInterceptor(
			grpc_ctxtags.WithFieldExtractor(
				grpc_ctxtags.TagBasedRequestFieldExtractor(TraceFieldTagName))),*/
			grpc_opentracing.StreamServerInterceptor(opts...)}
}

func getTracingOptions() []grpc_opentracing.Option {
	return []grpc_opentracing.Option{
		grpc_opentracing.WithTracer(Jtracer)}
}
