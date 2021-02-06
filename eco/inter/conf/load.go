package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	redis "gopkg.in/redis.v5"
	"gopkg.in/yaml.v2"
)

const (
	RedisConfigKey = "REDIS_CONFIG"
	ConfRedisAddr  = "REDIS_ADDR"
	ConfRedisPwd   = "REDIS_PWD"
	SvcVersionKey  = "SVC_VERSION"
	HostnameKey    = "HOSTNAME"
)

var loaders = map[string]func([]byte, interface{}) error{
	".json": LoadConfigFromJsonBytes,
	".yaml": LoadConfigFromYamlBytes,
	".yml":  LoadConfigFromYamlBytes,
}

func LoadConfig(file string, v interface{}) error {
	if content, err := ioutil.ReadFile(file); err != nil {
		return err
	} else if loader, ok := loaders[path.Ext(file)]; ok {
		return loader(content, v)
	} else {
		return fmt.Errorf("unrecoginized file type: %s", file)
	}
}

func LoadConfigFromJsonBytes(content []byte, v interface{}) error {
	return json.Unmarshal(content, v)
}

func LoadConfigFromYamlBytes(content []byte, v interface{}) error {
	return yaml.Unmarshal(content, v)
}

func MustLoad(path string, v interface{}) {
	if err := LoadConfig(path, v); err != nil {
		log.Fatalf("error: config file %s, %s", path, err.Error())
	}

}

func loadFromRedis(dest interface{}, path string) error {
	addr, pwd := getRedisAddrPwd()
	if addr == "" || pwd == "" {
		panic("failed to get config redis")
	}
	cli := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
	})

	if err := cli.Ping().Err(); err != nil {
		panic("failed to ping redis")
	}
	pathArr := strings.Split(path, "/")
	if len(pathArr) != 2 {
		panic("failed to get config file name")
	}
	res, err := cli.Get(pathArr[1]).Result()
	if err != nil {
		panic("failed to get config file data from redis")
	}

	if len(res) == 0 {
		panic("failed to get config file data from redis, data is empty")
	}

	confRendered, err := renderConfig(res)

	if err != nil {
		panic("failed to renderConfig config data "+err.Error())
	}

	return LoadConfigFromYamlBytes([]byte(confRendered), dest)
}

func getRedisAddrPwd() (string, string) {

	addr := os.Getenv(ConfRedisAddr)
	pwd := os.Getenv(ConfRedisPwd)
	return addr, pwd
}
