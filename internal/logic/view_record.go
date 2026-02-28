// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/polaris1119/logger"
	"go.mongodb.org/mongo-driver/bson"
)

type ViewRecordLogic struct{}

var DefaultViewRecord = ViewRecordLogic{}

func (ViewRecordLogic) Record(objid, objtype, uid int) {
	ctx := context.Background()

	total, err := db.GetCollection("view_record").CountDocuments(ctx, bson.M{"objid": objid, "objtype": objtype, "uid": uid})
	if err != nil {
		logger.Errorln("ViewRecord logic Record count error:", err)
		return
	}

	if total > 0 {
		return
	}

	viewRecord := &model.ViewRecord{
		Objid:   objid,
		Objtype: objtype,
		Uid:     uid,
	}

	id, err := db.NextID("view_record")
	if err != nil {
		logger.Errorln("ViewRecord logic Record NextID error:", err)
		return
	}
	viewRecord.Id = id

	if _, err = db.GetCollection("view_record").InsertOne(ctx, viewRecord); err != nil {
		logger.Errorln("ViewRecord logic Record insert Error:", err)
		return
	}

	return
}

func (ViewRecordLogic) FindUserNum(ctx context.Context, objid, objtype int) int64 {
	objLog := GetLogger(ctx)

	total, err := db.GetCollection("view_record").CountDocuments(ctx, bson.M{"objid": objid, "objtype": objtype})
	if err != nil {
		objLog.Errorln("ViewRecordLogic FindUserNum error:", err)
	}

	return total
}
