// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

// 不要修改常量的顺序
const (
	TypeTopic     = iota // 主题
	TypeArticle          // 博文
	TypeResource         // 资源
	TypeWiki             // WIKI
	TypeProject          // 开源项目
	TypeBook             // 图书
	TypeInterview        // 面试题
)

const (
	TypeComment = 100
	// 置顶
	TypeTop = 101
)

const (
	TopicURI    = "topics"
	ArticleURI  = "articles"
	ResourceURI = "resources"
	WikiURI     = "wiki"
	ProjectURI  = "p"
	BookURI     = "book"
)

var PathUrlMap = map[int]string{
	TypeTopic:     "/topics/",
	TypeArticle:   "/articles/",
	TypeResource:  "/resources/",
	TypeWiki:      "/wiki/",
	TypeProject:   "/p/",
	TypeBook:      "/book/",
	TypeInterview: "/interview/",
}

var TypeNameMap = map[int]string{
	TypeTopic:     "主题",
	TypeArticle:   "博文",
	TypeResource:  "资源",
	TypeWiki:      "Wiki",
	TypeProject:   "项目",
	TypeBook:      "图书",
	TypeInterview: "面试题",
}

// 评论信息（通用）
type Comment struct {
	Cid     int       `json:"cid" bson:"_id"`
	Objid   int       `json:"objid" bson:"objid"`
	Objtype int       `json:"objtype" bson:"objtype"`
	Content string    `json:"content" bson:"content"`
	Uid     int       `json:"uid" bson:"uid"`
	Floor   int       `json:"floor" bson:"floor"`
	Flag    int       `json:"flag" bson:"flag"`
	Ctime   OftenTime `json:"ctime" bson:"ctime"`

	Objinfo    map[string]interface{} `json:"objinfo" bson:"-"`
	ReplyFloor int                    `json:"reply_floor" bson:"-"`
}

func (*Comment) CollectionName() string {
	return "comments"
}
