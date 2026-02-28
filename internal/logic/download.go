// Copyright 2018 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"net/http"
	"strings"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/PuerkitoBio/goquery"
	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DownloadLogic struct{}

var DefaultDownload = DownloadLogic{}

func (DownloadLogic) FindAll(ctx context.Context) []*model.Download {
	downloads := make([]*model.Download, 0)
	opts := options.Find().SetSort(bson.D{{"seq", -1}})
	cursor, err := db.GetCollection("download").Find(ctx, bson.M{}, opts)
	if err != nil {
		logger.Errorln("DownloadLogic FindAll Error:", err)
		return downloads
	}
	if err = cursor.All(ctx, &downloads); err != nil {
		logger.Errorln("DownloadLogic FindAll cursor Error:", err)
	}

	return downloads
}

func (DownloadLogic) RecordDLTimes(ctx context.Context, filename string) error {
	_, err := db.GetCollection("download").UpdateOne(ctx,
		bson.M{"filename": filename},
		bson.M{"$inc": bson.M{"times": 1}})
	return err
}

func (DownloadLogic) AddNewDownload(ctx context.Context, version, selector string) error {
	objLog := GetLogger(ctx)

	resp, err := http.Get("https://golang.google.cn/dl/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return err
	}

	doc.Find(selector).Each(func(i int, versionSel *goquery.Selection) {
		idVal, exists := versionSel.Attr("id")
		if !exists {
			objLog.Errorln("add new download version not exist:", version)
			return
		}

		if idVal != version {
			objLog.Errorln("add new download version not match, expected:", version, "real:", idVal)
			return
		}

		downloads := make([]*model.Download, 0, 20)

		versionSel.Find("table tbody tr").Each(func(j int, dlSel *goquery.Selection) {
			download := &model.Download{
				Version: version,
			}

			if dlSel.HasClass("highlight") {
				download.IsRecommend = true
			}

			dlSel.Find("td").Each(func(k int, fieldSel *goquery.Selection) {
				val := fieldSel.Text()
				switch k {
				case 0:
					download.Filename = val
				case 1:
					download.Kind = val
				case 2:
					download.OS = val
				case 3:
					download.Arch = val
				case 4:
					download.Size = goutils.MustInt(strings.TrimRight(val, "MB"))
				case 5:
					download.Checksum = val
				}
			})

			if download.Kind == "" {
				objLog.Errorln("add new download Kind is empty:", version)
				return
			}

			count, err := db.GetCollection("download").CountDocuments(ctx, bson.M{"filename": download.Filename})
			if err != nil || count > 0 {
				return
			}

			downloads = append(downloads, download)
		})

		for i := len(downloads) - 1; i >= 0; i-- {
			id, err := db.NextID("download")
			if err != nil {
				objLog.Errorln("NextID download error:", err, "version:", version)
				continue
			}
			downloads[i].Id = id
			_, err = db.GetCollection("download").InsertOne(ctx, downloads[i])
			if err != nil {
				objLog.Errorln("insert download error:", err, "version:", version)
			}
		}

		// Set seq = _id for documents where seq is 0
		db.GetCollection("download").UpdateMany(ctx,
			bson.M{"seq": 0},
			mongo.Pipeline{{{"$set", bson.M{"seq": "$_id"}}}},
		)
	})

	return nil
}
