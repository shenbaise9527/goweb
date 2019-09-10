/**
 * Copyright(c) 2019 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : data.go
 * @Author      : shenbaise9527
 * @Create      : 2019-09-03 22:44:45
 * @Modified    : 2019-09-10 09:25:10
 * @version     : 1.0
 * @Description :
 */
package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	var err error
	// 设置字符集,datetime转化为time.Time类型,采用本地时区.
	db, err = sql.Open("mysql", "mtp2_test:muchinfo@tcp(127.0.0.1:3406)/web?charset=utf8&parseTime=true&loc=Local")
	if err != nil {
		log.Fatal(err)
	}
}
