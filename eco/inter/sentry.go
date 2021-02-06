package onecommon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/wednesdaysunny/onerpc/eco/inter/conf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func init() {
	initSentry()
}

func initSentry() {
	svcName := "unknown_service"
	svcRelease := "unknown_release"
	if val := conf.ConfSvcName(); val != "" {
		svcName = val
	}
	dsn := os.Getenv(fmt.Sprintf("%s_DSN", strings.ToUpper(svcName)))
	if dsn == "" {
		LogInfoLn("sentry disabled!")
		return
	}
	// SVC_VERSION in istio pod like `210123t105737-e63c8f99`
	if val := conf.ConfSvcVersion(); val != "" {
		items := strings.Split(val, "-")
		if len(items) == 2 {
			svcRelease = items[1]
		} else {
			svcRelease = val
		}
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		AttachStacktrace: true,
		ServerName:       svcName,
		Release:          svcRelease,
		Environment:      conf.ConfEnv(),
	})
	if err != nil {
		LogFatalLn(fmt.Sprintf("sentry.Init: %s", err))
		return
	}
	LogInfoLn("sentry enabled!")
}

func GenHttpRequestFromGrpcContext(ctx context.Context) *http.Request {
	var (
		method = ""
		header = http.Header{}
	)
	if mtd, ok := grpc.Method(ctx); ok {
		method = mtd
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, vals := range md {
			for _, v := range vals {
				header.Add(k, v)
			}
		}
	}
	return &http.Request{
		Method: "POST",
		URL: &url.URL{
			Path: method,
		},
		ProtoMajor: 2,
		Header:     header,
		RequestURI: method,
	}
}

func MarshalGrpcReq(req interface{}) []byte {
	var body []byte
	if data, err := json.Marshal(req); err == nil {
		body = data
	}
	return body
}

// RecoverRepanicWithSentry sends error captured in goroutine to sentry
func RecoverRepanicWithSentry(ctx context.Context, req interface{}) {
	if x := recover(); x != nil {
		hub := sentry.CurrentHub().Clone()
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetRequest(GenHttpRequestFromGrpcContext(ctx))
			scope.SetRequestBody(MarshalGrpcReq(req))
		})
		hub.RecoverWithContext(ctx, x)
		hub.Flush(5 * time.Second)
		panic(x)
	}
}

// CaptureExceptionWithSentry captures error and sends to sentry
func CaptureExceptionWithSentry(excp error) {
	sentry.CaptureException(excp)
}
