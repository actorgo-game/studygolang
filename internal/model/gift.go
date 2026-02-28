// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

const (
	GiftStateOnline  = 1
	GiftStateExpired = 3

	GiftTypRedeem   = 0
	GiftTypDiscount = 1
)

var GiftTypeMap = map[int]string{
	GiftTypRedeem:   "兑换码",
	GiftTypDiscount: "折扣",
}

type Gift struct {
	Id          int       `json:"id" bson:"_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	Price       int       `bson:"price"`
	TotalNum    int       `bson:"total_num"`
	RemainNum   int       `bson:"remain_num"`
	ExpireTime  time.Time `bson:"expire_time"`
	Supplier    string    `bson:"supplier"`
	BuyLimit    int       `bson:"buy_limit"`
	Typ         int       `bson:"typ"`
	State       int       `bson:"state"`
	CreatedAt   OftenTime `bson:"created_at"`

	TypShow string `bson:"-"`
}

func (this *Gift) AfterLoad() {
	this.TypShow = GiftTypeMap[this.Typ]
}

type GiftRedeem struct {
	Id        int       `json:"id" bson:"_id"`
	GiftId    int       `bson:"gift_id"`
	Code      string    `bson:"code"`
	Exchange  int       `bson:"exchange"`
	Uid       int       `bson:"uid"`
	UpdatedAt OftenTime `bson:"updated_at"`
}

type UserExchangeRecord struct {
	Id         int       `json:"id" bson:"_id"`
	GiftId     int       `bson:"gift_id"`
	Uid        int       `bson:"uid"`
	Remark     string    `bson:"remark"`
	ExpireTime time.Time `bson:"expire_time"`
	CreatedAt  OftenTime `bson:"created_at"`
}
