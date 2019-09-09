/**
 * Copyright(c) 2019 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : user.go
 * @Author      : shenbaise9527
 * @Create      : 2019-09-07 18:36:21
 * @Modified    : 2019-09-09 23:19:19
 * @version     : 1.0
 * @Description :
 */
package main

import (
	"errors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
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

//NewSession 创建用户的会话.
func (u *User) NewSession() (sess Session, err error) {
	sess = Session{
		UUID:      CreateUUID(),
		Email:     u.Email,
		UserID:    u.ID,
		CreatedAt: time.Now(),
	}

	statement := "insert into session (uuid, email, user_id, created_at) values (?, ?, ?, ?, ?)"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}

	result, err := stmt.Exec(&sess.UUID, &sess.Email, &sess.UserID, &sess.CreatedAt)
	if err != nil {
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		return
	}

	sess.ID = int(id)
	return
}

func (sess *Session) Check() bool {
	return true
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

func SessionByContext(c *gin.Context) (sess Session, err error) {
	sess = Session{}
	value, err := c.Cookie("goweb")
	if err != nil {
		return
	}

	sess.UUID = value
	if ok := sess.Check(); !ok {
		err = errors.New("invalid session")
	}

	return
}
