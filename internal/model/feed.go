// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import (
	"context"

	"github.com/polaris1119/config"
	"github.com/polaris1119/logger"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/studygolang/studygolang/db"
)

const FeedOffline = 1

type Feed struct {
	Id            int       `bson:"_id"`
	Title         string    `bson:"title"`
	Objid         int       `bson:"objid"`
	Objtype       int       `bson:"objtype"`
	Uid           int       `bson:"uid"`
	Author        string    `bson:"author"`
	Nid           int       `bson:"nid"`
	Lastreplyuid  int       `bson:"lastreplyuid"`
	Lastreplytime OftenTime `bson:"lastreplytime"`
	Tags          string    `bson:"tags"`
	Cmtnum        int       `bson:"cmtnum"`
	Likenum       int       `bson:"likenum"`
	Top           uint8     `bson:"top"`
	Seq           int       `bson:"seq"`
	State         int       `bson:"state"`
	CreatedAt     OftenTime `bson:"created_at"`
	UpdatedAt     OftenTime `json:"updated_at" bson:"updated_at"`

	User          *User                  `bson:"-"`
	Lastreplyuser *User                  `bson:"-"`
	Node          map[string]interface{} `bson:"-"`
	Uri           string                 `bson:"-"`
}

// PublishFeed 发布动态
func PublishFeed(object interface{}, objectExt interface{}, me *Me) {
	var feed *Feed
	switch objdoc := object.(type) {
	case *Topic:
		node := &TopicNode{}
		err := db.GetCollection("topics_node").FindOne(context.Background(), bson.M{"_id": objdoc.Nid}).Decode(node)
		if err == nil && !node.ShowIndex {
			return
		}

		cmtnum := 0
		if objectExt != nil {
			// 传递过来的是一个 *TopicEx 对象，类型是有的，即时值是 nil，这里也和 nil 是不等
			topicEx := objectExt.(*TopicEx)
			if topicEx != nil {
				cmtnum = topicEx.Reply
			}
		}

		feed = &Feed{
			Objid:         objdoc.Tid,
			Objtype:       TypeTopic,
			Title:         objdoc.Title,
			Uid:           objdoc.Uid,
			Tags:          objdoc.Tags,
			Cmtnum:        cmtnum,
			Nid:           objdoc.Nid,
			Top:           objdoc.Top,
			Lastreplyuid:  objdoc.Lastreplyuid,
			Lastreplytime: objdoc.Lastreplytime,
			UpdatedAt:     objdoc.Mtime,
		}
	case *Article:
		var uid int
		if objdoc.Domain == WebsiteSetting.Domain {
			userLogin := &UserLogin{}
			db.GetCollection("user_login").FindOne(context.Background(), bson.M{"username": objdoc.AuthorTxt}).Decode(userLogin)
			uid = userLogin.Uid
		}

		feed = &Feed{
			Objid:         objdoc.Id,
			Objtype:       TypeArticle,
			Title:         FilterTxt(objdoc.Title),
			Author:        objdoc.AuthorTxt,
			Uid:           uid,
			Tags:          objdoc.Tags,
			Cmtnum:        objdoc.Cmtnum,
			Top:           objdoc.Top,
			Lastreplyuid:  objdoc.Lastreplyuid,
			Lastreplytime: objdoc.Lastreplytime,
			UpdatedAt:     objdoc.Mtime,
		}
	case *Resource:
		cmtnum := 0
		if objectExt != nil {
			resourceEx := objectExt.(*ResourceEx)
			if resourceEx != nil {
				cmtnum = resourceEx.Cmtnum
			}
		}

		feed = &Feed{
			Objid:         objdoc.Id,
			Objtype:       TypeResource,
			Title:         objdoc.Title,
			Uid:           objdoc.Uid,
			Tags:          objdoc.Tags,
			Cmtnum:        cmtnum,
			Nid:           objdoc.Catid,
			Lastreplyuid:  objdoc.Lastreplyuid,
			Lastreplytime: objdoc.Lastreplytime,
			UpdatedAt:     objdoc.Mtime,
		}
	case *OpenProject:
		userLogin := &UserLogin{}
		db.GetCollection("user_login").FindOne(context.Background(), bson.M{"username": objdoc.Username}).Decode(userLogin)
		feed = &Feed{
			Objid:         objdoc.Id,
			Objtype:       TypeProject,
			Title:         objdoc.Category + " " + objdoc.Name,
			Author:        objdoc.Author,
			Uid:           userLogin.Uid,
			Tags:          objdoc.Tags,
			Cmtnum:        objdoc.Cmtnum,
			Lastreplyuid:  objdoc.Lastreplyuid,
			Lastreplytime: objdoc.Lastreplytime,
			UpdatedAt:     objdoc.Mtime,
		}
	case *Book:
		feed = &Feed{
			Objid:         objdoc.Id,
			Objtype:       TypeBook,
			Title:         "分享一本图书《" + objdoc.Name + "》",
			Uid:           objdoc.Uid,
			Tags:          objdoc.Tags,
			Cmtnum:        objdoc.Cmtnum,
			Lastreplyuid:  objdoc.Lastreplyuid,
			Lastreplytime: objdoc.Lastreplytime,
			UpdatedAt:     objdoc.UpdatedAt,
		}

		if me == nil {
			me = &Me{
				IsAdmin: true,
			}
		}
	}

	feedDay := config.ConfigFile.MustInt("feed", "day", 3)
	feed.Seq = feedDay * 24
	if me != nil && me.IsAdmin {
		feed.Seq += 100000
	}

	id, _ := db.NextID("feed")
	feed.Id = id
	_, err := db.GetCollection("feed").InsertOne(context.Background(), feed)
	if err != nil {
		logger.Errorln("publish feed:", object, " error:", err)
	}
}
