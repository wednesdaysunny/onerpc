package conf

import "fmt"

var (
	ProdEnv = "prod"
	SitEnv  = "sit"
	TestEnv = "test"
	RpcPort = 10010
)

func GenServiceAddr(env, svc string) string {
	port := GenServicePort(svc)
	return fmt.Sprintf(":%d", port)
}

func GenServiceName(sn string) string {
	env := ConfEnv()
	if env == "test" || env == "prod" {
		return sn
	}
	return "127.0.0.1"
}

func GenServicePort(svc string) int {
	env := ConfEnv()
	if env == ProdEnv || env == TestEnv {
		return RpcPort
	}

	// debug in local, generate service port from 20000
	// 参照前p47项目，通过byte转ascii值相加生成port，特殊情况下依然会存在冲突，可通过修改map中的value解决
	port := 0
	for _, bv := range []byte(svc + svc) {
		port += int(bv)
	}
	return port + 20000
}
