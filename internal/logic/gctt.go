// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"time"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GCTTLogic struct{}

var DefaultGCTT = GCTTLogic{}

func (self GCTTLogic) FindTranslator(ctx context.Context, me *model.Me) *model.GCTTUser {
	objLog := GetLogger(ctx)

	gcttUser := &model.GCTTUser{}
	err := db.GetCollection("gctt_user").FindOne(ctx, bson.M{"uid": me.Uid}).Decode(gcttUser)
	if err != nil {
		objLog.Errorln("GCTTLogic FindTranslator error:", err)
		return nil
	}

	return gcttUser
}

func (self GCTTLogic) FindOne(ctx context.Context, username string) *model.GCTTUser {
	objLog := GetLogger(ctx)

	gcttUser := &model.GCTTUser{}
	err := db.GetCollection("gctt_user").FindOne(ctx, bson.M{"username": username}).Decode(gcttUser)
	if err != nil {
		objLog.Errorln("GCTTLogic FindOne error:", err)
		return nil
	}

	return gcttUser
}

func (self GCTTLogic) BindUser(ctx context.Context, gcttUser *model.GCTTUser, uid int, githubUser *model.BindUser) error {
	objLog := GetLogger(ctx)

	var err error

	if gcttUser.Id > 0 {
		gcttUser.Uid = uid
		_, err = db.GetCollection("gctt_user").UpdateOne(ctx, bson.M{"_id": gcttUser.Id}, bson.M{"$set": gcttUser})
	} else {
		gcttUser = &model.GCTTUser{
			Username: githubUser.Username,
			Avatar:   githubUser.Avatar,
			Uid:      uid,
			JoinedAt: time.Now().Unix(),
		}
		var id int
		id, err = db.NextID("gctt_user")
		if err != nil {
			objLog.Errorln("GCTTLogic BindUser NextID error:", err)
			return err
		}
		gcttUser.Id = id
		_, err = db.GetCollection("gctt_user").InsertOne(ctx, gcttUser)
	}

	if err != nil {
		objLog.Errorln("GCTTLogic BindUser error:", err)
	}

	return err
}

func (self GCTTLogic) FindCoreUsers(ctx context.Context) []*model.GCTTUser {
	objLog := GetLogger(ctx)

	gcttUsers := make([]*model.GCTTUser, 0)
	opts := options.Find().SetSort(bson.D{{Key: "role", Value: 1}})
	cursor, err := db.GetCollection("gctt_user").Find(ctx, bson.M{"role": bson.M{"$ne": model.GCTTRoleTranslator}}, opts)
	if err != nil {
		objLog.Errorln("GCTTLogic FindUsers error:", err)
		return gcttUsers
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &gcttUsers)

	return gcttUsers
}

func (self GCTTLogic) FindUsers(ctx context.Context) []*model.GCTTUser {
	objLog := GetLogger(ctx)

	gcttUsers := make([]*model.GCTTUser, 0)
	opts := options.Find().SetSort(bson.D{{Key: "num", Value: -1}, {Key: "words", Value: -1}})
	cursor, err := db.GetCollection("gctt_user").Find(ctx, bson.M{"num": bson.M{"$gt": 0}}, opts)
	if err != nil {
		objLog.Errorln("GCTTLogic FindUsers error:", err)
		return gcttUsers
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &gcttUsers)

	return gcttUsers
}

func (self GCTTLogic) FindUnTranslateIssues(ctx context.Context, limit int) []*model.GCTTIssue {
	objLog := GetLogger(ctx)

	gcttIssues := make([]*model.GCTTIssue, 0)
	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := db.GetCollection("gctt_issue").Find(ctx, bson.M{"state": model.IssueOpened}, opts)
	if err != nil {
		objLog.Errorln("GCTTLogic FindUnTranslateIssues error:", err)
		return gcttIssues
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &gcttIssues)

	return gcttIssues
}

func (self GCTTLogic) FindIssues(ctx context.Context, paginator *Paginator, querystring string, args ...interface{}) []*model.GCTTIssue {
	objLog := GetLogger(ctx)

	gcttIssues := make([]*model.GCTTIssue, 0)

	filter := bson.M{}
	if querystring != "" && len(args) > 0 {
		filter["label"] = args[0]
	}

	sortField := bson.D{{Key: "_id", Value: -1}}
	if len(args) > 0 && args[0] == model.LabelClaimed {
		sortField = bson.D{{Key: "translating_at", Value: -1}}
	}

	opts := options.Find().
		SetSort(sortField).
		SetSkip(int64(paginator.Offset())).
		SetLimit(int64(paginator.PerPage()))

	cursor, err := db.GetCollection("gctt_issue").Find(ctx, filter, opts)
	if err != nil {
		objLog.Errorln("GCTTLogic FindIssues error:", err)
		return gcttIssues
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &gcttIssues)

	return gcttIssues
}

func (self GCTTLogic) IssueCount(ctx context.Context, querystring string, args ...interface{}) int64 {
	objLog := GetLogger(ctx)

	filter := bson.M{}
	if querystring != "" && len(args) > 0 {
		filter["label"] = args[0]
	}

	total, err := db.GetCollection("gctt_issue").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("GCTTLogic Count error:", err)
	}

	return total
}

func (self GCTTLogic) FindNewestGit(ctx context.Context) []*model.GCTTGit {
	objLog := GetLogger(ctx)

	gcttGits := make([]*model.GCTTGit, 0)
	opts := options.Find().
		SetSort(bson.D{{Key: "translated_at", Value: -1}}).
		SetLimit(10)

	cursor, err := db.GetCollection("gctt_git").Find(ctx, bson.M{"translated_at": bson.M{"$ne": 0}}, opts)
	if err != nil {
		objLog.Errorln("GCTTLogic FindNewestGit error:", err)
		return gcttGits
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &gcttGits)

	return gcttGits
}

func (self GCTTLogic) FindTimeLines(ctx context.Context) []*model.GCTTTimeLine {
	objLog := GetLogger(ctx)

	gcttTimeLines := make([]*model.GCTTTimeLine, 0)
	cursor, err := db.GetCollection("gctt_timeline").Find(ctx, bson.M{})
	if err != nil {
		objLog.Errorln("GCTTLogic FindTimeLines error:", err)
		return gcttTimeLines
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &gcttTimeLines)

	return gcttTimeLines
}
