/**
 * Copyright(c) 2019 Shenzhen shenbaise9527
 * All Rights Reserved
 * @File        : thread.go
 * @Author      : shenbaise9527
 * @Create      : 2019-09-03 22:48:16
 * @Modified    : 2019-09-10 14:32:25
 * @version     : 1.0
 * @Description :
 */
package main

import (
	"time"
)

//Thread 帖子信息.
type Thread struct {
	ID        int
	UUID      string
	Topic     string
	UserID    int
	CreatedAt time.Time
}

//Post 回复信息.
type Post struct {
	ID        int
	UUID      string
	Body      string
	UserID    int
	ThreadID  int
	CreatedAt time.Time
}

//CreatedAtDate 获取帖子创建时间.
func (thr *Thread) CreatedAtDate() string {
	return thr.CreatedAt.Format("2006-01-02 15:04:05")
}

//NumReplies 获取帖子总的回复数.
func (thr *Thread) NumReplies() (count int) {
	rows, err := db.Query("select count(*) from posts where thread_id=$1", thr.ID)
	if err != nil {
		logger.Errorf("Failed to query numreplies: %s", err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			logger.Errorf("Failed to scan numreplies: %s", err)
			return
		}
	}

	return
}

//Posts 获取帖子的所有回复.
func (thr *Thread) Posts() (posts []Post, err error) {
	rows, err := db.Query("select id, uuid, body, user_id, thread_id, created_at from posts where thread_id=$1", thr.ID)
	if err != nil {
		logger.Errorf("failed to query posts: %s", err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		pst := Post{}
		if err = rows.Scan(&pst.ID, &pst.UUID, &pst.Body, &pst.UserID, &pst.ThreadID, &pst.CreatedAt); err != nil {
			logger.Errorf("failed to scan posts: %s", err)
			return
		}

		posts = append(posts, pst)
	}

	return
}

//User 获取帖子的发起者.
func (thr *Thread) User() (user User) {
	user = queryUser(thr.UserID)

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
	rows, err := db.Query("select id, uuid, topic, user_id, created_at from threads order by created_at desc")
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		conv := Thread{}
		err = rows.Scan(&conv.ID, &conv.UUID, &conv.Topic, &conv.UserID, &conv.CreatedAt)
		if err != nil {
			return
		}

		threads = append(threads, conv)
	}

	return
}

//queryUser 根据用户ID查询用户.
func queryUser(userid int) (user User) {
	user = User{}
	rows, err := db.Query("select id, uuid, name, email, created_at from users where id = ?", userid)
	if err != nil {
		logger.Errorf("failed to query: %s", err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.ID, &user.UUID, &user.Name, &user.Email, &user.CreatedAt)
		if err != nil {
			logger.Errorf("failed to scan: %s", err)
			return
		}

		break
	}

	return
}
