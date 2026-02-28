// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"errors"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FavoriteLogic struct{}

var DefaultFavorite = FavoriteLogic{}

func (FavoriteLogic) Save(ctx context.Context, uid, objid, objtype int) error {
	objLog := GetLogger(ctx)

	favorite := &model.Favorite{}
	favorite.Uid = uid
	favorite.Objid = objid
	favorite.Objtype = objtype

	_, err := db.GetCollection("favorites").InsertOne(ctx, favorite)
	if err != nil {
		objLog.Errorln("save favorite error:", err)
		return errors.New("内部服务错误")
	}

	return nil
}

func (FavoriteLogic) Cancel(ctx context.Context, uid, objid, objtype int) error {
	_, err := db.GetCollection("favorites").DeleteOne(ctx, bson.M{"uid": uid, "objtype": objtype, "objid": objid})
	return err
}

// HadFavorite 某个用户是否已经收藏某个对象
func (FavoriteLogic) HadFavorite(ctx context.Context, uid, objid, objtype int) int {
	objLog := GetLogger(ctx)

	favorite := &model.Favorite{}
	err := db.GetCollection("favorites").FindOne(ctx, bson.M{"uid": uid, "objid": objid, "objtype": objtype}).Decode(favorite)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			objLog.Errorln("FavoriteLogic HadFavorite error:", err)
		}
		return 0
	}

	if favorite.Uid != 0 {
		return 1
	}

	return 0
}

func (FavoriteLogic) FindUserFavorites(ctx context.Context, uid, objtype, start, rows int) ([]*model.Favorite, int64) {
	objLog := GetLogger(ctx)

	favorites := make([]*model.Favorite, 0)
	filter := bson.M{"uid": uid, "objtype": objtype}
	opts := options.Find().
		SetSort(bson.D{{"objid", -1}}).
		SetLimit(int64(rows)).
		SetSkip(int64(start))
	cursor, err := db.GetCollection("favorites").Find(ctx, filter, opts)
	if err != nil {
		objLog.Errorln("FavoriteLogic FindUserFavorites error:", err)
		return nil, 0
	}
	if err = cursor.All(ctx, &favorites); err != nil {
		objLog.Errorln("FavoriteLogic FindUserFavorites cursor error:", err)
		return nil, 0
	}

	total, err := db.GetCollection("favorites").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("FavoriteLogic FindUserFavorites count error:", err)
		return nil, 0
	}

	return favorites, total
}
