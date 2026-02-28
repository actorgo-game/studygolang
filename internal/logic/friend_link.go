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

type FriendLinkLogic struct{}

var DefaultFriendLink = FriendLinkLogic{}

func (FriendLinkLogic) FindAll(ctx context.Context, limits ...int) []*model.FriendLink {
	objLog := GetLogger(ctx)

	friendLinks := make([]*model.FriendLink, 0)
	opts := options.Find().SetSort(bson.D{{"seq", 1}})
	if len(limits) > 0 {
		opts.SetLimit(int64(limits[0]))
	}
	cursor, err := db.GetCollection("friend_link").Find(ctx, bson.M{}, opts)
	if err != nil {
		objLog.Errorln("FriendLinkLogic FindAll error:", err)
		return nil
	}
	if err = cursor.All(ctx, &friendLinks); err != nil {
		objLog.Errorln("FriendLinkLogic FindAll cursor error:", err)
		return nil
	}

	return friendLinks
}
