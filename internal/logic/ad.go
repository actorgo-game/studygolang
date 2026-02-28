// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/polaris1119/set"
	"go.mongodb.org/mongo-driver/bson"
)

type AdLogic struct{}

var DefaultAd = AdLogic{}

func (AdLogic) FindAll(ctx context.Context, path string) map[string]*model.Advertisement {
	objLog := GetLogger(ctx)

	pageAds := make([]*model.PageAd, 0)
	filter := bson.M{
		"$or": []bson.M{
			{"path": path},
			{"path": "*"},
		},
		"is_online": 1,
	}
	cursor, err := db.GetCollection("page_ad").Find(ctx, filter)
	if err != nil {
		objLog.Errorln("AdLogic FindAll PageAd error:", err)
		return nil
	}
	if err = cursor.All(ctx, &pageAds); err != nil {
		objLog.Errorln("AdLogic FindAll PageAd cursor error:", err)
		return nil
	}

	adIdSet := set.New(set.NonThreadSafe)
	for _, pageAd := range pageAds {
		adIdSet.Add(pageAd.AdId)
	}

	if adIdSet.IsEmpty() {
		return nil
	}

	adList := make([]*model.Advertisement, 0)
	cursor, err = db.GetCollection("advertisement").Find(ctx, bson.M{"_id": bson.M{"$in": set.IntSlice(adIdSet)}})
	if err != nil {
		objLog.Errorln("AdLogic FindAll Advertisement error:", err)
		return nil
	}
	if err = cursor.All(ctx, &adList); err != nil {
		objLog.Errorln("AdLogic FindAll Advertisement cursor error:", err)
		return nil
	}

	adMap := make(map[int]*model.Advertisement, len(adList))
	for _, ad := range adList {
		adMap[ad.Id] = ad
	}

	posAdsMap := make(map[string]*model.Advertisement, len(pageAds))
	for _, pageAd := range pageAds {
		if adMap[pageAd.AdId].IsOnline {
			posAdsMap[pageAd.Position] = adMap[pageAd.AdId]
		}
	}

	return posAdsMap
}
