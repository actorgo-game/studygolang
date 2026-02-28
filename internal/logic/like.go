// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LikeLogic struct{}

var DefaultLike = LikeLogic{}

// HadLike 某个用户是否已经喜欢某个对象
func (LikeLogic) HadLike(ctx context.Context, uid, objid, objtype int) int {
	objLog := GetLogger(ctx)

	like := &model.Like{}
	err := db.GetCollection("likes").FindOne(ctx, bson.M{"uid": uid, "objid": objid, "objtype": objtype}).Decode(like)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			objLog.Errorln("LikeLogic HadLike error:", err)
		}
		return 0
	}

	return like.Flag
}

// FindUserLikeObjects 获取用户对一批对象是否喜欢的状态
// objids 两个值
func (LikeLogic) FindUserLikeObjects(ctx context.Context, uid, objtype int, objids ...int) (map[int]int, error) {
	objLog := GetLogger(ctx)

	if len(objids) < 2 {
		return nil, errors.New("参数错误")
	}

	littleId, greatId := objids[0], objids[1]
	if littleId > greatId {
		littleId, greatId = greatId, littleId
	}

	likes := make([]*model.Like, 0)
	filter := bson.M{
		"uid":     uid,
		"objtype": objtype,
		"objid":   bson.M{"$gte": littleId, "$lte": greatId},
	}
	cursor, err := db.GetCollection("likes").Find(ctx, filter)
	if err != nil {
		objLog.Errorln("LikeLogic FindUserLikeObjects error:", err)
		return nil, err
	}
	if err = cursor.All(ctx, &likes); err != nil {
		objLog.Errorln("LikeLogic FindUserLikeObjects cursor error:", err)
		return nil, err
	}

	likeFlags := make(map[int]int, len(likes))
	for _, like := range likes {
		likeFlags[like.Objid] = like.Flag
	}

	return likeFlags, nil
}

// FindUserRecentLikes 获取用户最近喜欢的对象（不过滤对象）
func (LikeLogic) FindUserRecentLikes(ctx context.Context, uid, limit int) (map[int]map[int]int, error) {
	objLog := GetLogger(ctx)

	likes := make([]*model.Like, 0)
	filter := bson.M{
		"uid":   uid,
		"ctime": bson.M{"$gt": time.Now().Add(-7 * 24 * time.Hour)},
	}
	opts := options.Find().SetLimit(int64(limit))
	cursor, err := db.GetCollection("likes").Find(ctx, filter, opts)
	if err != nil {
		objLog.Errorln("LikeLogic FindUserRecentLikes error:", err)
		return nil, err
	}
	if err = cursor.All(ctx, &likes); err != nil {
		objLog.Errorln("LikeLogic FindUserRecentLikes cursor error:", err)
		return nil, err
	}

	likeFlags := make(map[int]map[int]int, len(likes))
	for _, like := range likes {
		if _, ok := likeFlags[like.Objid]; ok {
			likeFlags[like.Objid][like.Objtype] = like.Flag
		} else {
			likeFlags[like.Objid] = map[int]int{
				like.Objtype: like.Flag,
			}
		}
	}

	return likeFlags, nil
}

// LikeObject 喜欢或取消喜欢
// objid 注册的喜欢对象
// uid 喜欢的人
func (LikeLogic) LikeObject(ctx context.Context, uid, objid, objtype, likeFlag int) error {
	objLog := GetLogger(ctx)

	like := &model.Like{}
	err := db.GetCollection("likes").FindOne(ctx, bson.M{"uid": uid, "objid": objid, "objtype": objtype}).Decode(like)
	if err != nil && err != mongo.ErrNoDocuments {
		objLog.Errorln("LikeLogic LikeObject get error:", err)
		return err
	}

	// 之前喜欢过
	if like.Uid != 0 {
		// 再喜欢直接返回
		if likeFlag == model.FlagLike {
			return nil
		}

		// 取消喜欢
		if likeFlag == model.FlagCancel {
			_, err = db.GetCollection("likes").DeleteOne(ctx, bson.M{"uid": uid, "objid": objid, "objtype": objtype})
			if err != nil {
				objLog.Errorln("LikeLogic LikeObject delete error:", err)
				return err
			}

			// 取消喜欢成功，更新对象的喜欢数
			if liker, ok := likers[objtype]; ok {
				go liker.UpdateLike(objid, -1)

				DefaultFeed.updateLike(objid, objtype, uid, -1)
			}

			return nil
		}

		return nil
	}

	like.Uid = uid
	like.Objid = objid
	like.Objtype = objtype
	like.Flag = likeFlag

	_, err = db.GetCollection("likes").InsertOne(ctx, like)
	if err != nil {
		objLog.Errorln("LikeLogic LikeObject error:", err)
		return err
	}

	// 喜欢成功
	if liker, ok := likers[objtype]; ok {
		go liker.UpdateLike(objid, 1)

		DefaultFeed.updateLike(objid, objtype, uid, 1)
	}

	go likeObservable.NotifyObservers(uid, objtype, objid)

	// TODO: 给被喜欢对象所有者发系统消息
	/*
		ext := map[string]interface{}{
			"objid":   objid,
			"objtype": objtype,
			"cid":     cid,
			"uid":     uid,
		}
		go SendSystemMsgTo(0, objtype, ext)
	*/

	return nil
}

var likers = make(map[int]Liker)

// 喜欢接口
type Liker interface {
	fmt.Stringer
	// 喜欢 回调接口，用于更新对象自身需要更新的数据
	UpdateLike(int, int)
}

// 注册喜欢对象，使得某种类型（主题、博文等）被喜欢了可以回调
func RegisterLikeObject(objtype int, liker Liker) {
	if liker == nil {
		panic("logic: Register liker is nil")
	}
	if _, dup := likers[objtype]; dup {
		panic("logic: Register called twice for liker " + liker.String())
	}
	likers[objtype] = liker
}
