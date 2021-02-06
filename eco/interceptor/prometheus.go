package interceptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	oc "github.com/wednesdaysunny/onerpc/eco/inter"
	"github.com/wednesdaysunny/onerpc/eco/inter/toolkit/strutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	MetricRequestTotal     = "request_total"
	MetricRequestDuration  = "request_duration"
	MetricResponseTotal    = "response_total"
	MetricResponseDuration = "response_duration"

	LabelDestinationApp     = "dst_app"
	LabelDestinationVersion = "dst_version"
	LabelSourceApp          = "src_app"
	LabelSourceVersion      = "src_version"
	LabelApp                = "app"
	LabelVersion            = "version"
	LabelMethod             = "method"
	LabelProtocol           = "protocol"
	LabelInstance           = "instance"
	LabelHostname           = "hostname"
	LabelNamespace          = "namespace"
	LabelResponseStatus     = "response_status"

	GrpcProtocol = "grpc"
	HttpProtocol = "http"
)

const (
	EnvHostnameKey = "HOSTNAME"
	EnvNodeName    = "NODE_NAME"
	UNKNOWN        = "unknown"
)

var (
	Prom        *PromMonitor
	promEnabled = false
	svcName     = UNKNOWN
	svcVersion  = UNKNOWN
	grpcMetrics = grpc_prometheus.NewServerMetrics()
)

type PromMonitor struct {
	RequestTotal     *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	ResponseTotal    *prometheus.CounterVec
	ResponseDuration *prometheus.HistogramVec
	Collectors       []MetricCollector
	Registry         *prometheus.Registry
	Lock             sync.Mutex
}

type MetricCollector struct {
	Collector prometheus.Collector
	JobName   string
}

func init() {
	if strutil.ToBool(os.Getenv("PROMETHEUS_ENABLED")) {
		promEnabled = true
	}
	if val := os.Getenv("SVC_NAME"); val != "" {
		svcName = val
	}
	if val := os.Getenv("SVC_VERSION"); val != "" {
		svcVersion = val
	}
}

func InitPrometheusWithGrpcServer(server *grpc.Server) {
	grpcMetrics.InitializeMetrics(server)
}

func (p *PromMonitor) addCollector(collector MetricCollector) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.Collectors = append(p.Collectors, collector)
}

func (p *PromMonitor) ObserveRequestDuration(labels prometheus.Labels, duration time.Duration) {
	latency := float64(duration.Nanoseconds()) * 1e-9
	p.RequestDuration.With(labels).Observe(latency)
}

func (p *PromMonitor) ObserveResponseDuration(labels prometheus.Labels, duration time.Duration) {
	latency := float64(duration.Nanoseconds()) * 1e-9
	p.ResponseDuration.With(labels).Observe(latency)
}

func (p *PromMonitor) IncrResponseTotal(labels prometheus.Labels) {
	p.ResponseTotal.With(labels).Inc()
}

func (p *PromMonitor) IncrRequestTotal(labels prometheus.Labels) {
	p.RequestTotal.With(labels).Inc()
}

func (p *PromMonitor) StartExporter() {
	defer func() {
		if err := recover(); err != nil {
			oc.LogErrorLn("StartExporter %s panic, err: %v", svcName, err)
		}
	}()
	p.Registry.MustRegister(grpcMetrics)
	for _, v := range p.Collectors {
		p.Registry.MustRegister(v.Collector)
	}
	go func() {
		prometheusExporterAddr := ":9095"
		fmt.Println("prometheus exporter listen", prometheusExporterAddr)
		http.Handle("/metrics", promhttp.HandlerFor(p.Registry, promhttp.HandlerOpts{}))
		log.Fatal(http.ListenAndServe(prometheusExporterAddr, nil))
	}()
}

type MetricLabels struct {
	labels    []string        `json:"labels"`
	labelsMap map[string]bool `json:"labels_map"`
	sync.Mutex
}

func NewMetricLabels() *MetricLabels {
	return &MetricLabels{
		labelsMap: map[string]bool{},
	}
}

func (ml *MetricLabels) SetLabels(labels ...string) {
	ml.Lock()
	defer ml.Unlock()
	for _, l := range labels {
		if ml.labelsMap[l] {
			continue
		}
		ml.labelsMap[l] = true
		ml.labels = append(ml.labels, l)
	}
}

func (ml MetricLabels) GetLabels() []string {
	ml.Lock()
	defer ml.Unlock()
	return ml.labels
}

func (ml MetricLabels) CreatePromLabels(labelsMap map[string]string) (prometheus.Labels, error) {
	promLabels := prometheus.Labels{}
	missingLabels := []string{}
	for _, label := range ml.GetLabels() {
		if v, ok := labelsMap[label]; ok {
			promLabels[label] = v
		} else {
			missingLabels = append(missingLabels, label)
		}
	}
	if len(missingLabels) > 0 {
		return promLabels, errors.New(fmt.Sprintf("labels %s missing in labels map", missingLabels))
	}
	return promLabels, nil
}

var (
	RequestDurationLabels  = NewMetricLabels()
	RequestTotalLabels     = NewMetricLabels()
	ResponseDurationLabels = NewMetricLabels()
	ResponseTotalLabels    = NewMetricLabels()
)

func NewPromMonitor() *PromMonitor {
	prom := &PromMonitor{
		Registry: prometheus.NewRegistry(),
	}

	RequestDurationLabels.SetLabels([]string{LabelNamespace, LabelProtocol, LabelSourceApp, LabelSourceVersion, LabelDestinationApp, LabelDestinationVersion, LabelMethod, LabelResponseStatus}...)
	prom.RequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    MetricRequestDuration,
		Help:    "The grpc request latencies in seconds.",
		Buckets: []float64{0.1, 0.3, 0.5, 1},
	}, RequestDurationLabels.GetLabels())

	RequestTotalLabels.SetLabels([]string{LabelNamespace, LabelProtocol, LabelSourceApp, LabelSourceVersion, LabelDestinationApp, LabelDestinationVersion, LabelMethod, LabelResponseStatus}...)
	prom.RequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: MetricRequestTotal,
		Help: "The grpc request total",
	}, RequestTotalLabels.GetLabels())

	ResponseDurationLabels.SetLabels([]string{LabelNamespace, LabelProtocol, LabelSourceApp, LabelSourceVersion, LabelDestinationApp, LabelDestinationVersion, LabelMethod, LabelResponseStatus}...)
	prom.ResponseDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    MetricResponseDuration,
		Help:    "The grpc response latencies in seconds.",
		Buckets: []float64{0.1, 0.3, 0.5, 1},
	}, ResponseDurationLabels.GetLabels())

	ResponseTotalLabels.SetLabels([]string{LabelNamespace, LabelProtocol, LabelSourceApp, LabelSourceVersion, LabelDestinationApp, LabelDestinationVersion, LabelMethod, LabelResponseStatus}...)
	prom.ResponseTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: MetricResponseTotal,
		Help: "The grpc response total",
	}, ResponseTotalLabels.GetLabels())

	prom.addCollector(MetricCollector{prom.RequestTotal, fmt.Sprintf("%s:%s", svcName, MetricRequestTotal)})
	prom.addCollector(MetricCollector{prom.RequestDuration, fmt.Sprintf("%s:%s", svcName, MetricRequestDuration)})
	prom.addCollector(MetricCollector{prom.ResponseTotal, fmt.Sprintf("%s:%s", svcName, MetricResponseTotal)})
	prom.addCollector(MetricCollector{prom.ResponseDuration, fmt.Sprintf("%s:%s", svcName, MetricResponseDuration)})

	prom.StartExporter()

	return prom
}

type appInfo struct {
	AppName    string `json:"app_name"`
	AppVersion string `json:"app_version"`
}

func GetAppInfoFromClientStream(s grpc.ClientStream) appInfo {
	var (
		info = appInfo{
			AppName:    UNKNOWN,
			AppVersion: UNKNOWN,
		}
	)
	if s == nil {
		return info
	}
	if header, err := s.Header(); err == nil {
		if vals := header.Get(LabelApp); len(vals) > 0 {
			info.AppName = vals[0]
		}
		if vals := header.Get(LabelVersion); len(vals) > 0 {
			info.AppVersion = vals[0]
		}
	}
	return info
}

func GetAppInfoFromMetaData(h metadata.MD) appInfo {
	var (
		info = appInfo{
			AppName:    UNKNOWN,
			AppVersion: UNKNOWN,
		}
	)
	if h == nil {
		return info
	}
	if vals := h.Get(LabelApp); len(vals) > 0 {
		info.AppName = vals[0]
	}
	if vals := h.Get(LabelVersion); len(vals) > 0 {
		info.AppVersion = vals[0]
	}
	return info
}

func GetStatusFromGrpcResponseErr(rpcResponseErr error) string {
	if stat, ok := status.FromError(rpcResponseErr); ok {
		return stat.Code().String()
	}
	return "CUSTOM_ERROR"
}

func GetUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, rsp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !isPrometheusEnabled() {
			return invoker(ctx, method, req, rsp, cc, opts...)
		}
		var (
			header metadata.MD
			begin  = time.Now()
		)

		// sending metadata
		ctx = metadata.AppendToOutgoingContext(ctx, LabelApp, svcName, LabelVersion, svcVersion)
		// add metadata receiver
		opts = append(opts, grpc.Header(&header))
		// invoke
		err := invoker(ctx, method, req, rsp, cc, opts...)
		// receiving metadata
		destAppInfo := GetAppInfoFromMetaData(header)
		responseStatus := GetStatusFromGrpcResponseErr(err)

		defer func() {
			if err := recover(); err != nil {
				oc.LogErrorLn("collector %s panic, err: %v", svcName, err)
			}
		}()
		go func() {
			metricClient := GetPromMonitor()

			labels, err := RequestDurationLabels.CreatePromLabels(map[string]string{
				LabelNamespace:          "one",
				LabelProtocol:           GrpcProtocol,
				LabelSourceApp:          svcName,
				LabelSourceVersion:      svcVersion,
				LabelDestinationApp:     destAppInfo.AppName,
				LabelDestinationVersion: destAppInfo.AppVersion,
				LabelMethod:             method,
				LabelResponseStatus:     responseStatus,
			})
			if err != nil {
				log.Errorln("CreatePromLabels", err)
				return
			}
			metricClient.IncrRequestTotal(labels)
			metricClient.ObserveRequestDuration(labels, time.Since(begin))
		}()
		return err
	}
}

func GetPrometheusStreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

		if !isPrometheusEnabled() {
			return streamer(ctx, desc, cc, method, opts...)
		}
		var (
			begin = time.Now()
		)

		// sending metadata
		ctx = metadata.AppendToOutgoingContext(ctx, LabelApp, svcName, LabelVersion, svcVersion)
		// invoke
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return clientStream, err
		}
		// receiving metadata
		destAppInfo := GetAppInfoFromClientStream(clientStream)
		responseStatus := GetStatusFromGrpcResponseErr(err)

		defer func() {
			if err := recover(); err != nil {
				oc.LogErrorLn("collector %s panic, err: %v", svcName, err)
			}
		}()
		go func() {
			metricClient := GetPromMonitor()

			labels, err := RequestDurationLabels.CreatePromLabels(map[string]string{
				LabelNamespace:          "one",
				LabelProtocol:           GrpcProtocol,
				LabelSourceApp:          svcName,
				LabelSourceVersion:      svcVersion,
				LabelDestinationApp:     destAppInfo.AppName,
				LabelDestinationVersion: destAppInfo.AppVersion,
				LabelMethod:             method,
				LabelResponseStatus:     responseStatus,
			})
			if err != nil {
				log.Errorln("CreatePromLabels", err)
			}
			metricClient.IncrRequestTotal(labels)
			metricClient.ObserveRequestDuration(labels, time.Since(begin))
		}()

		return clientStream, err
	}
}

func GetPrometheusUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !isPrometheusEnabled() {
			return handler(ctx, req)
		}

		var (
			begin  = time.Now()
			header = metadata.Pairs(
				LabelApp, svcName,
				LabelVersion, svcVersion,
			)
		)
		// receiving metadata
		md, _ := metadata.FromIncomingContext(ctx)
		sourceAppInfo := GetAppInfoFromMetaData(md)
		// invoke handler
		resp, err := handler(ctx, req)
		// sending metadata
		grpc.SendHeader(ctx, header)
		responseStatus := GetStatusFromGrpcResponseErr(err)

		defer func() {
			if err := recover(); err != nil {
				oc.LogErrorLn("collector %s panic, err: %v", svcName, err)
			}
		}()
		go func() {
			metricClient := GetPromMonitor()

			labels, err := ResponseDurationLabels.CreatePromLabels(map[string]string{
				LabelNamespace:          "one",
				LabelProtocol:           GrpcProtocol,
				LabelSourceApp:          sourceAppInfo.AppName,
				LabelSourceVersion:      sourceAppInfo.AppVersion,
				LabelDestinationApp:     svcName,
				LabelDestinationVersion: svcVersion,
				LabelMethod:             info.FullMethod,
				LabelResponseStatus:     responseStatus,
			})
			if err != nil {
				log.Errorln("CreatePromLabels", err)
				return
			}
			metricClient.IncrResponseTotal(labels)
			metricClient.ObserveResponseDuration(labels, time.Since(begin))
		}()
		return resp, err
	}
}

func GetPrometheusStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(src interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		if !isPrometheusEnabled() {
			return handler(src, ss)
		}

		var (
			begin  = time.Now()
			header = metadata.Pairs(
				LabelApp, svcName,
				LabelVersion, svcVersion,
			)
		)
		// receiving metadata
		md, _ := metadata.FromIncomingContext(ss.Context())
		sourceAppInfo := GetAppInfoFromMetaData(md)
		// sending metadata
		grpc.SendHeader(ss.Context(), header)
		// invoke handler
		err := handler(src, ss)

		responseStatus := GetStatusFromGrpcResponseErr(err)

		defer func() {
			if err := recover(); err != nil {
				oc.LogErrorLn("collector %s panic, err: %v", svcName, err)
			}
		}()
		go func() {
			metricClient := GetPromMonitor()
			labels, err := RequestTotalLabels.CreatePromLabels(map[string]string{
				LabelNamespace:          "one",
				LabelProtocol:           GrpcProtocol,
				LabelSourceApp:          sourceAppInfo.AppName,
				LabelSourceVersion:      sourceAppInfo.AppVersion,
				LabelDestinationApp:     svcName,
				LabelDestinationVersion: svcVersion,
				LabelMethod:             info.FullMethod,
				LabelResponseStatus:     responseStatus,
			})
			if err != nil {
				log.Errorln("CreatePromLabels", err)
				return
			}
			metricClient.IncrResponseTotal(labels)
			metricClient.ObserveResponseDuration(labels, time.Since(begin))
		}()
		return err
	}
}

func GetPrometheusClientInterceptors() ([]grpc.UnaryClientInterceptor, []grpc.StreamClientInterceptor) {
	if !isPrometheusEnabled() {
		return []grpc.UnaryClientInterceptor{}, []grpc.StreamClientInterceptor{}
	}
	unarys := []grpc.UnaryClientInterceptor{
		GetUnaryClientInterceptor(),
	}

	streams := []grpc.StreamClientInterceptor{
		GetPrometheusStreamClientInterceptor(),
	}

	return unarys, streams
}

func GetPrometheusServerInterceptors() ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	if !isPrometheusEnabled() {
		return []grpc.UnaryServerInterceptor{}, []grpc.StreamServerInterceptor{}
	}
	unarys := []grpc.UnaryServerInterceptor{
		grpcMetrics.UnaryServerInterceptor(),
		GetPrometheusUnaryServerInterceptor(),
	}

	streams := []grpc.StreamServerInterceptor{
		grpcMetrics.StreamServerInterceptor(),
		GetPrometheusStreamServerInterceptor(),
	}

	return unarys, streams
}

var l sync.Mutex

func GetPromMonitor() *PromMonitor {
	if Prom == nil {
		// to create a monitor
		l.Lock()
		defer l.Unlock()
		// if other goroutine has created
		if Prom == nil {
			Prom = NewPromMonitor()
		}
	}
	return Prom
}

func isPrometheusEnabled() bool {
	return promEnabled
}
