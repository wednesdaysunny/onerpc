package eco

import (
	"context"
	"errors"
	"fmt"
	"github.com/wednesdaysunny/onerpc/eco/interceptor"
	"strings"
	"time"


	"google.golang.org/grpc"
)

const (
	dialTimeout = time.Second * 3
	separator   = '/'
)

type (
	ClientOptions struct {
		PoolSize    int
		Timeout     time.Duration
		DialOptions []grpc.DialOption
	}

	ClientOption func(options *ClientOptions)

	client struct {
		conn *grpc.ClientConn
	}
)

func NewClient(target string, opts ...ClientOption) (*client, error) {
	var cli client
	if err := cli.dial(target, opts...); err != nil {
		return nil, err
	}

	return &cli, nil
}

func (c *client) Conn() *grpc.ClientConn {
	return c.conn
}

func (c *client) buildDialOptions(opts ...ClientOption) []grpc.DialOption {
	var cliOpts ClientOptions
	for _, opt := range opts {
		opt(&cliOpts)
	}

	options := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	{
		var (
			unary   []grpc.UnaryClientInterceptor
			streams []grpc.StreamClientInterceptor
		)
		if cliOpts.Timeout > 0 {
			unary = append(unary, interceptor.ClientTimeoutInterceptor(cliOpts.Timeout))
		}
		promUnary, promStream := interceptor.GetPrometheusClientInterceptors()
		if len(promUnary) > 0 && len(promStream) > 0 {
			unary = append(unary, promUnary...)
			streams = append(streams, promStream...)
		}

		if traceUnary := interceptor.ClientInterceptor(); traceUnary != nil {
			unary = append(unary, traceUnary)
		}

		if len(unary) > 0 {
			options = append(options, grpc.WithChainUnaryInterceptor(unary...))
		}
		if len(streams) > 0 {
			options = append(options, grpc.WithChainStreamInterceptor(streams...))
		}
	}

	return append(options, cliOpts.DialOptions...)
}

func (c *client) dial(server string, opts ...ClientOption) error {
	options := c.buildDialOptions(opts...)
	timeCtx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()
	conn, err := grpc.DialContext(timeCtx, server, options...)
	if err != nil {
		service := server
		if errors.Is(err, context.DeadlineExceeded) {
			pos := strings.LastIndexByte(server, separator)
			// len(server) - 1 is the index of last char
			if 0 < pos && pos < len(server)-1 {
				service = server[pos+1:]
			}
		}
		return fmt.Errorf("rpc dial: %s, error: %s, make sure rpc service %q is alread started",
			server, err.Error(), service)
	}

	c.conn = conn
	return nil
}

func WithDialOption(opt grpc.DialOption) ClientOption {
	return func(options *ClientOptions) {
		options.DialOptions = append(options.DialOptions, opt)
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(options *ClientOptions) {
		options.Timeout = timeout
	}
}

func WithPollSize(size int) ClientOption {
	return func(options *ClientOptions) {
		options.PoolSize = size
	}
}
