package onerpc

import (
	"fmt"
	"github.com/wednesdaysunny/onerpc/eco"
	"log"
	"time"

	oconf "github.com/wednesdaysunny/onerpc/eco/inter/conf"
	"google.golang.org/grpc"
)

type (
	ClientOption = eco.ClientOption

	Client interface {
		Conn() *grpc.ClientConn
	}

	RpcClient struct {
		client Client
	}
)

func MustNewClient(c oconf.RpcClientConf, options ...ClientOption) Client {
	cli, err := NewClient(c, options...)
	if err != nil {
		log.Fatal(err)
	}

	return cli
}

func NewClient(c oconf.RpcClientConf, options ...eco.ClientOption) (Client, error) {
	var opts []ClientOption
	if c.Timeout > 0 {
		opts = append(opts, eco.WithTimeout(time.Duration(c.Timeout)*time.Second))
	}
	if c.PollSize > 0 {
		opts = append(opts, eco.WithPollSize(int(c.PollSize)))
	}

	opts = append(opts, options...)

	var (
		client      Client
		err         error
		serviceName = oconf.GenServiceName(c.Name)
	)
	if serviceName != "" {
		port := oconf.GenServicePort(c.Name)
		target := fmt.Sprintf("%s:%d", serviceName, port)
		client, err = eco.NewClient(target, opts...)
	} else if len(c.Endpoints) > 0 {
		client, err = eco.NewClient(eco.BuildDirectTarget(c.Endpoints), opts...)
	}
	if err != nil {
		return nil, err
	}

	return &RpcClient{
		client: client,
	}, nil
}

func NewClientWithTarget(target string, opts ...ClientOption) (Client, error) {
	return eco.NewClient(target, opts...)
}

func (rc *RpcClient) Conn() *grpc.ClientConn {
	return rc.client.Conn()
}

// 端到端client
func NewDirectClientConf(endpoints []string, app, token string) oconf.RpcClientConf {
	return oconf.RpcClientConf{
		Endpoints: endpoints,
		App:       app,
		Token:     token,
	}
}

// istio 服务发现
func NewIstioClientConf(name string) oconf.RpcClientConf {
	return oconf.RpcClientConf{}
}
