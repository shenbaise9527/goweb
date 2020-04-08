/**
 * Copyright(c) 2019 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : goweb.go
 * @Author      : shenbaise9527
 * @Create      : 2019-08-14 22:00:51
 * @Modified    : 2020-03-14 23:09:04
 * @version     : 1.0
 * @Description :
 */
package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shenbaise9527/goweb/multinamedtemplate"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	uuid "github.com/satori/go.uuid"
)

func createMyRender() render.HTMLRender {
	r := multinamedtemplate.NewRender()
	r.AddFromFilesByNamed("public_index", "layout", "templates/layout.html", "templates/public.navbar.html", "templates/index.html")
	r.AddFromFilesByNamed("private_index", "layout", "templates/layout.html", "templates/private.navbar.html", "templates/index.html")
	r.AddFromFilesByNamed("login", "layout", "templates/layout.html", "templates/public.navbar.html", "templates/login.html")
	r.AddFromFilesByNamed("public_error", "layout", "templates/layout.html", "templates/public.navbar.html", "templates/error.html")
	r.AddFromFilesByNamed("private_error", "layout", "templates/layout.html", "templates/private.navbar.html", "templates/error.html")
	r.AddFromFilesByNamed("signup", "layout", "templates/layout.html", "templates/public.navbar.html", "templates/signup.html")
	r.AddFromFilesByNamed("newthread", "layout", "templates/layout.html", "templates/private.navbar.html", "templates/new.thread.html")
	r.AddFromFilesByNamed("public_thread", "layout", "templates/layout.html", "templates/public.navbar.html", "templates/public.thread.html")
	r.AddFromFilesByNamed("private_thread", "layout", "templates/layout.html", "templates/private.navbar.html", "templates/private.thread.html")

	return r
}

// GinRecoveryMiddleware recovery.
func GinRecoveryMiddleware() gin.HandlerFunc {
	return gin.RecoveryWithWriter(logWriter)
}

// GinAuthMiddleware check session.
func GinAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess, err := sessionByContext(c)
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()

			return
		}

		c.Set("userid", sess.UserID)
		c.Next()
	}
}

func main() {
	err := NewLogger("web.log")
	if err != nil {
		fmt.Println(err.Error())

		return
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(GinLoggerMiddleware())
	r.Use(GinRecoveryMiddleware())
	r.HTMLRender = createMyRender()

	// file.
	r.Static("/static", "./public")

	r.GET("/", index)
	r.GET("/err", errmsg)
	r.GET("/login", login)
	r.GET("/logout", logout)
	r.GET("/signup", signup)
	r.POST("/signup_account", signupAccount)
	r.POST("/authenticate", authenticate)
	r.GET("/thread/read", readThread)

	subRouter := r.Group("/thread")
	subRouter.Use(GinAuthMiddleware())
	{
		subRouter.GET("/new", newThread)
		subRouter.POST("/create", createThread)
		subRouter.POST("/post", postThread)
	}

	// 设置开启gorm日志.
	db.LogMode(true)

	// 设置gorm的日志插件.
	db.SetLogger(&GormLogger{})

	// 删除所有session.
	db.Delete(&Session{})

	// 关闭连接.
	defer db.Close()

	// 启动gin服务.
	err = r.Run(":8000")
	if err != nil {
		logger.Errorf("err: %s", err)
	}
}

//CreateUUID 创建UUID.
func CreateUUID() string {
	u4 := uuid.NewV4()
	return u4.String()
}

func errmsg(c *gin.Context) {
	_, err := sessionByContext(c)
	msg := c.Query("msg")
	if err != nil {
		c.HTML(http.StatusOK, "public_error", msg)
	} else {
		c.HTML(http.StatusOK, "private_error", msg)
	}
}

func index(c *gin.Context) {
	threads, err := Threads()
	if err != nil {
		logger.Errorf("Failed to load threads: %s.", err)
		jumptoerror(c, fmt.Sprintf("Failed to load threads: %s.", err))
		return
	}

	_, err = sessionByContext(c)
	if err != nil {
		c.HTML(http.StatusOK, "public_index", threads)
	} else {
		c.HTML(http.StatusOK, "private_index", threads)
	}
}

func login(c *gin.Context) {
	c.HTML(http.StatusOK, "login", nil)
}

func logout(c *gin.Context) {
	s, err := sessionByContext(c)
	if err != http.ErrNoCookie {
		logger.Warnf("Failed to get cookie: %s", err)
		_ = s.DelByUUID()
	}

	c.Redirect(http.StatusFound, "/")
}

func signup(c *gin.Context) {
	c.HTML(http.StatusOK, "signup", nil)
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
	c.HTML(http.StatusOK, "newthread", nil)
}

func createThread(c *gin.Context) {
	topic, flag := c.GetPostForm("topic")
	if !flag {
		jumptoerror(c, "Failed to get topic")

		return
	}

	thr := Thread{
		UUID:      CreateUUID(),
		Topic:     topic,
		UserID:    c.MustGet("userid").(int),
		CreatedAt: time.Now(),
	}

	err := thr.NewThread()
	if err != nil {
		jumptoerror(c, fmt.Sprintf("Failed to create thread: %s", err))

		return
	}

	c.Redirect(http.StatusFound, "/")
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
		if err != nil {
			c.HTML(http.StatusOK, "public_thread", &thr)
		} else {
			c.HTML(http.StatusOK, "private_thread", &thr)
		}
	}
}

func postThread(c *gin.Context) {
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

	err := thr.GetThreadByUUID()
	if err != nil {
		jumptoerror(c, fmt.Sprintf("Failed to read thread: %s", err))

		return
	}

	pst := Post{
		UUID:      CreateUUID(),
		UserID:    c.MustGet("userid").(int),
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
