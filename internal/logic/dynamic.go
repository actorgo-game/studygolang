// Copyright 2016 The StudyGolang Authors. All rights reserved.
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
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DynamicLogic struct{}

var DefaultDynamic = DynamicLogic{}

// FindBy 获取动态列表（分页）
func (DynamicLogic) FindBy(ctx context.Context, lastId int, limit int) []*model.Dynamic {
	dynamicList := make([]*model.Dynamic, 0)
	opts := options.Find().SetSort(bson.D{{"seq", -1}}).SetLimit(int64(limit))
	cursor, err := db.GetCollection("dynamic").Find(ctx, bson.M{"_id": bson.M{"$gt": lastId}}, opts)
	if err != nil {
		logger.Errorln("DynamicLogic FindBy Error:", err)
		return dynamicList
	}
	if err = cursor.All(ctx, &dynamicList); err != nil {
		logger.Errorln("DynamicLogic FindBy cursor Error:", err)
	}

	return dynamicList
}
