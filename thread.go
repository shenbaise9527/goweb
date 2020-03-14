/**
 * Copyright(c) 2019 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : thread.go
 * @Author      : shenbaise9527
 * @Create      : 2019-09-03 22:48:16
 * @Modified    : 2019-09-21 09:12:32
 * @version     : 1.0
 * @Description :
 */
package main

import (
	"errors"
	"time"
)

//Thread 帖子信息.
type Thread struct {
	ID        int       `gorm:"column:id;primary_key;auto_increment"`
	UUID      string    `gorm:"column:uuid"`
	Topic     string    `gorm:"column:topic"`
	UserID    int       `gorm:"column:user_id"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"`
}

//Post 回复信息.
type Post struct {
	ID        int       `gorm:"column:id;primary_key;auto_increment"`
	UUID      string    `gorm:"column:uuid"`
	Body      string    `gorm:"column:body"`
	UserID    int       `gorm:"column:user_id"`
	ThreadID  int       `gorm:"column:thread_id"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"`
}

//CreatedAtDate 获取帖子创建时间.
func (thr *Thread) CreatedAtDate() string {
	return thr.CreatedAt.Format("2006-01-02 15:04:05")
}

//NumReplies 获取帖子总的回复数.
func (thr *Thread) NumReplies() (count int) {
	db.Model(&Post{}).Where("thread_id = ?", thr.ID).Count(&count)

	return
}

//Posts 获取帖子的所有回复.
func (thr *Thread) Posts() (posts []Post, err error) {
	idb := db.Where("thread_id = ?", thr.ID).Find(&posts)
	err = idb.Error
	if idb.RecordNotFound() {
		logger.Debugf("cant find posts, thread_id: %d", thr.ID)
	} else if err != nil {
		logger.Errorf("failed to query posts, threadid: %d, err: %s", thr.ID, err)
	}

	return
}

//User 获取帖子的发起者.
func (thr *Thread) User() (user User) {
	user = queryUser(thr.UserID)

	return
}

// NewThread 创建帖子.
func (thr *Thread) NewThread() (err error) {
	err = db.Create(thr).Error
	if err != nil {
		logger.Errorf("Failed to insert threads: %s", err)

		return
	}

	logger.Debugf("new thread,id: %d, userid: %d, uuid: %s", thr.ID, thr.UserID, thr.UUID)

	return
}

// GetThreadByUUID 根据UUID获取帖子.
func (thr *Thread) GetThreadByUUID() (err error) {
	idb := db.Where("uuid = ?", thr.UUID).First(thr)
	err = idb.Error
	if idb.RecordNotFound() {
		logger.Debugf("cant find thread, uuid: %s", thr.UUID)
		err = errors.New("invalid thread uuid")
	} else if err != nil {
		logger.Errorf("Failed to query thread[uuid:%s]: %s", thr.UUID, err)
	}

	return
}

// NewPost 创建回复.
func (pst *Post) NewPost() (err error) {
	err = db.Create(pst).Error
	if err != nil {
		logger.Errorf("Failed to insert post: %s", err)

		return
	}

	logger.Debugf("new post,id: %d, userid: %d, threadid: %d",
		pst.ID, pst.UserID, pst.ThreadID)

	return
}

//CreatedAtDate 获取回复的时间.
func (pst *Post) CreatedAtDate() string {
	return pst.CreatedAt.Format("2006-01-02 15:04:05")
}

//User 获取回复的用户.
func (pst *Post) User() (user User) {
	user = queryUser(pst.UserID)

	return
}

//Threads 获取所有帖子.
func Threads() (threads []Thread, err error) {
	idb := db.Order("created_at desc").Find(&threads)
	if idb.RecordNotFound() {
		logger.Debug("cant find threads")
		err = nil
	} else {
		err = idb.Error
	}

	return
}

//queryUser 根据用户ID查询用户.
func queryUser(userid int) (user User) {
	user = User{}
	idb := db.Where("id = ?", userid).First(&user)
	err := idb.Error
	if idb.RecordNotFound() {
		logger.Debugf("cant find user: %d", userid)
	} else if err != nil {
		logger.Errorf("failed to query user, userid: %d, err: %s", userid, err)

		return
	}

	return
}
