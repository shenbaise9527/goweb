/**
 * Copyright(c) 2019 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : goweb.go
 * @Author      : shenbaise9527
 * @Create      : 2019-08-14 22:00:51
 * @Modified    : 2019-09-07 22:09:37
 * @version     : 1.0
 * @Description :
 */
package main

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

func createMyRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("public_index", "templates/layout.html", "templates/public.navbar.html", "templates/index.html")
	r.AddFromFiles("err", "templates/layout.html", "templates/public.navbar.html", "templates/error.html")

	return r
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	//gin.DisableConsoleColor()
	//f, err := os.Create("gin.log")
	f, err := os.OpenFile("gin.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	r := gin.Default()
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
	r.Run(":8000")
}

//CreateUUID 创建UUID.
func CreateUUID() string {
	u4 := uuid.NewV4()
	return fmt.Sprintf("%s", u4)
}

//Encrypt 加密.
func Encrypt(plaintext string) string {
	crypttext := fmt.Sprintf("%x", sha1.Sum([]byte(plaintext)))
	return crypttext
}

func errmsg(c *gin.Context) {
	msg := c.Query("msg")
	//c.HTML(http.StatusOK, "err", msg)
	files := []string{"templates/layout.html", "templates/public.navbar.html", "templates/error.html"}
	t := template.Must(template.ParseFiles(files...))
	t.ExecuteTemplate(c.Writer, "layout", msg)
}

func index(c *gin.Context) {
	threads, err := Threads()
	if err != nil {
		log.Printf("Failed to load threads: %v", err)
		return
	}

	//c.HTML(http.StatusOK, "public_index", threads)
	files := []string{"templates/layout.html", "templates/public.navbar.html", "templates/index.html"}
	t := template.Must(template.ParseFiles(files...))
	t.ExecuteTemplate(c.Writer, "layout", threads)
}

func login(c *gin.Context) {
	files := []string{"templates/layout.html", "templates/public.navbar.html", "templates/login.html"}
	t := template.New("layout")
	t = template.Must(t.ParseFiles(files...))
	t.Execute(c.Writer, nil)
}

func signup(c *gin.Context) {
	files := []string{"templates/layout.html", "templates/public.navbar.html", "templates/signup.html"}
	t := template.Must(template.ParseFiles(files...))
	t.ExecuteTemplate(c.Writer, "layout", nil)
}

func signupAccount(c *gin.Context) {
	u := User{
		UUID:     CreateUUID(),
		Name:     c.PostForm("name"),
		Email:    c.PostForm("email"),
		Password: c.PostForm("password"),
	}

	if err := u.Create(); err != nil {
		log.Printf("Failed to create user: %v", err)
	}

	c.Redirect(http.StatusFound, "/login")
}

func authenticate(c *gin.Context) {
	email := c.PostForm("email")
	user, err := UserByEmail(email)
	if err != nil {
		log.Printf("Failed to query user: %v", err)
		c.Redirect(http.StatusFound, "/login")

		return
	}

	if user.Password == Encrypt(c.PostForm("password")) {
		log.Printf("login successfully")
		c.Redirect(http.StatusFound, "/")
	} else {
		log.Printf("Faile to login: Password error")
		c.Redirect(http.StatusFound, "/login")
	}
}
