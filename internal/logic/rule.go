// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"net/url"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RuleLogic struct{}

var DefaultRule = RuleLogic{}

func (RuleLogic) FindBy(ctx context.Context, conds map[string]string, curPage, limit int) ([]*model.CrawlRule, int) {
	objLog := GetLogger(ctx)

	filter := bson.M{}
	for k, v := range conds {
		filter[k] = v
	}

	coll := db.GetCollection("crawl_rule")

	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("rule find count error:", err)
		return nil, 0
	}

	offset := int64((curPage - 1) * limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetSkip(offset).
		SetLimit(int64(limit))

	ruleList := make([]*model.CrawlRule, 0)
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		objLog.Errorln("rule find error:", err)
		return nil, 0
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &ruleList); err != nil {
		objLog.Errorln("rule cursor all error:", err)
		return nil, 0
	}

	return ruleList, int(total)
}

func (RuleLogic) FindById(ctx context.Context, id string) *model.CrawlRule {
	objLog := GetLogger(ctx)

	rule := &model.CrawlRule{}
	err := db.GetCollection("crawl_rule").FindOne(ctx, bson.M{"_id": id}).Decode(rule)
	if err != nil {
		objLog.Errorln("find rule error:", err)
		return nil
	}

	if rule.Id == 0 {
		return nil
	}

	return rule
}

func (RuleLogic) Save(ctx context.Context, form url.Values, opUser string) (errMsg string, err error) {
	objLog := GetLogger(ctx)

	rule := &model.CrawlRule{}
	err = schemaDecoder.Decode(rule, form)
	if err != nil {
		objLog.Errorln("rule Decode error", err)
		errMsg = err.Error()
		return
	}

	rule.OpUser = opUser

	if rule.Id != 0 {
		_, err = db.GetCollection("crawl_rule").UpdateOne(ctx, bson.M{"_id": rule.Id}, bson.M{"$set": rule})
	} else {
		var id int
		id, err = db.NextID("crawl_rule")
		if err != nil {
			errMsg = "内部服务器错误"
			objLog.Errorln("rule nextid error:", err)
			return
		}
		rule.Id = id
		_, err = db.GetCollection("crawl_rule").InsertOne(ctx, rule)
	}

	if err != nil {
		errMsg = "内部服务器错误"
		objLog.Errorln("rule save:", errMsg, ":", err)
		return
	}

	return
}

func (RuleLogic) Delete(ctx context.Context, id string) error {
	_, err := db.GetCollection("crawl_rule").DeleteOne(ctx, bson.M{"_id": id})
	return err
}
