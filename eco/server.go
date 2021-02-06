package eco

import (
	"google.golang.org/grpc"
)

type (
	RegisterFn func(*grpc.Server)

	Server interface {
		AddOptions(options ...grpc.ServerOption)
		AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor)
		AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor)
		SetName(string)
		Start(register RegisterFn) error
		SetGrpcServer(s *grpc.Server)
		GetGrpcServer() *grpc.Server
	}

	baseRpcServer struct {
		address            string
		options            []grpc.ServerOption
		streamInterceptors []grpc.StreamServerInterceptor
		unaryInterceptors  []grpc.UnaryServerInterceptor
		grpc               *grpc.Server
	}
)

func newBaseRpcServer(address string) *baseRpcServer {
	return &baseRpcServer{
		address: address,
	}
}

func (s *baseRpcServer) AddOptions(options ...grpc.ServerOption) {
	s.options = append(s.options, options...)
}

func (s *baseRpcServer) AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	s.streamInterceptors = append(s.streamInterceptors, interceptors...)
}

func (s *baseRpcServer) AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	s.unaryInterceptors = append(s.unaryInterceptors, interceptors...)
}

func (s *baseRpcServer) SetGrpcServer(sv *grpc.Server) {
	s.grpc = sv
}

func (s *baseRpcServer) GetGrpcServer() *grpc.Server {
	return s.grpc
}
