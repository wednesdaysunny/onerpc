package toolkit

import (
	"database/sql/driver"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"time"
	"unicode"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	oconf "github.com/wednesdaysunny/onerpc/eco/inter/conf"
)

// CreateDB 初始化MYSQL实例
func CreateDB(config oconf.MysqlConf) *gorm.DB {
	var (
		username = config.Username
		password = config.Password
		host     = config.Host
		port     = config.Port
		dbName   = config.DBName
		maxIdle  = config.MaxIdle
		maxOpen  = config.MaxConn
	)
	//connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
	//	username,
	//	password,
	//	host,
	//	port,
	//	dbName,
	//)
	if config.Charset == "" {
		config.Charset = "utf8"
	}
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True",
		username,
		password,
		host,
		port,
		dbName,
		config.Charset,
	)

	fmt.Println("Try to connect to MYSQL host: ", host, ", port: ", port)
	db, err := gorm.Open("mysql", connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to connect MYSQL %s:%d/%s: %s", host, port, dbName, err.Error()))
	}
	fmt.Println("Connected to MYSQL: ", host, ", port: ", port)

	if config.ShowLog {
		db.LogMode(true)
	}

	if config.LogType == "logrus" {
		db.SetLogger(Logger{})
	}

	/*
	 * maxIdle，最大空闲数，数据库连接的最大空闲时间。超过空闲时间，数据库连接将被标记为不可用，然后被释放。设为0表示无限制。
	 * maxOpen，用于设置Database最大可以打开的连接数
	 */
	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(maxOpen)
	db.AutoMigrate()

	return db
}

func CreateDBForUTF8MB4(config oconf.MysqlConf) *gorm.DB {
	var (
		username = config.Username
		password = config.Password
		host     = config.Host
		port     = config.Port
		dbName   = config.DBName
		maxIdle  = config.MaxIdle
		maxOpen  = config.MaxConn
	)
	//connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
	//	username,
	//	password,
	//	host,
	//	port,
	//	dbName,
	//)
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=%s",
		username,
		password,
		host,
		port,
		dbName,
		url.QueryEscape("Asia/Shanghai"),
	)

	fmt.Println("Try to connect to MYSQL host: ", host, ", port: ", port)
	db, err := gorm.Open("mysql", connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to connect MYSQL %s:%d/%s: %s", host, port, dbName, err.Error()))
	}
	fmt.Println("Connected to MYSQL: ", host, ", port: ", port)

	if !config.ShowLog {
		db.LogMode(true)
	}
	if config.LogType == "logrus" {
		db.SetLogger(Logger{})
	}

	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(maxOpen)
	db.AutoMigrate()

	return db
}

type DbSession struct {
	session *gorm.DB
	closed  bool
}

func BeginDbSession(db *gorm.DB) *DbSession {
	return &DbSession{db.Begin(), false}
}

func (s *DbSession) Api() *gorm.DB {
	if !s.closed {
		return s.session
	}
	return nil
}

func (s *DbSession) Commit() *gorm.DB {
	if !s.closed {
		s.closed = true
		return s.session.Commit()
	}
	return nil
}

func (s *DbSession) Rollback() *gorm.DB {
	if !s.closed {
		s.closed = true
		return s.session.Rollback()
	}
	return nil
}

// 为什么要加 defer sessin.End() 呢，不怕你现在写错，怕以后维护的人写错
// 为什么要给 goroutine 加 defer recover 呢？ 同理，这不是php或者python，挂了就真挂了
func (s *DbSession) End() *gorm.DB {
	if !s.closed {
		s.closed = true
		fmt.Println("DbSession closed by End call.")
		return s.session.Rollback()
	}
	return nil
}

var (
	sqlRegexp = regexp.MustCompile(`(\$\d+)|\?`)
)

const (
	CreatedAtColumn = "created_at"
	UpdatedAtColumn = "updated_at"
)

// TimeMixin mixin
type TimeMixin struct {
	CreatedAt time.Time `gorm:"column:created_at;type:TIMESTAMP(6);default:CURRENT_TIMESTAMP(6);index"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:TIMESTAMP(6);default:CURRENT_TIMESTAMP(6);index"`
}

// Logger customizes the gorm logger.
type Logger struct {
}

// Print uses logrus to log SQL.
func (Logger Logger) Print(values ...interface{}) {
	if len(values) > 1 {
		level := values[0]
		currentTime := "[" + gorm.NowFunc().Format("2006-01-02 15:04:05") + "]"
		source := fmt.Sprintf("(%v)", values[1])
		messages := []interface{}{source, currentTime}

		if level == "sql" {
			// duration
			messages = append(messages, fmt.Sprintf(" [%.2fms] ", float64(values[2].(time.Duration).Nanoseconds()/1e4)/100.0))
			// sql
			var sql string
			var formattedValues []string

			for _, value := range values[4].([]interface{}) {
				indirectValue := reflect.Indirect(reflect.ValueOf(value))
				if indirectValue.IsValid() {
					value = indirectValue.Interface()
					if t, ok := value.(time.Time); ok {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format(time.RFC3339)))
					} else if b, ok := value.([]byte); ok {
						if str := string(b); isPrintable(str) {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
						} else {
							formattedValues = append(formattedValues, "'<binary>'")
						}
					} else if r, ok := value.(driver.Valuer); ok {
						if value, err := r.Value(); err == nil && value != nil {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
						} else {
							formattedValues = append(formattedValues, "NULL")
						}
					} else {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
					}
				} else {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
				}
			}

			var formattedValuesLength = len(formattedValues)
			for index, value := range sqlRegexp.Split(values[3].(string), -1) {
				sql += value
				if index < formattedValuesLength {
					sql += formattedValues[index]
				}
			}

			messages = append(messages, sql)
		} else {
			messages = append(messages, values[2:]...)
		}
		fmt.Println(messages)
	}
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}
