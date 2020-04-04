/**
 * Copyright(c) 2020 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : logger.go
 * @Author      : shenbaise9527
 * @Create      : 2020-03-14 22:59:28
 * @Modified    : 2020-03-14 23:05:03
 * @version     : 1.0
 * @Description :
 */
package main

import (
	"database/sql/driver"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

// logger 日志对象.
var logger *logrus.Logger
var logWriter io.Writer

var (
	sqlRegexp                = regexp.MustCompile(`\?`)
	numericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
)

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

// GormLogger gorm日志插件.
type GormLogger struct {
}

// Print gorm的打印接口.
func (*GormLogger) Print(values ...interface{}) {
	if len(values) > 1 {
		var (
			sql             string
			formattedValues []string
			level           = values[0]
			source          = fmt.Sprintf("(%v)", values[1])
		)

		messages := []interface{}{source}

		if level == "sql" {
			// duration
			messages = append(messages, fmt.Sprintf("[%.2fms]", float64(values[2].(time.Duration).Nanoseconds()/1e4)/100.0))
			// sql

			for _, value := range values[4].([]interface{}) {
				indirectValue := reflect.Indirect(reflect.ValueOf(value))
				if indirectValue.IsValid() {
					value = indirectValue.Interface()
					if t, ok := value.(time.Time); ok {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
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
						switch value.(type) {
						case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
							formattedValues = append(formattedValues, fmt.Sprintf("%v", value))
						default:
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
						}
					}
				} else {
					formattedValues = append(formattedValues, "NULL")
				}
			}

			// differentiate between $n placeholders or else treat like ?
			if numericPlaceHolderRegexp.MatchString(values[3].(string)) {
				sql = values[3].(string)
				for index, value := range formattedValues {
					placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
					sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
				}
			} else {
				formattedValuesLength := len(formattedValues)
				for index, value := range sqlRegexp.Split(values[3].(string), -1) {
					sql += value
					if index < formattedValuesLength {
						sql += formattedValues[index]
					}
				}
			}

			messages = append(messages, sql)
			messages = append(messages, fmt.Sprintf("[%v]", strconv.FormatInt(values[5].(int64), 10)+" rows affected or returned "))
		} else {
			messages = append(messages, values[2:]...)
		}

		logger.Debug(messages)
	}

	return
}

// NewLogger 创建日志对象.
func NewLogger(logName string) error {
	logger = logrus.New()

	// 显示行号等信息.
	logger.SetReportCaller(true)

	// 禁止logrus的输出.
	src, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}

	logger.Out = src

	// 设置日志级别.
	logger.SetLevel(logrus.DebugLevel)

	// 设置分割规则.
	logWriter, err = rotatelogs.New(
		// 分割后的文件名.
		logName+".%Y-%m-%d.log",

		// 设置文件软连接,方便找到当前日志文件.
		rotatelogs.WithLinkName(logName),

		// 设置文件清理前的最长保存时间,参数=-1表示不清除.
		rotatelogs.WithMaxAge(7*24*time.Hour),

		// 设置文件清理前最多保存的个数,不能与WithMaxAge同时使用.
		//rotatelogs.WithRotationCount(10),

		// 设置日志分割时间,这里设置24小时分割一次.
		rotatelogs.WithRotationTime(24*time.Hour),
	)

	writerMap := lfshook.WriterMap{
		logrus.InfoLevel:  logWriter,
		logrus.FatalLevel: logWriter,
		logrus.DebugLevel: logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}

	lfHook := lfshook.NewHook(writerMap, &logrus.TextFormatter{
		// 格式化输出时间.
		TimestampFormat: "2006-01-02 15:04:05",
	})

	logger.AddHook(lfHook)

	return nil
}

//GinLoggerMiddleware 生成gin的日志插件.
func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		end := time.Now()
		latency := end.Sub(start)
		path := c.Request.URL.RequestURI()
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		logger.Infof("|%3d|%13v|%15s|%s %s|", statusCode, latency, clientIP, method, path)
	}
}
