// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"
	"github.com/studygolang/studygolang/util"

	"github.com/garyburd/redigo/redis"
	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"github.com/polaris1119/nosql"
	"github.com/polaris1119/times"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

var (
	beginAwardWeight = 50
)

type UserRichLogic struct{}

var DefaultUserRich = UserRichLogic{}

func (self UserRichLogic) AwardCooper() {
	redisClient := nosql.NewRedisClient()
	defer redisClient.Close()
	ymd := times.Format("ymd", time.Now().Add(-86400*time.Second))
	key := DefaultRank.getDAURankKey(ymd)

	var (
		cursor      uint64
		err         error
		resultSlice []interface{}
		count       = 20
	)

	for {
		cursor, resultSlice, err = redisClient.ZSCAN(key, cursor, "COUNT", count)
		if err != nil {
			logger.Errorln("AwardCooper ZSCAN error:", err)
			break
		}

		for len(resultSlice) > 0 {
			var (
				uid, weight int
				err         error
			)
			resultSlice, err = redis.Scan(resultSlice, &uid, &weight)
			if err != nil {
				logger.Errorln("AwardCooper redis Scan error:", err)
				continue
			}

			if weight < beginAwardWeight {
				continue
			}

			award := util.Max((weight-500)*5, 0) +
				util.UMin((weight-400), 100)*4 +
				util.UMin((weight-300), 100)*3 +
				util.UMin((weight-200), 100)*2 +
				util.UMin((weight-100), 100) +
				int(float64(util.UMin((weight-beginAwardWeight), beginAwardWeight))*0.5)

			userRank := redisClient.ZREVRANK(key, uid)
			desc := fmt.Sprintf("%s 的活跃度为 %d，排名第 %d，奖励 %d 铜币", ymd, weight, userRank, award)

			user := DefaultUser.FindOne(nil, "uid", uid)
			self.IncrUserRich(user, model.MissionTypeActive, award, desc)
		}

		if cursor == 0 {
			break
		}
	}
}

// IncrUserRich 增加或减少用户财富
func (self UserRichLogic) IncrUserRich(user *model.User, typ, award int, desc string) {
	if award == 0 {
		logger.Errorln("IncrUserRich, but award is empty!")
		return
	}

	ctx := context.Background()

	var (
		total int64 = -1
		err   error
	)

	if award > 0 && (typ == model.MissionTypeReplied || typ == model.MissionTypeActive) {
		total, err = db.GetCollection("user_balance_detail").CountDocuments(ctx, bson.M{"uid": user.Uid})
		if err != nil {
			logger.Errorln("IncrUserRich count error:", err)
			return
		}
	}

	session, sessionErr := db.GetClient().StartSession()
	if sessionErr != nil {
		logger.Errorln("IncrUserRich start session error:", sessionErr)
		return
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		var initialAward int
		if total == 0 {
			var autoErr error
			initialAward, autoErr = self.autoCompleteInitial(sessCtx, user)
			if autoErr != nil {
				logger.Errorln("IncrUserRich autoCompleteInitial error:", autoErr)
				return nil, autoErr
			}
		}

		user.Balance += initialAward + award
		if user.Balance < 0 {
			user.Balance = 0
		}
		_, txErr := db.GetCollection("user_info").UpdateOne(sessCtx,
			bson.M{"_id": user.Uid},
			bson.M{"$set": bson.M{"balance": user.Balance}})
		if txErr != nil {
			logger.Errorln("IncrUserRich update error:", txErr)
			return nil, txErr
		}

		balanceDetail := &model.UserBalanceDetail{
			Uid:     user.Uid,
			Type:    typ,
			Num:     award,
			Balance: user.Balance,
			Desc:    desc,
		}
		balanceDetail.Id, txErr = db.NextID("user_balance_detail")
		if txErr != nil {
			logger.Errorln("IncrUserRich NextID error:", txErr)
			return nil, txErr
		}
		_, txErr = db.GetCollection("user_balance_detail").InsertOne(sessCtx, balanceDetail)
		if txErr != nil {
			logger.Errorln("IncrUserRich insert error:", txErr)
			return nil, txErr
		}

		return nil, nil
	})

	if err != nil {
		logger.Errorln("IncrUserRich transaction error:", err)
	}
}

func (UserRichLogic) FindBalanceDetail(ctx context.Context, me *model.Me, p int, types ...int) []*model.UserBalanceDetail {
	objLog := GetLogger(ctx)

	filter := bson.M{"uid": me.Uid}
	if len(types) > 0 {
		filter["type"] = types[0]
	}

	balanceDetails := make([]*model.UserBalanceDetail, 0)
	cursor, err := db.GetCollection("user_balance_detail").Find(ctx, filter,
		options.Find().
			SetSort(bson.M{"_id": -1}).
			SetSkip(int64((p-1)*CommentPerNum)).
			SetLimit(int64(CommentPerNum)))
	if err != nil {
		objLog.Errorln("UserRichLogic FindBalanceDetail error:", err)
		return nil
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &balanceDetails); err != nil {
		objLog.Errorln("UserRichLogic FindBalanceDetail decode error:", err)
		return nil
	}

	for _, detail := range balanceDetails {
		detail.AfterLoad()
	}

	return balanceDetails
}

func (UserRichLogic) Total(ctx context.Context, uid int) int64 {
	total, err := db.GetCollection("user_balance_detail").CountDocuments(ctx, bson.M{"uid": uid})
	if err != nil {
		logger.Errorln("UserRichLogic Total error:", err)
	}
	return total
}

func (self UserRichLogic) FindRecharge(ctx context.Context, me *model.Me) int {
	objLog := GetLogger(ctx)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"uid": me.Uid}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "total": bson.M{"$sum": "$amount"}}}},
	}
	cursor, err := db.GetCollection("user_recharge").Aggregate(ctx, pipeline)
	if err != nil {
		objLog.Errorln("UserRichLogic FindRecharge error:", err)
		return 0
	}
	defer cursor.Close(ctx)

	var result struct {
		Total int `bson:"total"`
	}
	if cursor.Next(ctx) {
		if err = cursor.Decode(&result); err != nil {
			objLog.Errorln("UserRichLogic FindRecharge decode error:", err)
			return 0
		}
	}

	return result.Total
}

// Recharge 用户充值
func (self UserRichLogic) Recharge(ctx context.Context, uid string, form url.Values) {
	objLog := GetLogger(ctx)

	createdAt, _ := time.ParseInLocation("2006-01-02 15:04:05", form.Get("time"), time.Local)
	userRecharge := &model.UserRecharge{
		Uid:       goutils.MustInt(uid),
		Amount:    goutils.MustInt(form.Get("amount")),
		Channel:   form.Get("channel"),
		CreatedAt: createdAt,
	}

	session, sessionErr := db.GetClient().StartSession()
	if sessionErr != nil {
		objLog.Errorln("Recharge start session error:", sessionErr)
		return
	}
	defer session.EndSession(ctx)

	_, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		rechargeID, txErr := db.NextID("user_recharge")
		if txErr != nil {
			return nil, txErr
		}
		userRecharge.Id = rechargeID
		_, txErr = db.GetCollection("user_recharge").InsertOne(sessCtx, userRecharge)
		if txErr != nil {
			objLog.Errorln("UserRichLogic Recharge error:", txErr)
			return nil, txErr
		}

		user := DefaultUser.FindOne(ctx, "uid", uid)
		me := &model.Me{
			Uid:     user.Uid,
			Balance: user.Balance,
		}

		award := goutils.MustInt(form.Get("copper"))
		desc := fmt.Sprintf("%s 充值 ￥%d，获得 %d 个铜币", times.Format("Ymd"), userRecharge.Amount, award)
		txErr = DefaultMission.changeUserBalance(sessCtx, me, model.MissionTypeAdd, award, desc)
		if txErr != nil {
			objLog.Errorln("UserRichLogic changeUserBalance error:", txErr)
			return nil, txErr
		}

		return nil, nil
	})

	if err != nil {
		objLog.Errorln("UserRichLogic Recharge transaction error:", err)
	}
}

func (UserRichLogic) add(ctx context.Context, balanceDetail *model.UserBalanceDetail) error {
	if balanceDetail.Id == 0 {
		var err error
		balanceDetail.Id, err = db.NextID("user_balance_detail")
		if err != nil {
			return err
		}
	}
	_, err := db.GetCollection("user_balance_detail").InsertOne(ctx, balanceDetail)
	return err
}

func (UserRichLogic) autoCompleteInitial(ctx context.Context, user *model.User) (int, error) {
	mission := &model.Mission{}
	err := db.GetCollection("mission").FindOne(ctx, bson.M{"_id": model.InitialMissionId}).Decode(mission)
	if err != nil {
		return 0, err
	}
	if mission.Id == 0 {
		return 0, errors.New("初始资本任务不存在！")
	}

	balanceDetail := &model.UserBalanceDetail{
		Uid:     user.Uid,
		Type:    model.MissionTypeInitial,
		Num:     mission.Fixed,
		Balance: mission.Fixed,
		Desc:    fmt.Sprintf("获得%s %d 铜币", model.BalanceTypeMap[mission.Type], mission.Fixed),
	}
	balanceDetail.Id, err = db.NextID("user_balance_detail")
	if err != nil {
		return 0, err
	}
	_, err = db.GetCollection("user_balance_detail").InsertOne(ctx, balanceDetail)

	return mission.Fixed, err
}
