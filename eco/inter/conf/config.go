package conf

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const (
	ConfEnvDevelopment = "development"
	ConfEnvTest        = "ivktest"
	ConfEnvProduction  = "ivkprod"
	ConfEnvStage       = "ivkstage"
	ConfPushEnvIOS     = "ivkios"
	ConfPushEnvAndroid = "ivkandroid"
)

type (
	RedisConf struct {
		Host        string `yaml:"host"`
		Port        int    `yaml:"port"`
		Auth        string `yaml:"auth"`
		IdleTimeout int    `yaml:"idle_timeout"`
	}

	MysqlConf struct {
		Username       string `yaml:"username"`
		Password       string `yaml:"password"`
		Host           string `yaml:"host"`
		Port           int    `yaml:"port"`
		DBName         string `yaml:"db_name"`
		MaxIdle        int    `yaml:"max_idle"`
		MaxConn        int    `yaml:"max_conn"`
		LogType        string `yaml:"log_type"`
		ShowLog        bool   `yaml:"show_log"`
		NotCreateTable bool   `yaml:"not_create_table"`
		AutoMerge      bool   `yaml:"auto_merge"`
		Charset        string `yaml:"charset"`
	}

	NsqConsumerConf struct {
		Enable        bool
		LookupAddress []string `yaml:"lookup_address_list"`
	}

	NsqProducerConf struct {
		Enable      bool
		NsqdAddress string `yaml:"addr"`
	}

	EsConf struct {
		EsUrls string `yaml:"es_urls"`
	}

	COSConf struct {
		BucketUrl string `yaml:"bucket_url" json:"bucket_url"`
		SecretID  string `yaml:"secret_id" json:"secret_id"`
		SecretKey string `yaml:"secret_key" json:"secret_key"`
	}
	ConfigLog struct {
		Level      int    `yaml:"level"`
		SentryDSN  string `yaml:"sentry_dsn"`
		Path       string `yaml:"path"`
		OutputDest string `yaml:"output_dest"`
	}

	// ConfigRpcCacheRedis sets the RPC cache backend by Redis
	RpcCacheRedisConf struct {
		RedisType   string            `yaml:"redis_type"` // cluster or ring, default is ring
		Enabled     bool              `yaml:"enabled"`
		Addrs       map[string]string `yaml:"addrs"`
		Password    string            `yaml:"password"`
		IdleTimeout int               `yaml:"idle_timeout"`
	}

	EventConf struct {
		Enabled  bool     `yaml:"enabled" json:"enabled"`
		Type     string   `yaml:"type" json:"type"`
		Addrs    []string `yaml:"addrs" json:"addrs"`
		Topic    string   `yaml:"topic" json:"topic"`
		Encoding string   `yaml:"encoding" json:"encoding"`

		Compression string `yaml:"compression" json:"compression"`
	}

	PrometheusConf struct {
		Enabled  bool   `yaml:"enabled"`
		GateAddr string `yaml:"gate_addr"`
		Interval int64  `yaml:"interval"`
	}
)

type (
	RestConf struct {
		Name     string `yaml:"name"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Verbose  bool   `yaml:"verbose"`
		MaxConns int    `yaml:"max_conns"`
		MaxBytes int64  `yaml:"max_bytes"`
		// milliseconds
		Timeout      int64 `yaml:"timeout"`
		CpuThreshold int64 `yaml:"cpu_threshold"`
	}

	RpcServerConf struct {
		Name          string            `yaml:"name"`
		Log           ConfigLog         `yaml:"log"`
		Mode          string            `yaml:"mode"`
		MetricsUrl    string            `yaml:"metrics_url"`
		Prometheus    PrometheusConf    `yaml:"prometheus"`
		ListenOn      string            `yaml:"listenon"`
		Auth          bool              `yaml:"auth"`
		Redis         RedisConf         `yaml:"redis"`
		Mysql         MysqlConf         `yaml:"mysql"`
		Es            EsConf            `yaml:"es"`
		NsqConsumer   NsqConsumerConf   `yaml:"nsq_consumer"`
		NsqProducer   NsqProducerConf   `yaml:"nsq_producer"`
		StrictControl bool              `yaml:"strict_control"`
		Timeout       int64             `yaml:"timeout"` // never set it to 0, if zero, the underlying will set to 2s automatically
		RpcCacheRedis RpcCacheRedisConf `yaml:"rpc_cache_redis"`
		Cos           COSConf           `yaml:"cos"`
	}

	RpcClientConf struct {
		Endpoints []string `yaml:"endpoints"`
		App       string   `yaml:"app"`
		Token     string   `yaml:"token"`
		Timeout   int64    `yaml:"timeout"`
		Name      string   `yaml:"name"`
		Env       string   `yaml:"env"` // prod or sit
		PollSize  int64    `yaml:"poll_size"`
	}
)

func (cc RpcClientConf) HasCredential() bool {
	return len(cc.App) > 0 && len(cc.Token) > 0
}

func ConfEnv() string {
	if env := os.Getenv("CONFIGOR_ENV"); env != "" {
		return env
	} else {
		return "local"
	}
}

func ConfSvcName() string {
	return os.Getenv("SVC_NAME")
}

func ConfSvcVersion() string {
	return os.Getenv("SVC_VERSION")
}

func ConfPushEnv() string {
	return os.Getenv("PUSH_ENV")
}

func GenConfigurationFile(file string) string {
	var (
		envFile string
		extname = path.Ext(file)
		env     = ConfEnv()
	)
	if env == "" || env == "local" {
		return file
	}

	if extname == "" {
		envFile = fmt.Sprintf("%v.%v", file, env)
	} else {
		envFile = fmt.Sprintf("%v.%v%v", strings.TrimSuffix(file, extname), env, extname)
	}

	return envFile
}
