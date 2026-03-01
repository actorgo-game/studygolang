// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

const (
	LinkForm    = "只是链接"
	ContentForm = "包括内容"
)

// 资源信息
type Resource struct {
	Id            int       `json:"id" bson:"_id"`
	Title         string    `json:"title" bson:"title"`
	Form          string    `json:"form" bson:"form"`
	Content       string    `json:"content" bson:"content"`
	Url           string    `json:"url" bson:"url"`
	Uid           int       `json:"uid" bson:"uid"`
	Catid         int       `json:"catid" bson:"catid"`
	CatName       string    `json:"-" bson:"-"`
	Lastreplyuid  int       `json:"lastreplyuid" bson:"lastreplyuid"`
	Lastreplytime OftenTime `json:"lastreplytime" bson:"lastreplytime"`
	Tags          string    `json:"tags" bson:"tags"`
	Ctime         OftenTime `json:"ctime" bson:"ctime"`
	Mtime         OftenTime `json:"mtime" bson:"mtime"`

	// 排行榜阅读量
	RankView int `json:"rank_view" bson:"-"`
}

func (this *Resource) BeforeInsert() {
	if this.Tags == "" {
		this.Tags = AutoTag(this.Title+this.CatName, this.Content, 4)
	}

	this.Lastreplytime = NewOftenTime()
	now := OftenTime(time.Now())
	if time.Time(this.Ctime).IsZero() {
		this.Ctime = now
	}
	if time.Time(this.Mtime).IsZero() {
		this.Mtime = now
	}
}

// 资源扩展（计数）信息
type ResourceEx struct {
	Id      int       `json:"-" bson:"_id"`
	Viewnum int       `json:"viewnum" bson:"viewnum"`
	Cmtnum  int       `json:"cmtnum" bson:"cmtnum"`
	Likenum int       `json:"likenum" bson:"likenum"`
	Mtime   time.Time `json:"mtime" bson:"mtime"`
}

type ResourceInfo struct {
	Resource
	ResourceEx
}

func (*ResourceInfo) CollectionName() string {
	return "resource"
}

// 资源分类信息
type ResourceCat struct {
	Catid int    `json:"catid" bson:"_id"`
	Name  string `json:"name" bson:"name"`
	Intro string `json:"intro" bson:"intro"`
	Ctime string `json:"ctime" bson:"ctime"`
}

func (*ResourceCat) CollectionName() string {
	return "resource_category"
}
