// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/polaris1119/logger"
)

const (
	ArticleStatusNew = iota
	ArticleStatusOnline
	ArticleStatusOffline
)

var LangSlice = []string{"中文", "英文"}
var ArticleStatusSlice = []string{"未上线", "已上线", "已下线"}

// 抓取的文章信息
type Article struct {
	Id            int       `json:"id" bson:"_id"`
	Domain        string    `json:"domain" bson:"domain"`
	Name          string    `json:"name" bson:"name"`
	Title         string    `json:"title" bson:"title"`
	Cover         string    `json:"cover" bson:"cover"`
	Author        string    `json:"author" bson:"author"`
	AuthorTxt     string    `json:"author_txt" bson:"author_txt"`
	Lang          int       `json:"lang" bson:"lang"`
	PubDate       string    `json:"pub_date" bson:"pub_date"`
	Url           string    `json:"url" bson:"url"`
	Content       string    `json:"content" bson:"content"`
	Txt           string    `json:"txt" bson:"txt"`
	Tags          string    `json:"tags" bson:"tags"`
	Css           string    `json:"css" bson:"css"`
	Viewnum       int       `json:"viewnum" bson:"viewnum"`
	Cmtnum        int       `json:"cmtnum" bson:"cmtnum"`
	Likenum       int       `json:"likenum" bson:"likenum"`
	Lastreplyuid  int       `json:"lastreplyuid" bson:"lastreplyuid"`
	Lastreplytime OftenTime `json:"lastreplytime" bson:"lastreplytime"`
	Top           uint8     `json:"top" bson:"top"`
	Markdown      bool      `json:"markdown" bson:"markdown"`
	GCTT          bool      `json:"gctt" bson:"gctt"`
	CloseReply    bool      `json:"close_reply" bson:"close_reply"`
	Status        int       `json:"status" bson:"status"`
	OpUser        string    `json:"op_user" bson:"op_user"`
	Ctime         OftenTime `json:"ctime" bson:"ctime"`
	Mtime         OftenTime `json:"mtime" bson:"mtime"`

	IsSelf bool  `json:"is_self" bson:"-"`
	User   *User `json:"-" bson:"-"`
	// 排行榜阅读量
	RankView      int   `json:"rank_view" bson:"-"`
	LastReplyUser *User `json:"last_reply_user" bson:"-"`
}

func (this *Article) AfterLoad() {
	this.IsSelf = strconv.Itoa(this.Id) == this.Url
}

func (this *Article) BeforeInsert() {
	if this.Tags == "" {
		this.Tags = AutoTag(this.Title, this.Txt, 4)
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

func (this *Article) AfterInsert() {
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

func (*Article) CollectionName() string {
	return "articles"
}

type ArticleGCTT struct {
	ArticleID  int    `bson:"_id"`
	Author     string `bson:"author"`
	AuthorURL  string `bson:"author_url"`
	Translator string `bson:"translator"`
	Checker    string `bson:"checker"`
	URL        string `bson:"url"`

	Avatar   string   `bson:"-"`
	Checkers []string `bson:"-"`
}

func (*ArticleGCTT) CollectionName() string {
	return "article_gctt"
}

func (this *ArticleGCTT) AfterLoad() {
	this.Checkers = strings.Split(this.Checker, ",")
}

// 抓取网站文章的规则
type CrawlRule struct {
	Id      int    `json:"id" bson:"_id"`
	Domain  string `json:"domain" bson:"domain"`
	Subpath string `json:"subpath" bson:"subpath"`
	Lang    int    `json:"lang" bson:"lang"`
	Name    string `json:"name" bson:"name"`
	Title   string `json:"title" bson:"title"`
	Author  string `json:"author" bson:"author"`
	InUrl   bool   `json:"in_url" bson:"in_url"`
	PubDate string `json:"pub_date" bson:"pub_date"`
	Content string `json:"content" bson:"content"`
	Ext     string `json:"ext" bson:"ext"`
	OpUser  string `json:"op_user" bson:"op_user"`
	Ctime   string `json:"ctime" bson:"ctime"`
}

func (this *CrawlRule) ParseExt() map[string]string {
	if this.Ext == "" {
		return nil
	}

	extMap := make(map[string]string)
	err := json.Unmarshal([]byte(this.Ext), &extMap)
	if err != nil {
		logger.Errorln("parse crawl rule ext error:", err)
		return nil
	}

	return extMap
}

const (
	AutoCrawlOn = 0
	AutoCrawOff = 1
)

// 网站自动抓取规则
type AutoCrawlRule struct {
	Id             int    `json:"id" bson:"_id"`
	Website        string `json:"website" bson:"website"`
	AllUrl         string `json:"all_url" bson:"all_url"`
	IncrUrl        string `json:"incr_url" bson:"incr_url"`
	Keywords       string `json:"keywords" bson:"keywords"`
	ListSelector   string `json:"list_selector" bson:"list_selector"`
	ResultSelector string `json:"result_selector" bson:"result_selector"`
	PageField      string `json:"page_field" bson:"page_field"`
	MaxPage        int    `json:"max_page" bson:"max_page"`
	Ext            string `json:"ext" bson:"ext"`
	OpUser         string `json:"op_user" bson:"op_user"`
	Mtime          string `json:"mtime" bson:"mtime"`

	ExtMap map[string]string `json:"-" bson:"-"`
}

func (this *AutoCrawlRule) AfterLoad() {
	if this.Ext == "" {
		return
	}

	this.ExtMap = make(map[string]string)
	err := json.Unmarshal([]byte(this.Ext), &this.ExtMap)
	if err != nil {
		logger.Errorln("parse auto crawl rule ext error:", err)
		return
	}
}
