// Copyright 2017 The StudyGolang Authors. All rights reserved.
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

type GiftLogic struct{}

var DefaultGift = GiftLogic{}

func (self GiftLogic) FindAllOnline(ctx context.Context) []*model.Gift {
	objLog := GetLogger(ctx)

	gifts := make([]*model.Gift, 0)
	cursor, err := db.GetCollection("gift").Find(ctx, bson.M{"state": model.GiftStateOnline})
	if err != nil {
		objLog.Errorln("GiftLogic FindAllOnline error:", err)
		return nil
	}
	if err = cursor.All(ctx, &gifts); err != nil {
		objLog.Errorln("GiftLogic FindAllOnline cursor error:", err)
		return nil
	}

	for _, gift := range gifts {
		if gift.ExpireTime.Before(time.Now()) {
			gift.State = model.GiftStateExpired
			go self.doExpire(gift)
		}
	}

	return gifts
}

func (self GiftLogic) Exchange(ctx context.Context, me *model.Me, giftId int) error {
	objLog := GetLogger(ctx)

	gift := &model.Gift{}
	err := db.GetCollection("gift").FindOne(ctx, bson.M{"_id": giftId}).Decode(gift)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			objLog.Errorln("GiftLogic Exchange error:", err)
		}
		return err
	}

	if gift.RemainNum == 0 {
		return errors.New("已兑完")
	}

	total, err := db.GetCollection("user_exchange_record").CountDocuments(ctx, bson.M{"gift_id": giftId, "uid": me.Uid})
	if err != nil {
		objLog.Errorln("GiftLogic Count UserExchangeRecord error:", err)
		return err
	}

	if gift.BuyLimit <= int(total) {
		return errors.New("已兑换过")
	}

	if gift.Typ == model.GiftTypRedeem {
		return self.exchangeRedeem(gift, me)
	} else if gift.Typ == model.GiftTypDiscount {
		return self.exchangeDiscount(gift, me)
	}

	return nil
}

func (self GiftLogic) FindExchangeRecords(ctx context.Context, me *model.Me) []*model.UserExchangeRecord {
	objLog := GetLogger(ctx)

	records := make([]*model.UserExchangeRecord, 0)
	opts := options.Find().SetSort(bson.D{{"_id", -1}})
	cursor, err := db.GetCollection("user_exchange_record").Find(ctx, bson.M{"uid": me.Uid}, opts)
	if err != nil {
		objLog.Errorln("GiftLogic FindExchangeRecords error:", err)
		return nil
	}
	if err = cursor.All(ctx, &records); err != nil {
		objLog.Errorln("GiftLogic FindExchangeRecords cursor error:", err)
		return nil
	}

	return records
}

func (self GiftLogic) UserCanExchange(ctx context.Context, me *model.Me, gifts []*model.Gift) {
	num := len(gifts)
	if num == 0 {
		return
	}
	objLog := GetLogger(ctx)

	giftIds := make([]int, num)
	for i, gift := range gifts {
		giftIds[i] = gift.Id
	}

	exchangeRecords := make([]*model.UserExchangeRecord, 0)
	cursor, err := db.GetCollection("user_exchange_record").Find(ctx, bson.M{
		"gift_id": bson.M{"$in": giftIds},
		"uid":     me.Uid,
	})
	if err != nil {
		objLog.Errorln("GiftLogic FindUserGifts error:", err)
		return
	}
	if err = cursor.All(ctx, &exchangeRecords); err != nil {
		objLog.Errorln("GiftLogic FindUserGifts cursor error:", err)
		return
	}
	for _, record := range exchangeRecords {
		for _, gift := range gifts {
			if record.GiftId == gift.Id {
				gift.BuyLimit--
				break
			}
		}
	}
}

func (self GiftLogic) exchangeRedeem(gift *model.Gift, me *model.Me) error {
	ctx := context.Background()
	giftRedeem := &model.GiftRedeem{}
	err := db.GetCollection("gift_redeem").FindOne(ctx, bson.M{"gift_id": gift.Id, "exchange": 0}).Decode(giftRedeem)
	if err != nil {
		return err
	}

	if giftRedeem.Id == 0 {
		return errors.New("no more gift redeem")
	}

	return self.doExchange(gift, me, "兑换码："+giftRedeem.Code, func(sessCtx context.Context) error {
		_, err := db.GetCollection("gift_redeem").UpdateOne(sessCtx,
			bson.M{"_id": giftRedeem.Id, "exchange": 0},
			bson.M{"$set": bson.M{"exchange": 1, "uid": me.Uid}})
		return err
	})
}

func (self GiftLogic) exchangeDiscount(gift *model.Gift, me *model.Me) error {
	return self.doExchange(gift, me, "已兑换，我们会尽快联系合作方处理", nil)
}

func (self GiftLogic) doExchange(gift *model.Gift, me *model.Me, remark string, moreOp func(ctx context.Context) error) error {
	if me.Balance < gift.Price {
		return errors.New("兑换失败：铜币不够！")
	}

	ctx := context.Background()
	session, err := db.GetClient().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		exchangeRecord := &model.UserExchangeRecord{
			GiftId:     gift.Id,
			Uid:        me.Uid,
			Remark:     remark,
			ExpireTime: gift.ExpireTime,
		}
		id, err := db.NextID("user_exchange_record")
		if err != nil {
			return nil, err
		}
		exchangeRecord.Id = id

		_, err = db.GetCollection("user_exchange_record").InsertOne(sessCtx, exchangeRecord)
		if err != nil {
			return nil, err
		}

		if moreOp != nil {
			if err = moreOp(sessCtx); err != nil {
				return nil, err
			}
		}

		_, err = db.GetCollection("gift").UpdateOne(sessCtx, bson.M{"_id": gift.Id}, bson.M{"$inc": bson.M{"remain_num": -1}})
		if err != nil {
			return nil, err
		}

		desc := fmt.Sprintf("兑换 %s 消费 %d 铜币", gift.Name, gift.Price)
		err = DefaultMission.changeUserBalance(sessCtx, me, model.MissionTypeGift, -gift.Price, desc)
		return nil, err
	})

	return err
}

func (self GiftLogic) doExpire(gift *model.Gift) {
	ctx := context.Background()
	db.GetCollection("gift").UpdateOne(ctx, bson.M{"_id": gift.Id}, bson.M{"$set": bson.M{"state": gift.State}})
}
