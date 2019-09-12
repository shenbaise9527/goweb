/**
 * Copyright(c) 2019 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : user.go
 * @Author      : shenbaise9527
 * @Create      : 2019-09-07 18:36:21
 * @Modified    : 2019-09-12 11:10:25
 * @version     : 1.0
 * @Description :
 */
package main

import (
	"crypto/sha1"
	"fmt"
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
	result, err := stmt.Exec(u.UUID, u.Name, u.Email, encrypt(u.Password), time.Now())
	if err != nil {
		return
	}

	id, _ := result.LastInsertId()
	rows, _ := result.RowsAffected()
	logger.Debugf("insert successfully, id: %d, rows: %d", id, rows)

	return
}

//newSession 创建用户的会话.
func (u *User) newSession() (sess Session, err error) {
	sess = Session{
		UUID:      CreateUUID(),
		Email:     u.Email,
		UserID:    u.ID,
		CreatedAt: time.Now(),
	}

	statement := "insert into sessions (uuid, email, user_id, created_at) values (?, ?, ?, ?)"
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

//Login 登录,并返回session
func (u *User) Login() (sess Session, err error) {
	userFromDB := User{}
	userFromDB, err = u.userByEmail()
	if err != nil {
		return
	}

	if userFromDB.Password == encrypt(u.Password) {
		sess = Session{}
		sess, err = userFromDB.newSession()
		if err != nil {
			logger.Errorf("Failed to create session: %v", err)

			return
		}
	}

	return
}

//userByEmail 通过邮箱获取用户信息.
func (u *User) userByEmail() (user User, err error) {
	user = User{}
	rows, err := db.Query("select id, uuid, name, email, password, created_at from user where email = ?", u.Email)
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

func (s *Session) Check() bool {
	rows, err := db.Query("select id, email, user_id, created_at from sessions where uuid = ?", s.UUID)
	if err != nil {
		logger.Errorf("Failed to query session: %s", err)

		return false
	}

	flag := false
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&s.ID, &s.Email, &s.UserID, &s.CreatedAt)
		if err != nil {
			logger.Errorf("Failed to scan session: %s", err)

			break
		}

		flag = true
		break
	}

	return flag
}

func (s *Session) GetUser() (u User, err error) {
	u = User{}
	rows, err := db.Query("select id, uuid, name, email, password, created_at from user where id = ?", s.UserID)
	if err != nil {
		logger.Errorf("Failed to query user by session: %s", err)

		return
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&u.ID, &u.UUID, &u.Name, &u.Email, &u.Password, &u.CreatedAt)
		if err != nil {
			logger.Errorf("Failed to scan user by session: %s", err)

			return
		}

		break
	}

	return
}

//encrypt 加密.
func encrypt(plaintext string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(plaintext)))
}
