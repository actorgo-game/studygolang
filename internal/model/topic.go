// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

const (
	FlagNoAudit = iota
	FlagNormal
	FlagAuditDelete
	FlagUserDelete
)

const (
	// 最多附言条数
	AppendMaxNum = 3
)

const (
	PermissionPublic = iota // 公开
	PermissionLogin         // 登录可见
	PermissionFollow        // 关注可见（暂未实现）
	PermissionPay           // 知识星球或其他方式付费可见
	PermissionOnlyMe        // 自己可见
)

// 社区主题信息
type Topic struct {
	Tid           int       `json:"tid" bson:"_id"`
	Title         string    `json:"title" bson:"title"`
	Content       string    `json:"content" bson:"content"`
	Nid           int       `json:"nid" bson:"nid"`
	Uid           int       `json:"uid" bson:"uid"`
	Flag          uint8     `json:"flag" bson:"flag"`
	Lastreplyuid  int       `json:"lastreplyuid" bson:"lastreplyuid"`
	Lastreplytime OftenTime `json:"lastreplytime" bson:"lastreplytime"`
	EditorUid     int       `json:"editor_uid" bson:"editor_uid"`
	Top           uint8     `json:"top" bson:"top"`
	TopTime       int64     `json:"top_time" bson:"top_time"`
	Tags          string    `json:"tags" bson:"tags"`
	Permission    int       `json:"permission" bson:"permission"`
	CloseReply    bool      `json:"close_reply" bson:"close_reply"`
	Ctime         OftenTime `json:"ctime" bson:"ctime"`
	Mtime         OftenTime `json:"mtime" bson:"mtime"`

	// 为了方便，加上Node（节点名称，数据表没有）
	Node string `bson:"-"`
	// 排行榜阅读量
	RankView int `json:"rank_view" bson:"-"`
}

func (*Topic) CollectionName() string {
	return "topics"
}

func (this *Topic) BeforeInsert() {
	if this.Tags == "" {
		this.Tags = AutoTag(this.Title, this.Content, 4)
	}
}

// 社区主题扩展（计数）信息
type TopicEx struct {
	Tid   int       `json:"-" bson:"tid"`
	View  int       `json:"viewnum" bson:"view"`
	Reply int       `json:"cmtnum" bson:"reply"`
	Like  int       `json:"likenum" bson:"like"`
	Mtime time.Time `json:"mtime" bson:"mtime"`
}

func (*TopicEx) CollectionName() string {
	return "topics_ex"
}

// 社区主题扩展（计数）信息，用于 incr 更新
type TopicUpEx struct {
	Tid   int       `json:"-" bson:"_id"`
	View  int       `json:"viewnum" bson:"view"`
	Reply int       `json:"cmtnum" bson:"reply"`
	Like  int       `json:"likenum" bson:"like"`
	Mtime time.Time `json:"mtime" bson:"mtime"`
}

func (*TopicUpEx) CollectionName() string {
	return "topics_ex"
}

type TopicInfo struct {
	Topic
	TopicEx
}

func (*TopicInfo) CollectionName() string {
	return "topics"
}

type TopicAppend struct {
	Id        int       `bson:"_id"`
	Tid       int       `bson:"tid"`
	Content   string    `bson:"content"`
	CreatedAt OftenTime `bson:"created_at"`
}

// 社区主题节点信息
type TopicNode struct {
	Nid       int       `json:"nid" bson:"_id"`
	Parent    int       `json:"parent" bson:"parent"`
	Logo      string    `json:"logo" bson:"logo"`
	Name      string    `json:"name" bson:"name"`
	Ename     string    `json:"ename" bson:"ename"`
	Seq       int       `json:"seq" bson:"seq"`
	Intro     string    `json:"intro" bson:"intro"`
	ShowIndex bool      `json:"show_index" bson:"show_index"`
	Ctime     time.Time `json:"ctime" bson:"ctime"`

	Level int `json:"-" bson:"-"`
}

func (*TopicNode) CollectionName() string {
	return "topics_node"
}

// 推荐节点
type RecommendNode struct {
	Id        int       `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	Parent    int       `json:"parent" bson:"parent"`
	Nid       int       `json:"nid" bson:"nid"`
	Seq       int       `json:"seq" bson:"seq"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type NodeInfo struct {
	RecommendNode
	TopicNode
}

func (*NodeInfo) CollectionName() string {
	return "recommend_node"
}
