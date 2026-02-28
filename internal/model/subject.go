// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import (
	"time"
)

// Subject 专栏
type Subject struct {
	Id          int       `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Cover       string    `json:"cover" bson:"cover"`
	Description string    `json:"description" bson:"description"`
	Uid         int       `json:"uid" bson:"uid"`
	Contribute  bool      `json:"contribute" bson:"contribute"`
	Audit       bool      `json:"audit" bson:"audit"`
	ArticleNum  int       `json:"article_num" bson:"article_num"`
	CreatedAt   OftenTime `json:"created_at" bson:"created_at"`
	UpdatedAt   OftenTime `json:"updated_at" bson:"updated_at"`

	User *User `json:"user" bson:"-"`
}

// SubjectAdmin 专栏管理员
type SubjectAdmin struct {
	Id        int       `json:"id" bson:"_id"`
	Sid       int       `json:"sid" bson:"sid"`
	Uid       int       `json:"uid" bson:"uid"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

const (
	ContributeStateNew = iota
	ContributeStateOnline
	ContributeStateOffline
)

// SubjectArticle 专栏文章
type SubjectArticle struct {
	Id        int       `json:"id" bson:"_id"`
	Sid       int       `json:"sid" bson:"sid"`
	ArticleId int       `json:"article_id" bson:"article_id"`
	State     int       `json:"state" bson:"state"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// SubjectArticles join 需要
type SubjectArticles struct {
	Article
	Sid       int       `bson:"sid"`
	CreatedAt time.Time `bson:"created_at"`
}

func (*SubjectArticles) CollectionName() string {
	return "articles"
}

// SubjectFollower 专栏关注者
type SubjectFollower struct {
	Id        int       `json:"id" bson:"_id"`
	Sid       int       `json:"sid" bson:"sid"`
	Uid       int       `json:"uid" bson:"uid"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`

	User    *User  `bson:"-"`
	TimeAgo string `bson:"-"`
}
