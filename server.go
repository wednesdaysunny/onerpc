package onerpc

import (
	"github.com/wednesdaysunny/onerpc/eco"
	"github.com/wednesdaysunny/onerpc/eco/interceptor"
	"log"
	"os"
	"strings"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	oc "github.com/wednesdaysunny/onerpc/eco/inter"
	oconf "github.com/wednesdaysunny/onerpc/eco/inter/conf"
	"github.com/wednesdaysunny/onerpc/eco/inter/toolkit/netutil"
	"google.golang.org/grpc"
)

const (
	envPodIp       = "POD_IP"
	defaultMsgSize = 20971520 //20M
)

type RpcServer struct {
	server   eco.Server
	register eco.RegisterFn
}

func MustNewServer(c oconf.RpcServerConf, register eco.RegisterFn) *RpcServer {
	{
		interceptor.InitCache(c.RpcCacheRedis)
		interceptor.InitJaeger(oconf.GenServiceName(c.Name))
	}
	server, err := NewServer(c, register)
	if err != nil {
		log.Fatal(err)
	}

	interceptor.InitPrometheusWithGrpcServer(server.GetGrpcServer())

	return server
}

func NewServer(c oconf.RpcServerConf, register eco.RegisterFn) (*RpcServer, error) {
	var (
		err    error
		server eco.Server
	)

	server = eco.NewRpcServer(c.ListenOn)

	server.SetName(c.Name)
	interceptors, err := BuildInterceptors(c)
	if err != nil {
		return nil, err
	}

	rpcServer := &RpcServer{
		server:   server,
		register: register,
	}

	{
		grpcServer := grpc.NewServer(interceptors...)
		rpcServer.SetGrpcServer(grpcServer)
	}

	return rpcServer, nil
}

func (rs *RpcServer) AddOptions(options ...grpc.ServerOption) {
	rs.server.AddOptions(options...)
}

func (rs *RpcServer) AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	rs.server.AddStreamInterceptors(interceptors...)
}

func (rs *RpcServer) AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	rs.server.AddUnaryInterceptors(interceptors...)
}

func (rs *RpcServer) GetGrpcServer() *grpc.Server {
	return rs.server.GetGrpcServer()
}

func (rs *RpcServer) SetGrpcServer(sv *grpc.Server) {
	rs.server.SetGrpcServer(sv)
}

func (rs *RpcServer) Start() {
	if err := rs.server.Start(rs.register); err != nil {
		oc.LogErrorLn(err)
		panic(err)
	}
}

func (rs *RpcServer) Stop() {

}

func figureOutListenOn(listenOn string) string {
	fields := strings.Split(listenOn, ":")
	if len(fields) == 0 {
		return listenOn
	}

	host := fields[0]
	if len(host) > 0 && host != "0.0.0.0" {
		return listenOn
	}

	ip := os.Getenv(envPodIp)
	if len(ip) == 0 {
		ip = netutil.InternalIp()
	}
	if len(ip) == 0 {
		return listenOn
	} else {
		return strings.Join(append([]string{ip}, fields[1:]...), ":")
	}
}

func BuildInterceptors(c oconf.RpcServerConf) ([]grpc.ServerOption, error) {

	var (
		unary   []grpc.UnaryServerInterceptor
		streams []grpc.StreamServerInterceptor
	)
	{
		unary = append(unary,
			interceptor.RecoverInterceptorV2(),
			interceptor.LoggingInterceptor,
		)
		streams = append(streams, grpcrecovery.StreamServerInterceptor(grpcrecovery.WithRecoveryHandler(func(p interface{}) (err error) {
			oc.LogRecover(p)
			return oc.ErrInternal
		})))
	}
	{
		mUnary, mStream := interceptor.GetPrometheusServerInterceptors()
		if len(mUnary) > 0 && len(mStream) > 0 {
			unary = append(unary, mUnary...)
			streams = append(streams, mStream...)
		}

		if tUnary := interceptor.ServerInterceptor(oconf.GenServiceName(c.Name)); tUnary != nil {
			unary = append(unary, tUnary)
		}
	}
	{
		sentryUnaryInterceptor, sentryStreamInterceptor := interceptor.GetSentryServerInterceptors()
		if len(sentryUnaryInterceptor) > 0 || len(sentryStreamInterceptor) > 0 {
			unary = append(unary, sentryUnaryInterceptor...)
			streams = append(streams, sentryStreamInterceptor...)
		}
	}
	unary = append(unary, interceptor.CacheUnaryServerInterceptor())
	options := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(defaultMsgSize),
		grpc.MaxSendMsgSize(defaultMsgSize),
		grpcmiddleware.WithUnaryServerChain(unary...),
		grpcmiddleware.WithStreamServerChain(streams...),
	}

	return options, nil
}
