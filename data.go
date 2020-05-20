/**
 * Copyright(c) 2019 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : data.go
 * @Author      : shenbaise9527
 * @Create      : 2019-09-03 22:44:45
 * @Modified    : 2020-05-20 21:57:14
 * @version     : 1.0
 * @Description :
 */
package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func NewDB() error {
	var err error
	// 设置字符集,datetime转化为time.Time类型,采用本地时区.
	//db, err = gorm.Open("mysql", "mtp2_test:muchinfo@tcp(127.0.0.1:3406)/web?charset=utf8&parseTime=true&loc=Local")
	db, err = gorm.Open("mysql", "web:muchinfo@tcp(127.0.0.1:3406)/web?charset=utf8&parseTime=true&loc=Local")
	if err != nil {
		return err
	}

	// 设置连接池信息.
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	return nil
}
