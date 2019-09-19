/**
 * Copyright(c) 2019 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : goweb.go
 * @Author      : shenbaise9527
 * @Create      : 2019-08-14 22:00:51
 * @Modified    : 2019-09-19 18:45:11
 * @version     : 1.0
 * @Description :
 */
package main

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

func createMyRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("public_index", "templates/layout.html", "templates/public.navbar.html", "templates/index.html")
	r.AddFromFiles("err", "templates/layout.html", "templates/public.navbar.html", "templates/error.html")

	return r
}

var (
	sqlRegexp                = regexp.MustCompile(`\?`)
	numericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
)

var logger *logrus.Logger

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

//gormLogger gorm日志插件.
type gormLogger struct {
}

func (dblooger *gormLogger) Print(values ...interface{}) {
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

func createLogger(logName string) (loggerClient *logrus.Logger) {
	loggerClient = logrus.New()

	// 显示行号等信息.
	loggerClient.SetReportCaller(true)

	// 禁止logrus的输出.
	src, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("err", err)
	}

	loggerClient.Out = src

	// 设置日志级别.
	loggerClient.SetLevel(logrus.DebugLevel)

	// 设置分割规则.
	logWriter, err := rotatelogs.New(
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

	loggerClient.AddHook(lfHook)

	return
}

func NewLogger() gin.HandlerFunc {
	logger = createLogger("web.log")

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

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(NewLogger())
	r.Use(gin.Recovery())
	//r.HTMLRender = createMyRender()

	// file.
	r.Static("/static", "./public")

	r.GET("/", index)
	r.GET("/err", errmsg)
	r.GET("/login", login)
	r.GET("/logout", logout)
	r.GET("/signup", signup)
	r.POST("/signup_account", signupAccount)
	r.POST("/authenticate", authenticate)

	r.GET("/thread/new", newThread)
	r.POST("/thread/create", createThread)
	r.GET("/thread/read", readThread)
	r.POST("/thread/post", postThread)

	// 设置开启gorm日志.
	db.LogMode(true)

	// 设置gorm的日志插件.
	db.SetLogger(&gormLogger{})

	// 删除所有session.
	db.Delete(&Session{})

	// 关闭连接.
	defer db.Close()

	// 启动gin服务.
	err := r.Run(":8000")
	if err != nil {
		logger.Errorf("err: %s", err)
	}
}

//CreateUUID 创建UUID.
func CreateUUID() string {
	u4 := uuid.NewV4()
	return fmt.Sprintf("%s", u4)
}

func errmsg(c *gin.Context) {
	_, err := sessionByContext(c)
	msg := c.Query("msg")
	var files []string
	if err != nil {
		files = []string{"templates/layout.html", "templates/public.navbar.html", "templates/error.html"}
	} else {
		files = []string{"templates/layout.html", "templates/private.navbar.html", "templates/error.html"}
	}

	execTemplate(c, files, msg)
}

func index(c *gin.Context) {
	threads, err := Threads()
	if err != nil {
		logger.Errorf("Failed to load threads: %s.", err)
		jumptoerror(c, fmt.Sprintf("Failed to load threads: %s.", err))
		return
	}

	//c.HTML(http.StatusOK, "public_index", threads)
	_, err = sessionByContext(c)
	var files []string
	if err != nil {
		files = []string{"templates/layout.html", "templates/public.navbar.html", "templates/index.html"}
	} else {
		files = []string{"templates/layout.html", "templates/private.navbar.html", "templates/index.html"}
	}

	execTemplate(c, files, threads)
}

func login(c *gin.Context) {
	files := []string{"templates/layout.html", "templates/public.navbar.html", "templates/login.html"}
	execTemplate(c, files, nil)
}

func logout(c *gin.Context) {
	s, err := sessionByContext(c)
	if err != http.ErrNoCookie {
		logger.Warnf("Failed to get cookie: %s", err)
		s.DelByUUID()
	}

	c.Redirect(http.StatusFound, "/")
}

func signup(c *gin.Context) {
	files := []string{"templates/layout.html", "templates/public.navbar.html", "templates/signup.html"}
	execTemplate(c, files, nil)
}

func signupAccount(c *gin.Context) {
	u := User{
		UUID:     CreateUUID(),
		Name:     c.PostForm("name"),
		Email:    c.PostForm("email"),
		Password: c.PostForm("password"),
	}

	if err := u.Create(); err != nil {
		logger.Errorf("Failed to create user: %s.", err)
		jumptoerror(c, fmt.Sprintf("Failed to create user: %s.", err))

		return
	}

	c.Redirect(http.StatusFound, "/login")
}

func authenticate(c *gin.Context) {
	u := User{
		Email:    c.PostForm("email"),
		Password: c.PostForm("password"),
	}

	sess, err := u.Login()
	if err != nil {
		logger.Errorf("Failed to login: %s.", err)
		jumptoerror(c, fmt.Sprintf("Failed to login: %s.", err))
	} else {
		c.SetCookie("goweb", sess.UUID, 300, "", "", false, true)
		c.Redirect(http.StatusFound, "/")
	}
}

func newThread(c *gin.Context) {
	_, err := sessionByContext(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
	} else {
		files := []string{"templates/layout.html", "templates/private.navbar.html", "templates/new.thread.html"}
		execTemplate(c, files, nil)
	}
}

func createThread(c *gin.Context) {
	sess, err := sessionByContext(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
	} else {
		topic, flag := c.GetPostForm("topic")
		if !flag {
			jumptoerror(c, fmt.Sprintf("Failed to get topic: %s", err))

			return
		}

		thr := Thread{
			UUID:      CreateUUID(),
			Topic:     topic,
			UserID:    sess.UserID,
			CreatedAt: time.Now(),
		}

		err = thr.NewThread()
		if err != nil {
			jumptoerror(c, fmt.Sprintf("Failed to create thread: %s", err))

			return
		}

		c.Redirect(http.StatusFound, "/")
	}
}

func readThread(c *gin.Context) {
	uuid := c.Query("id")
	thr := Thread{
		UUID: uuid,
	}

	err := thr.GetThreadByUUID()
	if err != nil {
		jumptoerror(c, fmt.Sprintf("Failed to read thread: %s", err))
	} else {
		_, err = sessionByContext(c)
		var files []string
		if err != nil {
			files = []string{"templates/layout.html", "templates/public.navbar.html", "templates/public.thread.html"}
		} else {
			files = []string{"templates/layout.html", "templates/private.navbar.html", "templates/private.thread.html"}
		}

		execTemplate(c, files, &thr)
	}
}

func postThread(c *gin.Context) {
	s, err := sessionByContext(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")

		return
	}

	body, flag := c.GetPostForm("body")
	if !flag {
		jumptoerror(c, "data error")

		return
	}

	uuid, flag := c.GetPostForm("uuid")
	if !flag {
		jumptoerror(c, "data error")

		return
	}

	thr := Thread{
		UUID: uuid,
	}

	err = thr.GetThreadByUUID()
	if err != nil {
		jumptoerror(c, fmt.Sprintf("Failed to read thread: %s", err))

		return
	}

	pst := Post{
		UUID:      CreateUUID(),
		UserID:    s.UserID,
		ThreadID:  thr.ID,
		Body:      body,
		CreatedAt: time.Now(),
	}

	err = pst.NewPost()
	if err != nil {
		jumptoerror(c, fmt.Sprintf("Failed to read thread: %s", err))

		return
	}

	url := fmt.Sprintf("/thread/read?id=%s", uuid)
	c.Redirect(http.StatusFound, url)
}

func jumptoerror(c *gin.Context, msg string) {
	c.Redirect(http.StatusFound, fmt.Sprintf("/err?msg=%s", msg))
}

func execTemplate(c *gin.Context, files []string, data interface{}) {
	t := template.Must(template.ParseFiles(files...))
	t.ExecuteTemplate(c.Writer, "layout", data)
}

func sessionByContext(c *gin.Context) (sess Session, err error) {
	sess = Session{}
	value, err := c.Cookie("goweb")
	if err != nil {
		logger.Errorf("Failed to get session: %s", err)

		return
	}

	sess.UUID = value
	if ok := sess.Check(); !ok {
		err = errors.New("invalid session")
	}

	return
}
