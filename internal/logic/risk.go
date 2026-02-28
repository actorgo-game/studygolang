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

	"github.com/polaris1119/nosql"
	"go.mongodb.org/mongo-driver/bson"
)

type RiskLogic struct{}

var DefaultRisk = RiskLogic{}

// AddBlackIP 加入 IP 黑名单
func (RiskLogic) AddBlackIP(ip string) error {
	redisClient := nosql.NewRedisClient()
	defer redisClient.Close()

	key := "black:ip"
	return redisClient.HSET(key, ip, "1")
}

// AddBlackIPByUID 通过用户 UID 将最后一次登录 IP 加入黑名单
func (self RiskLogic) AddBlackIPByUID(uid int) error {
	ctx := context.Background()
	userLogin := &model.UserLogin{}
	err := db.GetCollection("user_login").FindOne(ctx, bson.M{"_id": uid}).Decode(userLogin)
	if err != nil {
		return err
	}

	if userLogin.LoginIp != "" {
		return self.AddBlackIP(userLogin.LoginIp)
	}

	return nil
}

// IsBlackIP 是否是 IP 黑名单
func (RiskLogic) IsBlackIP(ip string) bool {
	redisClient := nosql.NewRedisClient()
	defer redisClient.Close()

	key := "black:ip"
	val, err := redisClient.HGET(key, ip)
	if err != nil {
		return false
	}

	return val == "1"
}
