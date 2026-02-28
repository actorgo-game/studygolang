// Copyright 2018 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import (
	"time"
)

// 微信绑定用户信息
type WechatUser struct {
	Id         int       `bson:"_id"`
	Openid     string    `bson:"openid"`
	Nickname   string    `bson:"nickname"`
	Avatar     string    `bson:"avatar"`
	SessionKey string    `bson:"session_key"`
	OpenInfo   string    `bson:"open_info"`
	Uid        int       `bson:"uid"`
	CreatedAt  time.Time `bson:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at"`
}

const (
	AutoReplyTypWord      = iota // 关键词回复
	AutoReplyTypNotFound         // 收到消息（未命中关键词且未搜索到）
	AutoReplyTypSubscribe        // 被关注回复
)

// WechatAutoReply 微信自动回复
type WechatAutoReply struct {
	Id        int       `bson:"_id"`
	Typ       uint8     `bson:"typ"`
	Word      string    `bson:"word"`
	MsgType   string    `bson:"msg_type"`
	Content   string    `bson:"content"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}
