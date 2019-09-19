/**
 * Copyright(c) 2019 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : user.go
 * @Author      : shenbaise9527
 * @Create      : 2019-09-07 18:36:21
 * @Modified    : 2019-09-18 15:52:00
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
	UUID      string    `gorm: "column:uuid;type:varchar(64)"`
	Name      string    `gorm: "column:name;type:varchar(256)"`
	Email     string    `gorm: "column:email;type:varchar(256)"`
	Password  string    `gorm: "column:password;type:varchar(256)"`
	CreatedAt time.Time `gorm: "column:created_at;type:datetime"`
}

//Session 会话信息.
type Session struct {
	ID        int
	UUID      string    `gorm: "column:uuid;type:varchar(64)"`
	Email     string    `gorm: "column:email;type:varchar(256)"`
	UserID    int       `gorm: "column:user_id;type:int(11)"`
	CreatedAt time.Time `gorm: "column:created_at;type:datetime"`
}

//TableName 对应的表名.
func (u *User) TableName() string {
	return "user"
}

//Create 创建一个新用户.
func (u *User) Create() (err error) {
	idb := db.Create(u)
	err = idb.Error
	if err != nil {
		return
	}

	rows := idb.RowsAffected
	id := u.ID
	logger.Debugf("insert user successfully, id: %d, name: %s, rows: %d", id, u.Name, rows)

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

	idb := db.Create(&sess)
	err = idb.Error
	if err != nil {
		return
	}

	rows := idb.RowsAffected
	id := sess.ID
	logger.Debugf("insert session successfully, id: %d, name: %s, rows: %d", id, u.Name, rows)

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
	err = db.Where("email = ?", u.Email).First(&user).Error

	return
}

func (s *Session) Check() bool {
	idb := db.Where("uuid = ?", s.UUID).First(s)
	err := idb.Error
	if err != nil {
		logger.Errorf("Failed to query session,uuid: %s, err: %s", s.UUID, err)

		return false
	}

	return 1 == idb.RowsAffected
}

func (s *Session) GetUser() (u User, err error) {
	u = User{}
	err = db.Where("id = ?", s.UserID).First(&u).Error
	if err != nil {
		logger.Errorf("Failed to query user by session,userid: %d, err: %s", s.UserID, err)

		return
	}

	return
}

func (s *Session) DelByUUID() (err error) {
	err = db.Where("uuid = ?", s.UUID).Delete(Session{}).Error
	if err != nil {
		logger.Errorf("Failed to del session,uuid:%s, err: %s", s.UUID, err)
	}

	return
}

//encrypt 加密.
func encrypt(plaintext string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(plaintext)))
}
