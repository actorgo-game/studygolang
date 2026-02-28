// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

var BalanceTypeMap = map[int]string{
	MissionTypeLogin:    "每日登录奖励",
	MissionTypeInitial:  "初始资本",
	MissionTypeShare:    "分享获得",
	MissionTypeAdd:      "充值获得",
	MissionTypeReply:    "创建回复",
	MissionTypeTopic:    "创建主题",
	MissionTypeArticle:  "发表文章",
	MissionTypeResource: "分享资源",
	MissionTypeWiki:     "创建WIKI",
	MissionTypeProject:  "发布项目",
	MissionTypeBook:     "分享图书",
	MissionTypeAppend:   "增加附言",
	MissionTypeTop:      "置顶",
	MissionTypeModify:   "修改",
	MissionTypeReplied:  "回复收益",
	MissionTypeAward:    "额外赠予",
	MissionTypeActive:   "活跃奖励",
	MissionTypeGift:     "兑换物品",
	MissionTypePunish:   "处罚",
	MissionTypeSpam:     "Spam",
}

type UserBalanceDetail struct {
	Id        int       `json:"id" bson:"_id"`
	Uid       int       `json:"uid" bson:"uid"`
	Type      int       `json:"type" bson:"type"`
	Num       int       `json:"num" bson:"num"`
	Balance   int       `json:"balance" bson:"balance"`
	Desc      string    `json:"desc" bson:"desc"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`

	TypeShow string `json:"type_show" bson:"-"`
}

func (this *UserBalanceDetail) AfterLoad() {
	this.TypeShow = BalanceTypeMap[this.Type]
}

type UserRecharge struct {
	Id        int       `json:"id" bson:"_id"`
	Uid       int       `json:"uid" bson:"uid"`
	Amount    int       `json:"amount" bson:"amount"`
	Channel   string    `json:"channel" bson:"channel"`
	Remark    string    `json:"remark" bson:"remark"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
