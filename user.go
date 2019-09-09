/**
 * Copyright(c) 2019 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : user.go
 * @Author      : shenbaise9527
 * @Create      : 2019-09-07 18:36:21
 * @Modified    : 2019-09-08 12:11:55
 * @version     : 1.0
 * @Description :
 */
package main

import (
	"log"
	"time"
)

//User 用户信息.
type User struct {
	ID        int
	UUID      string
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
}

//Session 会话信息.
type Session struct {
	ID        int
	UUID      string
	Email     string
	UserID    int
	CreatedAt time.Time
}

//Create 创建一个新用户.
func (u *User) Create() (err error) {
	statement := "insert into user (uuid, name, email, password, created_at) values (?, ?, ?, ?, ?)"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}

	defer stmt.Close()
	result, err := stmt.Exec(u.UUID, u.Name, u.Email, Encrypt(u.Password), time.Now())
	if err != nil {
		return
	}

	id, _ := result.LastInsertId()
	rows, _ := result.RowsAffected()
	log.Printf("insert successfully, id: %d, rows: %d", id, rows)
	//err = stmt.QueryRow(u.UUID, u.Name, u.Email, Encrypt(u.Password), time.Now()).Scan(&u.ID, &u.UUID, &u.Name, &u.Email, &u.Password, &u.CreatedAt)

	return
}

//UserByEmail 通过邮箱获取用户信息.
func UserByEmail(email string) (user User, err error) {
	user = User{}
	rows, err := db.Query("select id, uuid, name, email, password, created_at from user where email = ?", email)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.ID, &user.UUID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
		if err != nil {
			break
		}

		break
	}

	return
}
