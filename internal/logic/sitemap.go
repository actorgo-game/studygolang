// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/util"

	"github.com/polaris1119/config"
	"github.com/polaris1119/logger"

	"github.com/studygolang/studygolang/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var funcMap = template.FuncMap{
	"time_format": func(i interface{}) string {
		if t, ok := i.(time.Time); ok {
			return t.Format(time.RFC3339)
		} else if t, ok := i.(model.OftenTime); ok {
			return time.Time(t).Format(time.RFC3339)
		}
		return ""
	},
}

var sitemapTpl = template.Must(template.New("sitemap.xml").Funcs(funcMap).ParseFiles(config.TemplateDir + "/sitemap.xml"))
var sitemapIndexTpl = template.Must(template.ParseFiles(config.TemplateDir + "/sitemapindex.xml"))

var sitemapPath = config.ROOT + "/sitemap/"

func init() {
	if !util.Exist(sitemapPath) {
		err := os.MkdirAll(sitemapPath, 0777)
		if err != nil {
			panic(err)
		}
	}
}

func GenSitemap() {
	sitemapFiles := []string{}
	ctx := context.Background()

	loc := "http://" + WebsiteSetting.Domain
	if WebsiteSetting.OnlyHttps {
		loc = "https://" + WebsiteSetting.Domain
	}
	home := map[string]string{
		"loc":      loc,
		"lastmode": time.Now().Format(time.RFC3339),
	}

	var (
		little = 1
		step   = 4999
		large  = little + step
	)

	// 文章
	var (
		articles = make([]*model.Article, 0)
		err      error
	)
	for {
		sitemapFile := "sitemap_article_" + strconv.Itoa(large) + ".xml"

		filter := bson.M{
			"_id":    bson.M{"$gte": little, "$lte": large},
			"status": bson.M{"$ne": model.ArticleStatusOffline},
		}
		opts := options.Find().SetProjection(bson.M{"_id": 1, "mtime": 1})
		cursor, findErr := db.GetCollection("articles").Find(ctx, filter, opts)
		little = large + 1
		large = little + step

		if findErr != nil {
			continue
		}
		err = cursor.All(ctx, &articles)
		cursor.Close(ctx)
		if err != nil {
			continue
		}

		if len(articles) == 0 {
			break
		}

		data := map[string]interface{}{
			"home":     home,
			"articles": articles,
		}

		if err = output(sitemapFile, data); err == nil {
			sitemapFiles = append(sitemapFiles, sitemapFile)
		}

		articles = make([]*model.Article, 0)
	}

	little = 1
	large = little + step

	// 主题（帖子）
	topics := make([]*model.Topic, 0)
	for {
		sitemapFile := "sitemap_topic_" + strconv.Itoa(large) + ".xml"

		filter := bson.M{
			"_id":  bson.M{"$gte": little, "$lte": large},
			"flag": bson.M{"$in": []int{0, 1}},
		}
		opts := options.Find().SetProjection(bson.M{"_id": 1, "mtime": 1})
		cursor, findErr := db.GetCollection("topics").Find(ctx, filter, opts)
		little = large + 1
		large = little + step

		if findErr != nil {
			continue
		}
		err = cursor.All(ctx, &topics)
		cursor.Close(ctx)
		if err != nil {
			continue
		}

		if len(topics) == 0 {
			break
		}

		data := map[string]interface{}{
			"home":   home,
			"topics": topics,
		}

		if err = output(sitemapFile, data); err == nil {
			sitemapFiles = append(sitemapFiles, sitemapFile)
		}

		topics = make([]*model.Topic, 0)
	}

	little = 1
	large = little + step

	// 资源
	resources := make([]*model.Resource, 0)
	for {
		sitemapFile := "sitemap_resource_" + strconv.Itoa(large) + ".xml"

		filter := bson.M{
			"_id": bson.M{"$gte": little, "$lte": large},
		}
		opts := options.Find().SetProjection(bson.M{"_id": 1, "mtime": 1})
		cursor, findErr := db.GetCollection("resource").Find(ctx, filter, opts)
		little = large + 1
		large = little + step

		if findErr != nil {
			logger.Errorln("sitemap resource find error:", findErr)
			continue
		}
		err = cursor.All(ctx, &resources)
		cursor.Close(ctx)
		if err != nil {
			continue
		}

		if len(resources) == 0 {
			break
		}

		data := map[string]interface{}{
			"home":      home,
			"resources": resources,
		}

		if err = output(sitemapFile, data); err == nil {
			sitemapFiles = append(sitemapFiles, sitemapFile)
		}

		resources = make([]*model.Resource, 0)
	}

	little = 1
	large = little + step

	// 项目
	projects := make([]*model.OpenProject, 0)
	for {
		sitemapFile := "sitemap_project_" + strconv.Itoa(large) + ".xml"

		filter := bson.M{
			"_id": bson.M{"$gte": little, "$lte": large},
		}
		opts := options.Find().SetProjection(bson.M{"_id": 1, "uri": 1, "mtime": 1})
		cursor, findErr := db.GetCollection("open_project").Find(ctx, filter, opts)
		little = large + 1
		large = little + step

		if findErr != nil {
			continue
		}
		err = cursor.All(ctx, &projects)
		cursor.Close(ctx)
		if err != nil {
			continue
		}

		if len(projects) == 0 {
			break
		}

		data := map[string]interface{}{
			"home":     home,
			"projects": projects,
		}

		if err = output(sitemapFile, data); err == nil {
			sitemapFiles = append(sitemapFiles, sitemapFile)
		}

		projects = make([]*model.OpenProject, 0)
	}

	little = 1
	large = little + step

	// 图书
	books := make([]*model.Book, 0)
	for {
		sitemapFile := "sitemap_book_" + strconv.Itoa(large) + ".xml"

		filter := bson.M{
			"_id": bson.M{"$gte": little, "$lte": large},
		}
		opts := options.Find().SetProjection(bson.M{"_id": 1, "updated_at": 1})
		cursor, findErr := db.GetCollection("book").Find(ctx, filter, opts)
		little = large + 1
		large = little + step

		if findErr != nil {
			continue
		}
		err = cursor.All(ctx, &books)
		cursor.Close(ctx)
		if err != nil {
			continue
		}

		if len(books) == 0 {
			break
		}

		data := map[string]interface{}{
			"home":  home,
			"books": books,
		}

		if err = output(sitemapFile, data); err == nil {
			sitemapFiles = append(sitemapFiles, sitemapFile)
		}

		books = make([]*model.Book, 0)
	}

	little = 1
	large = little + step

	// wiki
	wikis := make([]*model.Wiki, 0)
	for {
		sitemapFile := "sitemap_wiki_" + strconv.Itoa(large) + ".xml"

		filter := bson.M{
			"_id": bson.M{"$gte": little, "$lte": large},
		}
		opts := options.Find().SetProjection(bson.M{"_id": 1, "uri": 1, "mtime": 1})
		cursor, findErr := db.GetCollection("wiki").Find(ctx, filter, opts)
		little = large + 1
		large = little + step

		if findErr != nil {
			continue
		}
		err = cursor.All(ctx, &wikis)
		cursor.Close(ctx)
		if err != nil {
			continue
		}

		if len(wikis) == 0 {
			break
		}

		data := map[string]interface{}{
			"home":  home,
			"wikis": wikis,
		}

		if err = output(sitemapFile, data); err == nil {
			sitemapFiles = append(sitemapFiles, sitemapFile)
		}

		wikis = make([]*model.Wiki, 0)
	}

	file, err := os.Create(sitemapPath + "sitemapindex.xml")
	if err != nil {
		logger.Errorln("gen sitemap index file error:", err)
		return
	}
	defer file.Close()

	err = sitemapIndexTpl.Execute(file, map[string]interface{}{
		"home":         home,
		"sitemapFiles": sitemapFiles,
	})
	if err != nil {
		logger.Errorln("execute sitemap index template error:", err)
	}
}

func output(filename string, data map[string]interface{}) (err error) {
	var file *os.File
	file, err = os.Create(sitemapPath + filename)
	if err != nil {
		logger.Errorln("open file error:", err)
		return
	}
	defer file.Close()
	if err = sitemapTpl.Execute(file, data); err != nil {
		logger.Errorln("execute template error:", err)
	}

	return
}
