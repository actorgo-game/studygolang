// Copyright 2014 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import (
	"net/url"
	"time"
)

const (
	ProjectStatusNew     = 0
	ProjectStatusOnline  = 1
	ProjectStatusOffline = 2
)

// 开源项目信息
type OpenProject struct {
	Id            int       `json:"id" bson:"_id"`
	Name          string    `json:"name" bson:"name"`
	Category      string    `json:"category" bson:"category"`
	Uri           string    `json:"uri" bson:"uri"`
	Home          string    `json:"home" bson:"home"`
	Doc           string    `json:"doc" bson:"doc"`
	Download      string    `json:"download" bson:"download"`
	Src           string    `json:"src" bson:"src"`
	Logo          string    `json:"logo" bson:"logo"`
	Desc          string    `json:"desc" bson:"desc"`
	Repo          string    `json:"repo" bson:"repo"`
	Author        string    `json:"author" bson:"author"`
	Licence       string    `json:"licence" bson:"licence"`
	Lang          string    `json:"lang" bson:"lang"`
	Os            string    `json:"os" bson:"os"`
	Tags          string    `json:"tags" bson:"tags"`
	Username      string    `json:"username,omitempty" bson:"username"`
	Viewnum       int       `json:"viewnum,omitempty" bson:"viewnum"`
	Cmtnum        int       `json:"cmtnum,omitempty" bson:"cmtnum"`
	Likenum       int       `json:"likenum,omitempty" bson:"likenum"`
	Lastreplyuid  int       `json:"lastreplyuid" bson:"lastreplyuid"`
	Lastreplytime OftenTime `json:"lastreplytime" bson:"lastreplytime"`
	Status        int       `json:"status" bson:"status"`
	Ctime         OftenTime `json:"ctime,omitempty" bson:"ctime"`
	Mtime         OftenTime `json:"mtime,omitempty" bson:"mtime"`

	User *User `json:"user" bson:"-"`
	// 排行榜阅读量
	RankView      int   `json:"rank_view" bson:"-"`
	LastReplyUser *User `json:"last_reply_user" bson:"-"`
}

func (this *OpenProject) BeforeInsert() {
	if this.Tags == "" {
		this.Tags = AutoTag(this.Name+this.Category, this.Desc, 4)
	}

	this.Lastreplytime = NewOftenTime()
}

func (this *OpenProject) AfterInsert() {
	go func() {
		// AfterInsert 时，自增 ID 还未赋值，这里 sleep 一会，确保自增 ID 有值
		for {
			if this.Id > 0 {
				PublishFeed(this, nil, nil)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (this *OpenProject) AfterLoad() {
	if this.Logo == "" {
		this.Logo = WebsiteSetting.ProjectDfLogo
	}
	this.Uri = url.QueryEscape(this.Uri)
}
