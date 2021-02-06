package toolkit

import (
	"fmt"

	"strconv"
	"time"

	oconf "github.com/wednesdaysunny/onerpc/eco/inter/conf"
	"gopkg.in/redis.v5"
)

// InitRedis initializes the redis connection.
// Usage:
// client := InitRedis(config)
func InitRedis(config oconf.RedisConf) *redis.Client {

	hostAddress := fmt.Sprintf("%s:%d", config.Host, config.Port)
	_client := redis.NewClient(&redis.Options{
		Addr:     hostAddress,
		Password: config.Auth, // no password set
		DB:       0,           // use default DB
		PoolSize: 100,
	})
	if _, err := _client.Ping().Result(); err != nil {
		fmt.Println("Failed to connect Redis.", err)
		_client = nil
	}

	return _client
}

func InitRedisRing(config ...oconf.RedisConf) *redis.Ring {

	servers := map[string]string{}
	var pass string
	var timeoutRead, timeoutWrite, timeoutConnect, timeoutIdle time.Duration

	for i, v := range config {
		servers[strconv.Itoa(i)] = fmt.Sprintf("%s:%d", v.Host, v.Port)
		if i == 0 {
			pass = v.Auth
			timeoutIdle = time.Second * time.Duration(v.IdleTimeout)
		}
	}
	client := redis.NewRing(&redis.RingOptions{
		Addrs:        servers,
		Password:     pass,
		IdleTimeout:  timeoutIdle,
		ReadTimeout:  timeoutRead,
		WriteTimeout: timeoutWrite,
		DialTimeout:  timeoutConnect,
		PoolSize:     200,
	})

	if _, err := client.Ping().Result(); err != nil {
		fmt.Println("Failed to connect Redis.")
		client = nil
	}

	return client
}
