package eco

import (
	"net"
)

type (
	ServerOption func(options *rpcServerOptions)

	rpcServerOptions struct {
		// TODO add opions
	}

	rpcServer struct {
		name string
		*baseRpcServer
	}
)

func NewRpcServer(address string, opts ...ServerOption) Server {
	var options rpcServerOptions
	for _, opt := range opts {
		opt(&options)
	}

	return &rpcServer{
		baseRpcServer: newBaseRpcServer(address),
	}
}

func (s *rpcServer) SetName(name string) {
	s.name = name
}

func (s *rpcServer) Start(register RegisterFn) error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	register(s.GetGrpcServer())
	// we need to make sure all others are wrapped up
	// so we do graceful stop at shutdown phase instead of wrap up phase
	// TODO grace stop
	shutdownCalled := AddShutdownListener(func() {
		s.GetGrpcServer().GracefulStop()
	})
	err = s.GetGrpcServer().Serve(lis)
	shutdownCalled()

	return err
}
