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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LearningMaterialLogic struct{}

var DefaultLearningMaterial = LearningMaterialLogic{}

func (LearningMaterialLogic) FindAll(ctx context.Context) []*model.LearningMaterial {
	objLog := GetLogger(ctx)

	learningMaterials := make([]*model.LearningMaterial, 0)
	opts := options.Find().SetSort(bson.D{{"type", 1}, {"seq", -1}})
	cursor, err := db.GetCollection("learning_material").Find(ctx, bson.M{}, opts)
	if err != nil {
		objLog.Errorln("LearningMaterialLogic FindAll error:", err)
		return nil
	}
	if err = cursor.All(ctx, &learningMaterials); err != nil {
		objLog.Errorln("LearningMaterialLogic FindAll cursor error:", err)
		return nil
	}

	return learningMaterials
}
