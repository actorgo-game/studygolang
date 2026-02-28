// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

const (
	IsFreeFalse = iota
	IsFreeTrue
)

type Book struct {
	Id            int       `json:"id" bson:"_id"`
	Name          string    `json:"name" bson:"name"`
	Ename         string    `json:"ename" bson:"ename"`
	Cover         string    `json:"cover" bson:"cover"`
	Author        string    `json:"author" bson:"author"`
	Translator    string    `json:"translator" bson:"translator"`
	Lang          int       `json:"lang" bson:"lang"`
	PubDate       string    `json:"pub_date" bson:"pub_date"`
	Desc          string    `json:"desc" bson:"desc"`
	Tags          string    `json:"tags" bson:"tags"`
	Catalogue     string    `json:"catalogue" bson:"catalogue"`
	IsFree        bool      `json:"is_free" bson:"is_free"`
	OnlineUrl     string    `json:"online_url" bson:"online_url"`
	DownloadUrl   string    `json:"download_url" bson:"download_url"`
	BuyUrl        string    `json:"buy_url" bson:"buy_url"`
	Price         float32   `json:"price" bson:"price"`
	Lastreplyuid  int       `json:"lastreplyuid" bson:"lastreplyuid"`
	Lastreplytime OftenTime `json:"lastreplytime" bson:"lastreplytime"`
	Viewnum       int       `json:"viewnum" bson:"viewnum"`
	Cmtnum        int       `json:"cmtnum" bson:"cmtnum"`
	Likenum       int       `json:"likenum" bson:"likenum"`
	Uid           int       `json:"uid" bson:"uid"`
	CreatedAt     OftenTime `json:"created_at" bson:"created_at"`
	UpdatedAt     OftenTime `json:"updated_at" bson:"updated_at"`

	// 排行榜阅读量
	RankView int `json:"rank_view" bson:"-"`
}

func (this *Book) AfterInsert() {
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
