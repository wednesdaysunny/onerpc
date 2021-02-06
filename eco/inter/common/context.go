package common

import (
	"context"
	"time"

	"gopkg.in/redis.v5"

	"google.golang.org/grpc"

	//"os"
	"regexp"

	gmeta "google.golang.org/grpc/metadata"
)

func BuildTimeoutContext(ctx context.Context, t int64) context.Context {
	timeOutCtx, _ := context.WithTimeout(ctx, time.Duration(t)*time.Second)
	return timeOutCtx
}


func PbMetaGet(k string, ctx context.Context) string {
	md, ok := gmeta.FromIncomingContext(ctx)
	if !ok {
		md, ok = gmeta.FromOutgoingContext(ctx)
		if !ok {
			return ""
		}
	}
	var v string

	for sk, sv := range md {
		if regexp.MustCompile("(?i)^" + k + "$").Match([]byte(sk)) {
			if len(sv) > 0 {
				v = sv[0]
			}
			break
		}
	}
	return v
}