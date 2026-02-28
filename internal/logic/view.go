// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/polaris1119/config"
	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"go.mongodb.org/mongo-driver/bson"
)

type view struct {
	objtype int
	objid   int

	num    int
	locker sync.Mutex
}

func newView(objtype, objid int) *view {
	return &view{objtype: objtype, objid: objid}
}

func (this *view) incr() {
	this.locker.Lock()
	defer this.locker.Unlock()

	this.num++
}

// flush 将浏览数刷入数据库中
func (this *view) flush() {
	this.locker.Lock()
	defer this.locker.Unlock()

	ctx := context.Background()
	switch this.objtype {
	case model.TypeTopic:
		db.GetCollection("topics_ex").UpdateOne(ctx, bson.M{"_id": this.objid}, bson.M{"$inc": bson.M{"view": this.num}})
	case model.TypeArticle:
		db.GetCollection("articles").UpdateOne(ctx, bson.M{"_id": this.objid}, bson.M{"$inc": bson.M{"viewnum": this.num}})
	case model.TypeResource:
		db.GetCollection("resource_ex").UpdateOne(ctx, bson.M{"_id": this.objid}, bson.M{"$inc": bson.M{"viewnum": this.num}})
	case model.TypeProject:
		db.GetCollection("open_project").UpdateOne(ctx, bson.M{"_id": this.objid}, bson.M{"$inc": bson.M{"viewnum": this.num}})
	case model.TypeWiki:
		db.GetCollection("wiki").UpdateOne(ctx, bson.M{"_id": this.objid}, bson.M{"$inc": bson.M{"viewnum": this.num}})
	case model.TypeBook:
		db.GetCollection("book").UpdateOne(ctx, bson.M{"_id": this.objid}, bson.M{"$inc": bson.M{"viewnum": this.num}})
	case model.TypeInterview:
		db.GetCollection("interview_question").UpdateOne(ctx, bson.M{"_id": this.objid}, bson.M{"$inc": bson.M{"viewnum": this.num}})
	}

	DefaultRank.GenDayRank(this.objtype, this.objid, this.num)

	this.num = 0
}

type views struct {
	data  map[string]*view
	users map[string]bool

	locker sync.Mutex
}

func newViews() *views {
	return &views{data: make(map[string]*view), users: make(map[string]bool)}
}

// TODO: 用户登录了，应该用用户标识，而不是IP
func (this *views) Incr(req *http.Request, objtype, objid int, uids ...int) {
	ua := req.UserAgent()
	spiders := config.ConfigFile.MustValueArray("global", "spider", ",")
	for _, spider := range spiders {
		if strings.Contains(ua, spider) {
			return
		}
	}

	go DefaultViewSource.Record(req, objtype, objid)

	key := strconv.Itoa(objtype) + strconv.Itoa(objid)

	var userKey string

	if len(uids) > 0 {
		userKey = fmt.Sprintf("%s_uid_%d", key, uids[0])
	} else {
		userKey = fmt.Sprintf("%s_ip_%d", key, goutils.Ip2long(goutils.RemoteIp(req)))
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	if _, ok := this.users[userKey]; ok {
		return
	} else {
		this.users[userKey] = true
	}

	if _, ok := this.data[key]; !ok {
		this.data[key] = newView(objtype, objid)
	}

	this.data[key].incr()

	if len(uids) > 0 {
		ViewObservable.NotifyObservers(uids[0], objtype, objid)
	} else {
		ViewObservable.NotifyObservers(0, objtype, objid)
	}
}

func (this *views) Flush() {
	logger.Debugln("start views flush")
	this.locker.Lock()
	defer this.locker.Unlock()

	for _, view := range this.data {
		view.flush()
	}

	this.data = make(map[string]*view)
	this.users = make(map[string]bool)

	logger.Debugln("end views flush")
}

var Views = newViews()
