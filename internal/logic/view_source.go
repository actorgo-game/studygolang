// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"net/http"
	"strings"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/polaris1119/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ViewSourceLogic struct{}

var DefaultViewSource = ViewSourceLogic{}

// Record 记录浏览来源
func (ViewSourceLogic) Record(req *http.Request, objtype, objid int) {
	referer := req.Referer()
	if referer == "" || strings.Contains(referer, WebsiteSetting.Domain) {
		return
	}

	ctx := context.Background()

	viewSource := &model.ViewSource{}
	err := db.GetCollection("view_source").FindOne(ctx, bson.M{"objid": objid, "objtype": objtype}).Decode(viewSource)
	if err != nil && err != mongo.ErrNoDocuments {
		logger.Errorln("ViewSourceLogic Record find error:", err)
		return
	}

	if viewSource.Id == 0 {
		viewSource.Objid = objid
		viewSource.Objtype = objtype

		id, err := db.NextID("view_source")
		if err != nil {
			logger.Errorln("ViewSourceLogic Record NextID error:", err)
			return
		}
		viewSource.Id = id

		_, err = db.GetCollection("view_source").InsertOne(ctx, viewSource)
		if err != nil {
			logger.Errorln("ViewSourceLogic Record insert error:", err)
			return
		}
	}

	field := "other"
	referer = strings.ToLower(referer)
	ses := []string{"google", "baidu", "bing", "sogou", "so"}
	for _, se := range ses {
		if strings.Contains(referer, se+".") {
			field = se
			break
		}
	}

	_, err = db.GetCollection("view_source").UpdateOne(ctx, bson.M{"_id": viewSource.Id}, bson.M{"$inc": bson.M{field: 1}})
	if err != nil {
		logger.Errorln("ViewSourceLogic Record update error:", err)
		return
	}
}

// FindOne 获得浏览来源
func (ViewSourceLogic) FindOne(ctx context.Context, objid, objtype int) *model.ViewSource {
	objLog := GetLogger(ctx)

	viewSource := &model.ViewSource{}
	err := db.GetCollection("view_source").FindOne(ctx, bson.M{"objid": objid, "objtype": objtype}).Decode(viewSource)
	if err != nil && err != mongo.ErrNoDocuments {
		objLog.Errorln("ViewSourceLogic FindOne error:", err)
	}

	return viewSource
}
