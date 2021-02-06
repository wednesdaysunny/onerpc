package interceptor

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/wednesdaysunny/onerpc/eco/inter/toolkit"

	"github.com/gogo/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	stdc "github.com/wednesdaysunny/onerpc/eco/inter/common"
	cache "gopkg.in/go-redis/cache.v5"
	redis "gopkg.in/redis.v5"

	std "github.com/wednesdaysunny/onerpc/eco/inter"
	oconf "github.com/wednesdaysunny/onerpc/eco/inter/conf"
	ocr "github.com/wednesdaysunny/onerpc/eco/inter/toolkit/reflect"
	"google.golang.org/grpc"
)

var (
	CacheMgrIns             *CacheManager
	cacheLock               sync.Mutex
	json                    = jsoniter.ConfigCompatibleWithStandardLibrary
	customServerInterceptor grpc.UnaryServerInterceptor
)

type RespFunc func() interface{}

type CacheSetting struct {
	Method     interface{}
	Expiration time.Duration
	AnonOnly   bool
	//Resp       interface{}
	InitFunc RespFunc
}

type CacheManager struct {
	enabled bool
	codec   *cache.Codec
	config  map[string]CacheSetting
}

type cachedObj struct {
	Data   []byte
	IvkErr *std.Err
}

func InitCache(conf oconf.RpcCacheRedisConf) {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	if CacheMgrIns != nil {
		return
	}

	if !conf.Enabled {
		CacheMgrIns = &CacheManager{}
	} else {
		marshal := func(v interface{}) ([]byte, error) {
			if _, ok := v.(*cachedObj); !ok {
				return nil, std.ErrRpcCacheMarshal
			} else if b, err := json.Marshal(v); err != nil {
				return nil, std.ErrRpcCacheMarshal
			} else {
				return b, nil
			}
		}
		unmarshal := func(b []byte, v interface{}) error {
			if _, ok := v.(*cachedObj); !ok {
				return std.ErrRpcCacheUnmarshal
			} else if err := json.Unmarshal(b, v); err != nil {
				return std.ErrRpcCacheUnmarshal
			} else {
				return nil
			}
		}
		var codec = &cache.Codec{
			Marshal:   marshal,
			Unmarshal: unmarshal,
		}
		if conf.RedisType == "cluster" {
			var addrs []string
			for _, v := range conf.Addrs {
				addrs = append(addrs, v)
			}
			rc := redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:       addrs,
				Password:    conf.Password,
				IdleTimeout: time.Second * time.Duration(conf.IdleTimeout),
			})
			codec.Redis = rc
		} else {
			rc := redis.NewRing(&redis.RingOptions{
				Addrs:       conf.Addrs,
				Password:    conf.Password,
				IdleTimeout: time.Second * time.Duration(conf.IdleTimeout),
			})
			codec.Redis = rc
		}

		CacheMgrIns = &CacheManager{
			enabled: true,
			codec:   codec,
		}
	}
}

func ConfigRpcCache(settings []CacheSetting) {
	if CacheMgrIns == nil {
		std.LogErrorLn("Call initCache first to initialize")
		return
	}

	conf := make(map[string]CacheSetting)
	for _, setting := range settings {
		if n := ocr.PbMethodName(setting.Method); n != "" {
			conf[n] = setting
		}
	}

	CacheMgrIns.config = conf
}

type ret struct {
	obj interface{}
	err error
}

// CacheUnaryServerInterceptor returns a new unary server interceptor for server cache.
func CacheUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	if customServerInterceptor != nil {
		return customServerInterceptor
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (rsp interface{}, err error) {
		settingKey := settingName(info.FullMethod)
		if !CacheMgrIns.enabled || settingKey == "" {
			return handler(ctx, req)
		}
		settings, ok := CacheMgrIns.config[settingKey]
		if !ok || settings.Expiration <= 0 {
			return handler(ctx, req)
		}
		if settings.AnonOnly {
			userID := stdc.PbGetUser(ctx)
			if userID > 0 {
				return handler(ctx, req)
			}
		}
		if key, err := getCacheKey(ctx, settingKey, req); err != nil {
			// NOTE: we probably should NOT go ahead and call the logic
			// because that could cause unbearable traffic
			std.LogErrorc("redis", err, "fail to fetch rpc content from cache")
			return nil, err
		} else {
			v, err := CacheMgrIns.codec.Do(&cache.Item{
				Key:        key,
				Object:     new(cachedObj), // destination
				Expiration: settings.Expiration,
				Func: func() (interface{}, error) {
					retChan := make(chan ret, 1)
					go func() {
						defer func() {
							if e := recover(); e != nil {
								std.LogRecover(e)
							}
						}()

						rsp, err = handler(ctx, req)
						retChan <- ret{rsp, err}
					}()

					timeout := settings.Expiration / 2
					select {
					case ret := <-retChan:
						var cobj cachedObj
						if ret.err != nil {
							cobj.IvkErr = std.ErrFromGoErr(ret.err)
							if std.IsIvankaErr(cobj.IvkErr, std.ErrInternalFromString) {
								return ret.obj, ret.err
							}
						} else {
							cobj.Data, err = toolkit.MarshalResp(ret.obj)
							if err != nil {
								return nil, err
							}
						}
						return &cobj, nil
					case <-time.After(timeout):
						std.LogErrorc("rpc", nil, fmt.Sprintf("fail to call rpc %s: timeout", settingKey))
						return nil, std.ErrRpcCacheTimeout
					}
				},
			})
			if err != nil {
				if std.IsIvankaErr(err, std.ErrRpcCacheTimeout) {
					return nil, err
				} else {
					std.LogErrorc("rpc", err, "fail to call rpc")
					return nil, std.ErrRpcCache
				}
			} else {
				if cobj, ok := v.(*cachedObj); !ok {
					std.LogErrorc("rpc", nil, "rpc cache: invalid return type")
					return nil, std.ErrRpcCache
				} else if cobj.IvkErr != nil {
					return nil, cobj.IvkErr
				} else {
					response := settings.InitFunc()

					if err := toolkit.UnmarshalResp(cobj.Data, response); err != nil {
						std.LogErrorc("rpc", err, "fail to unmarshal response")
						return nil, err
					}
					return response, nil
				}
			}
		}

		return handler(ctx, req)
	}
}

func settingName(service string) string {
	// /package.service/method -> service.method
	dot := strings.Index(service, ".")
	if dot < 0 {
		return ""
	}
	return strings.Replace(service[dot+1:], "/", ".", 1)
}

func getCacheKey(ctx context.Context, methodName string, req interface{}) (string, error) {
	if msg, ok := req.(proto.Message); !ok {
		return "", std.ErrRpcCacheMarshal
	} else if b, err := proto.Marshal(msg); err != nil {
		return "", std.ErrRpcCacheMarshal
	} else {
		platform := stdc.PbGetPlatform(ctx)
		//return "rpcc:" + methodName + ":" + base64.RawURLEncoding.EncodeToString(b), nil
		return "grpc:" + methodName + ":" + base64.RawURLEncoding.EncodeToString(b) + ":" + platform, nil
	}
}
