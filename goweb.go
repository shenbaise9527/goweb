/**
 * Copyright(c) 2019 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : goweb.go
 * @Author      : shenbaise9527
 * @Create      : 2019-08-14 22:00:51
 * @Modified    : 2019-09-11 22:58:52
 * @version     : 1.0
 * @Description :
 */
package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

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

func createLogger(logName string) (loggerClient *logrus.Logger) {
	loggerClient = logrus.New()
	src, err := os.OpenFile(logName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("err", err)
	}

	loggerClient.Out = src
	loggerClient.SetLevel(logrus.DebugLevel)
	logWriter, err := rotatelogs.New(
		logName+".%Y-%m-%d.log",
		rotatelogs.WithLinkName(logName),
		rotatelogs.WithMaxAge(7*24*time.Hour),
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
		TimestampFormat: "2006-01-02 15:04:05",
	})

	loggerClient.AddHook(lfHook)

	return
}

var logger *logrus.Logger

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
	//gin.DisableConsoleColor()
	r := gin.New()
	r.Use(NewLogger())
	r.Use(gin.Recovery())
	//r.HTMLRender = createMyRender()

	// file.
	r.Static("/static", "./public")

	r.GET("/", index)
	r.GET("/err", errmsg)
	r.GET("/login", login)
	r.GET("/signup", signup)
	r.POST("/signup_account", signupAccount)
	r.POST("/authenticate", authenticate)

	r.GET("/thread/new", newThread)
	r.POST("/thread/create", createThread)
	r.GET("/thread/read", readThread)
	r.POST("/thread/post", postThread)
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
	files := []string{"templates/layout.html", "templates/public.navbar.html", "templates/index.html"}
	execTemplate(c, files, threads)
}

func login(c *gin.Context) {
	files := []string{"templates/layout.html", "templates/public.navbar.html", "templates/login.html"}
	execTemplate(c, files, nil)
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
		jumptoerror(c, fmt.Sprintf("invalid session: ", err))
	} else {
		files := []string{"templates/layout.html", "templates/private.navbar.html", "templates/new.thread.html"}
		execTemplate(c, files, nil)
	}
}

func createThread(c *gin.Context) {
	sess, err := sessionByContext(c)
	if err != nil {
		jumptoerror(c, fmt.Sprintf("invalid session: ", err))
	} else {
		topic, flag := c.GetPostForm("topic")
		if !flag {
			jumptoerror(c, fmt.Sprintf("Failed to get topic: ", err))

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
			jumptoerror(c, fmt.Sprintf("Failed to create thread: ", err))

			return
		}

		c.Redirect(http.StatusFound, "/")
	}
}

func readThread(c *gin.Context) {
}

func postThread(c *gin.Context) {
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
