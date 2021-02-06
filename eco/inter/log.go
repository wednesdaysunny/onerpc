package onecommon

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	oconf "github.com/wednesdaysunny/onerpc/eco/inter/conf"

	"github.com/sirupsen/logrus"
	stdnet "github.com/wednesdaysunny/onerpc/eco/inter/toolkit/net"
)

// LogFields indicates the log's tags
type LogFields map[string]interface{}

var (
	logFile    *os.File
	loc        *time.Location
	BaseLogger *logrus.Logger
	fireLogger *logrus.Logger
)

const (
	// TagTopic flags the topic
	TagTopic = "topic"

	// TopicCodeTrace traces the running of code
	TopicCodeTrace = "code_trace"

	// TopicBugReport indicates the bug report topic
	TopicBugReport = "bug_report"

	// TopicCrash indicates the program's panics
	TopicCrash = "crash"

	// TopicUserActivity indicates the user activity like web access, user login/logout
	TopicUserActivity = "user_activity"

	// TagCategory tags the log category
	TagCategory = "category"

	// TagError tags the error category
	TagError = "error"

	// CategoryRPC indicates the rpc category
	CategoryRPC = "rpc"

	// CategoryRedis indicates the redis category
	CategoryRedis = "redis"

	// CategoryMySQL indicates the MySQL category
	CategoryMySQL = "mysql"

	// CategoryElasticsearch indicates the Elasticsearch category
	CategoryElasticsearch = "elasticsearch"
)

var (
	//用于搜集启动信息的专用DSN
	baseDSN            = "https://ba20ccce6b9b4ab3bf49e6847be149cc:e7e67cc3a0794b568be1fcbf095f810c@sentry.xuangubao.cn/66"
	fireTimeoutSeconds = time.Second * 3
	ProdEnv            = "prod"
	TestEnv            = "test"
)

func init() {

	//BaseLogger用于捕获服务启动层面的panic信息, 此时有可能业务Logger(即default Logger)还没有成功启动
	BaseLogger = logrus.New()
	BaseLogger.SetLevel(logrus.PanicLevel)

	//转用于触发报警的Logger, 并没有采用默认的, 因为业务开发人员很容易弱化日志级别概念, 导致过多fire发生
	fireLogger = logrus.New()
	fireLogger.SetLevel(logrus.ErrorLevel)
}

// 初始化log设置
func InitLog(conf oconf.ConfigLog) {
	if conf.Level == 0 {
		conf.Level = 4
	}
	logrus.SetLevel(logrus.Level(conf.Level))
	loc, _ = time.LoadLocation("Asia/Shanghai")

	// output_dest为file时输出至指定目录下的log文件
	// 其他统一设置为标准输出&标准错误输出
	logrus.SetFormatter(&logrus.JSONFormatter{})
	switch {
	case conf.OutputDest == "file":
		if conf.Path == "" {
			BaseLogger.Panicln("log file path is empty")
		}

		// redirect all stdout & stderr to file
		redirect(logName(conf.Path, "127.0.0.1", time.Now()))

		go func() {
			// check rotation every 1 minute
			ticker := time.NewTicker(60 * time.Second)
			for range ticker.C {
				logRotate()
			}
		}()
	default:
		logrus.Infoln("log output to standard output & err output")
	}
}

func logName(path string, ip string, time time.Time) string {
	year, month, day := time.Date()

	// replace all dots
	path = strings.Replace(path, ".", "_", -1)
	ip = strings.Replace(ip, ".", "_", -1)
	return fmt.Sprintf("%s.%04d%02d%02d.%s.log", path, year, int(month), day, ip)
}

func logRotate() {
	if logFile == nil {
		return
	}

	parts := strings.Split(logFile.Name(), ".")
	path := parts[0]
	date := parts[1]
	ip := parts[2]

	now := time.Now()
	if loc == nil {
		LogErrorc("timezone", nil, "fail to load Asia/Shanghai")
	} else {
		now = now.In(loc)
	}

	year, month, day := now.Date()
	if date >= fmt.Sprintf("%04d%02d%02d", year, int(month), day) {
		return
	}

	fmt.Println("redirect to: ", logName(path, ip, now))

	// redirect all stdout & stderr to file
	curFile := logFile
	defer curFile.Close()
	redirect(logName(path, ip, now))
}

func redirect(fullPath string) {
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666|os.ModeSticky)
	if err == nil {
		logFile = file
		logrus.SetOutput(file)
		syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd()))
		syscall.Dup2(int(file.Fd()), int(os.Stdout.Fd()))
	} else {
		panic("log file open error: " + err.Error())
	}
}

func getIP(CIDRs []string) string {
	var ip string
	var err error
	if len(CIDRs) > 0 {
		ip, err = stdnet.LocalIPAddrWithin(CIDRs)
		if err != nil {
			return ""
		}
	} else {
		ip, err = stdnet.LocalIPAddr()
		if err != nil {
			return ""
		}
	}
	return ip
}

// LogInfo records Info level information which helps trace the running of program and
// moreover the production infos
func LogInfo(fields LogFields, message string) {
	logrus.WithFields(logrus.Fields{
		TagTopic: TopicCodeTrace,
	}).WithFields(map[string]interface{}(fields)).Info(message)
}

// LogInfoc records the running infos
func LogInfoc(category string, message string) {
	logrus.WithFields(logrus.Fields{
		TagTopic:    TopicCodeTrace,
		TagCategory: category,
	}).Info(message)
}

// LogWarn records the warnings which are expected to be removed, but not influence the
// running of the program
func LogWarn(fields LogFields, message string) {
	logrus.WithFields(logrus.Fields{
		TagTopic: TopicBugReport,
	}).WithFields(map[string]interface{}(fields)).Warn(message)
}

// LogWarnc records the running warnings which are expected to be noticed
func LogWarnc(category string, err error, message string) {
	logrus.WithFields(logrus.Fields{
		TagTopic:    TopicBugReport,
		TagCategory: category,
		TagError:    err,
	}).Warn(message)
}

// LogError records the running errors which are expected to be solved soon
func LogError(fields LogFields, message string) {
	logrus.WithFields(logrus.Fields{
		TagTopic: TopicBugReport,
	}).WithFields(map[string]interface{}(fields)).Error(message)
}

// LogErrorc records the running errors which are expected to be solved soon
func LogErrorc(category string, err error, message string) {
	logrus.WithFields(logrus.Fields{
		TagTopic:    TopicBugReport,
		TagCategory: category,
		TagError:    err,
	}).Error(message)
}

// LogPanic records the running errors which are expected to be severe soon
func LogPanic(fields LogFields, message string) {
	logrus.WithFields(logrus.Fields{
		TagTopic: TopicBugReport,
	}).WithFields(map[string]interface{}(fields)).Panic(message)
}

// LogPanicc records the running errors which are expected to be severe soon
func LogPanicc(category string, err error, message string) {
	logrus.WithFields(logrus.Fields{
		TagTopic:    TopicBugReport,
		TagCategory: category,
		TagError:    err,
	}).Panic(message)
}

// LogInfoLn records Info level information which helps trace the running of program and
// moreover the production infos
func LogInfoLn(args ...interface{}) {
	logrus.Infoln(args...)
}

func LogImportantInfoLn(args ...interface{}) {
	newArgs := []interface{}{"ImportantLog"}
	newArgs = append(newArgs, args...)
	logrus.Infoln(newArgs...)
}

// LogWarnLn records the program warning
func LogWarnLn(args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		TagTopic: TopicBugReport,
	}).Warnln(args...)
}

// LogErrorLn records the program error, go to fix it!
func LogErrorLn(args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		TagTopic: TopicBugReport,
	}).Errorln(args...)
}

func LogImportantErrorLn(args ...interface{}) {
	newArgs := []interface{}{"ImportantLog"}
	newArgs = append(newArgs, args...)
	logrus.WithFields(logrus.Fields{
		TagTopic: TopicBugReport,
	}).Errorln(newArgs...)
}

// LogFatalLn records the program fatal error, developer should follow immediately
func LogFatalLn(args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		TagTopic: TopicBugReport,
	}).Fatalln(args...)
}

// LogPanicLn records the program fatal error, developer should fix otherwise the company dies
func LogPanicLn(args ...interface{}) {
	logrus.WithFields(logrus.Fields{
		TagTopic: TopicBugReport,
	}).Panicln(args...)
}

// LogDebugLn records debug information which helps trace the running of program
func LogDebugLn(args ...interface{}) {
	logrus.Debugln(args...)
}

// LogDebugc records the running infos
func LogDebugc(category string, message string) {
	logrus.WithFields(logrus.Fields{
		TagTopic:    TopicCodeTrace,
		TagCategory: category,
	}).Debug(message)
}

// LogUserActivity records user activity, like user access page, login/logout
func LogUserActivity(fields LogFields, message string) {
	logrus.WithFields(logrus.Fields{
		TagTopic: TopicUserActivity,
	}).WithFields(map[string]interface{}(fields)).Infoln(message)
}

// LogRecover records when program crashes
func LogRecover(e interface{}) {
	logrus.WithFields(logrus.Fields{
		"type":       "panicaccess",
		TagTopic:     TopicCrash,
		"error":      e,
		"stacktrace": string(debug.Stack()),
	}).Errorln("Recovered panic")
}

func LogErrorLnWithFire(args ...interface{}) {
	fireLogger.WithFields(logrus.Fields{
		TagTopic: TopicBugReport,
	}).Errorln(args...)
}

const (
	Unknown = "Unknown"
)
