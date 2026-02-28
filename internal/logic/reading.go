// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/polaris1119/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReadingLogic struct{}

var DefaultReading = ReadingLogic{}

func (ReadingLogic) FindLastList(beginTime string) ([]*model.MorningReading, error) {
	ctx := context.Background()
	readings := make([]*model.MorningReading, 0)
	filter := bson.M{"ctime": bson.M{"$gt": beginTime}, "rtype": 0}
	opts := options.Find().SetSort(bson.D{{"_id", -1}})
	cursor, err := db.GetCollection("morning_reading").Find(ctx, filter, opts)
	if err != nil {
		return readings, err
	}
	err = cursor.All(ctx, &readings)
	return readings, err
}

// 获取晨读列表（分页）
func (ReadingLogic) FindBy(ctx context.Context, limit, rtype int, lastIds ...int) []*model.MorningReading {
	objLog := GetLogger(ctx)

	filter := bson.M{"rtype": rtype}
	if len(lastIds) > 0 && lastIds[0] > 0 {
		filter["_id"] = bson.M{"$lt": lastIds[0]}
	}

	readingList := make([]*model.MorningReading, 0)
	opts := options.Find().SetSort(bson.D{{"_id", -1}}).SetLimit(int64(limit))
	cursor, err := db.GetCollection("morning_reading").Find(ctx, filter, opts)
	if err != nil {
		objLog.Errorln("ResourceLogic FindReadings Error:", err)
		return nil
	}
	if err = cursor.All(ctx, &readingList); err != nil {
		objLog.Errorln("ResourceLogic FindReadings cursor Error:", err)
		return nil
	}

	return readingList
}

// 【我要晨读】
func (ReadingLogic) IReading(ctx context.Context, id int) string {
	objLog := GetLogger(ctx)

	reading := &model.MorningReading{}
	err := db.GetCollection("morning_reading").FindOne(ctx, bson.M{"_id": id}).Decode(reading)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			objLog.Errorln("reading logic IReading error:", err)
		}
		return "/readings"
	}

	if reading.Id == 0 {
		return "/readings"
	}

	go db.GetCollection("morning_reading").UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$inc": bson.M{"clicknum": 1}})

	if reading.Inner == 0 {
		return "/wr?u=" + reading.Url
	}

	return "/articles/" + strconv.Itoa(reading.Inner)
}

// FindReadingByPage 获取晨读列表（分页）
func (ReadingLogic) FindReadingByPage(ctx context.Context, conds map[string]string, curPage, limit int) ([]*model.MorningReading, int) {
	objLog := GetLogger(ctx)

	filter := bson.M{}
	for k, v := range conds {
		filter[k] = v
	}

	offset := (curPage - 1) * limit
	readingList := make([]*model.MorningReading, 0)
	opts := options.Find().
		SetSort(bson.D{{"_id", -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))
	cursor, err := db.GetCollection("morning_reading").Find(ctx, filter, opts)
	if err != nil {
		objLog.Errorln("reading find error:", err)
		return nil, 0
	}
	if err = cursor.All(ctx, &readingList); err != nil {
		objLog.Errorln("reading find cursor error:", err)
		return nil, 0
	}

	total, err := db.GetCollection("morning_reading").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("reading find count error:", err)
		return nil, 0
	}

	return readingList, int(total)
}

// SaveReading 保存晨读
func (ReadingLogic) SaveReading(ctx context.Context, form url.Values, username string) (errMsg string, err error) {
	reading := &model.MorningReading{}
	err = schemaDecoder.Decode(reading, form)
	if err != nil {
		logger.Errorln("reading SaveReading error", err)
		errMsg = err.Error()
		return
	}

	readings := make([]*model.MorningReading, 0)
	bgCtx := context.Background()
	if reading.Inner != 0 {
		reading.Url = ""
		cursor, findErr := db.GetCollection("morning_reading").Find(bgCtx,
			bson.M{"inner": reading.Inner},
			options.Find().SetSort(bson.D{{"_id", -1}}))
		if findErr == nil {
			cursor.All(bgCtx, &readings)
		}
	} else {
		cursor, findErr := db.GetCollection("morning_reading").Find(bgCtx,
			bson.M{"url": reading.Url},
			options.Find().SetSort(bson.D{{"_id", -1}}))
		if findErr == nil {
			cursor.All(bgCtx, &readings)
		}
	}

	reading.Moreurls = strings.TrimSpace(reading.Moreurls)
	if strings.Contains(reading.Moreurls, "\n") {
		reading.Moreurls = strings.Join(strings.Split(reading.Moreurls, "\n"), ",")
	}

	reading.Username = username

	logger.Debugln(reading.Rtype, "id=", reading.Id)
	if reading.Id != 0 {
		_, err = db.GetCollection("morning_reading").UpdateOne(bgCtx, bson.M{"_id": reading.Id}, bson.M{"$set": reading})
	} else {
		if len(readings) > 0 {
			logger.Errorln("reading report:", reading)
			errMsg, err = "已经存在了!!", errors.New("已经存在了!!")
			return
		}
		id, idErr := db.NextID("morning_reading")
		if idErr != nil {
			errMsg = "内部服务器错误"
			err = idErr
			return
		}
		reading.Id = id
		_, err = db.GetCollection("morning_reading").InsertOne(bgCtx, reading)
	}

	if err != nil {
		errMsg = "内部服务器错误"
		logger.Errorln("reading save:", errMsg, ":", err)
		return
	}

	return
}

// FindById 获取单条晨读
func (ReadingLogic) FindById(ctx context.Context, id int) *model.MorningReading {
	reading := &model.MorningReading{}
	err := db.GetCollection("morning_reading").FindOne(ctx, bson.M{"_id": id}).Decode(reading)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Errorln("reading logic FindReadingById Error:", err)
		}
		return nil
	}

	if reading.Id == 0 {
		return nil
	}

	return reading
}
