// Copyright 2014 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package logic

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/util"

	"github.com/polaris1119/config"
	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"github.com/polaris1119/set"

	"github.com/studygolang/studygolang/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SearcherLogic struct {
	maxRows int

	engineUrl string
}

var DefaultSearcher = SearcherLogic{maxRows: 100, engineUrl: config.ConfigFile.MustValue("search", "engine_url")}

func (self SearcherLogic) Indexing(isAll bool) {
	go self.IndexingOpenProject(isAll)
	go self.IndexingTopic(isAll)
	go self.IndexingResource(isAll)
	self.IndexingArticle(isAll)
}

// IndexingArticle 索引博文
func (self SearcherLogic) IndexingArticle(isAll bool) {
	solrClient := NewSolrClient()

	ctx := context.Background()

	var (
		articleList []*model.Article
		err         error
	)

	id := 0
	for {
		articleList = make([]*model.Article, 0)

		var filter bson.M
		if isAll {
			filter = bson.M{"_id": bson.M{"$gt": id}}
		} else {
			timeAgo := time.Now().Add(-2 * time.Minute)
			filter = bson.M{"mtime": bson.M{"$gt": timeAgo}}
		}

		opts := options.Find().
			SetSort(bson.D{{Key: "_id", Value: 1}}).
			SetLimit(int64(self.maxRows))

		cursor, findErr := db.GetCollection("articles").Find(ctx, filter, opts)
		if findErr != nil {
			logger.Errorln("IndexingArticle error:", findErr)
			break
		}
		err = cursor.All(ctx, &articleList)
		cursor.Close(ctx)
		if err != nil {
			logger.Errorln("IndexingArticle cursor error:", err)
			break
		}

		if len(articleList) == 0 {
			break
		}

		for _, article := range articleList {
			logger.Infoln("deal article_id:", article.Id)

			if id < article.Id {
				id = article.Id
			}

			if article.Tags == "" {
				article.Tags = model.AutoTag(article.Title, article.Txt, 4)
				if article.Tags != "" {
					db.GetCollection("articles").UpdateOne(ctx, bson.M{"_id": article.Id}, bson.M{"$set": bson.M{"tags": article.Tags}})
				}
			}

			document := model.NewDocument(article, nil)
			if article.Status != model.ArticleStatusOffline {
				solrClient.PushAdd(model.NewDefaultArgsAddCommand(document))
			} else {
				solrClient.PushDel(model.NewDelCommand(document))
			}
		}

		solrClient.Post()

		if !isAll {
			break
		}
	}
}

func (self SearcherLogic) IndexingTopic(isAll bool) {
	solrClient := NewSolrClient()

	ctx := context.Background()

	var (
		topicList []*model.Topic
		err       error
	)

	id := 0
	for {
		topicList = make([]*model.Topic, 0)

		var filter bson.M
		if isAll {
			filter = bson.M{"_id": bson.M{"$gt": id}}
		} else {
			timeAgo := time.Now().Add(-2 * time.Minute)
			filter = bson.M{"mtime": bson.M{"$gt": timeAgo}}
		}

		opts := options.Find().
			SetSort(bson.D{{Key: "_id", Value: 1}}).
			SetLimit(int64(self.maxRows))

		cursor, findErr := db.GetCollection("topics").Find(ctx, filter, opts)
		if findErr != nil {
			logger.Errorln("IndexingTopic error:", findErr)
			break
		}
		err = cursor.All(ctx, &topicList)
		cursor.Close(ctx)
		if err != nil {
			logger.Errorln("IndexingTopic cursor error:", err)
			break
		}

		if len(topicList) == 0 {
			break
		}

		tids := util.Models2Intslice(topicList, "Tid")

		topicExList := make([]*model.TopicUpEx, 0)
		exCursor, exErr := db.GetCollection("topics_ex").Find(ctx, bson.M{"_id": bson.M{"$in": tids}})
		topicExMap := make(map[int]*model.TopicUpEx)
		if exErr == nil {
			exCursor.All(ctx, &topicExList)
			exCursor.Close(ctx)
			for _, ex := range topicExList {
				topicExMap[ex.Tid] = ex
			}
		}

		for _, topic := range topicList {
			logger.Infoln("deal topic_id:", topic.Tid)

			if id < topic.Tid {
				id = topic.Tid
			}

			if topic.Tags == "" {
				topic.Tags = model.AutoTag(topic.Title, topic.Content, 4)
				if topic.Tags != "" {
					db.GetCollection("topics").UpdateOne(ctx, bson.M{"_id": topic.Tid}, bson.M{"$set": bson.M{"tags": topic.Tags}})
				}
			}

			if topic.Permission == model.PermissionPay {
				topic.Content = "付费用户可见！"
			}

			topicEx := topicExMap[topic.Tid]

			document := model.NewDocument(topic, topicEx)
			addCommand := model.NewDefaultArgsAddCommand(document)

			solrClient.PushAdd(addCommand)
		}

		solrClient.Post()

		if !isAll {
			break
		}
	}
}

func (self SearcherLogic) IndexingResource(isAll bool) {
	solrClient := NewSolrClient()

	ctx := context.Background()

	var (
		resourceList []*model.Resource
		err          error
	)

	id := 0
	for {
		resourceList = make([]*model.Resource, 0)

		var filter bson.M
		if isAll {
			filter = bson.M{"_id": bson.M{"$gt": id}}
		} else {
			timeAgo := time.Now().Add(-2 * time.Minute)
			filter = bson.M{"mtime": bson.M{"$gt": timeAgo}}
		}

		opts := options.Find().
			SetSort(bson.D{{Key: "_id", Value: 1}}).
			SetLimit(int64(self.maxRows))

		cursor, findErr := db.GetCollection("resource").Find(ctx, filter, opts)
		if findErr != nil {
			logger.Errorln("IndexingResource error:", findErr)
			break
		}
		err = cursor.All(ctx, &resourceList)
		cursor.Close(ctx)
		if err != nil {
			logger.Errorln("IndexingResource cursor error:", err)
			break
		}

		if len(resourceList) == 0 {
			break
		}

		ids := util.Models2Intslice(resourceList, "Id")

		resourceExList := make([]*model.ResourceEx, 0)
		exCursor, exErr := db.GetCollection("resource_ex").Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
		resourceExMap := make(map[int]*model.ResourceEx)
		if exErr == nil {
			exCursor.All(ctx, &resourceExList)
			exCursor.Close(ctx)
			for _, ex := range resourceExList {
				resourceExMap[ex.Id] = ex
			}
		}

		for _, resource := range resourceList {
			logger.Infoln("deal resource_id:", resource.Id)

			if id < resource.Id {
				id = resource.Id
			}

			if resource.Tags == "" {
				resource.Tags = model.AutoTag(resource.Title+resource.CatName, resource.Content, 4)
				if resource.Tags != "" {
					db.GetCollection("resource").UpdateOne(ctx, bson.M{"_id": resource.Id}, bson.M{"$set": bson.M{"tags": resource.Tags}})
				}
			}

			resourceEx := resourceExMap[resource.Id]

			document := model.NewDocument(resource, resourceEx)
			addCommand := model.NewDefaultArgsAddCommand(document)

			solrClient.PushAdd(addCommand)
		}

		solrClient.Post()

		if !isAll {
			break
		}
	}
}

// IndexingOpenProject 索引开源项目
func (self SearcherLogic) IndexingOpenProject(isAll bool) {
	solrClient := NewSolrClient()

	ctx := context.Background()

	var (
		projectList []*model.OpenProject
		err         error
	)

	id := 0
	for {
		projectList = make([]*model.OpenProject, 0)

		var filter bson.M
		if isAll {
			filter = bson.M{"_id": bson.M{"$gt": id}}
		} else {
			timeAgo := time.Now().Add(-2 * time.Minute)
			filter = bson.M{"mtime": bson.M{"$gt": timeAgo}}
		}

		opts := options.Find().
			SetSort(bson.D{{Key: "_id", Value: 1}}).
			SetLimit(int64(self.maxRows))

		cursor, findErr := db.GetCollection("open_project").Find(ctx, filter, opts)
		if findErr != nil {
			logger.Errorln("IndexingArticle error:", findErr)
			break
		}
		err = cursor.All(ctx, &projectList)
		cursor.Close(ctx)
		if err != nil {
			logger.Errorln("IndexingOpenProject cursor error:", err)
			break
		}

		if len(projectList) == 0 {
			break
		}

		for _, project := range projectList {
			logger.Infoln("deal project_id:", project.Id)

			if id < project.Id {
				id = project.Id
			}

			if project.Tags == "" {
				project.Tags = model.AutoTag(project.Name+project.Category, project.Desc, 4)
				if project.Tags != "" {
					db.GetCollection("open_project").UpdateOne(ctx, bson.M{"_id": project.Id}, bson.M{"$set": bson.M{"tags": project.Tags}})
				}
			}

			document := model.NewDocument(project, nil)
			if project.Status != model.ProjectStatusOffline {
				solrClient.PushAdd(model.NewDefaultArgsAddCommand(document))
			} else {
				solrClient.PushDel(model.NewDelCommand(document))
			}
		}

		solrClient.Post()

		if !isAll {
			break
		}
	}

}

const searchContentLen = 350

// DoSearch 搜索
func (this *SearcherLogic) DoSearch(q, field string, start, rows int) (*model.ResponseBody, error) {
	selectUrl := this.engineUrl + "/select?"

	var values = url.Values{
		"wt":             []string{"json"},
		"hl":             []string{"true"},
		"hl.fl":          []string{"title,content"},
		"hl.simple.pre":  []string{"<em>"},
		"hl.simple.post": []string{"</em>"},
		"hl.fragsize":    []string{strconv.Itoa(searchContentLen)},
		"start":          []string{strconv.Itoa(start)},
		"rows":           []string{strconv.Itoa(rows)},
	}

	if q == "" {
		values.Add("q", "*:*")
	} else if field == "tag" {
		values.Add("q", "*:*")
		values.Add("fq", "tags:"+q)
		values.Add("sort", "viewnum desc")
		q = ""
		field = ""
	} else {
		ctx := context.Background()
		searchStat := &model.SearchStat{}
		err := db.GetCollection("search_stat").FindOne(ctx, bson.M{"keyword": q}).Decode(searchStat)
		if err == nil && searchStat.Id > 0 {
			db.GetCollection("search_stat").UpdateOne(ctx, bson.M{"keyword": q}, bson.M{"$inc": bson.M{"times": 1}})
		} else {
			searchStat.Keyword = q
			searchStat.Times = 1
			newID, idErr := db.NextID("search_stat")
			if idErr == nil {
				searchStat.Id = newID
				_, insertErr := db.GetCollection("search_stat").InsertOne(ctx, searchStat)
				if insertErr != nil {
					db.GetCollection("search_stat").UpdateOne(ctx, bson.M{"keyword": q}, bson.M{"$inc": bson.M{"times": 1}})
				}
			}
		}
	}

	if field != "" {
		values.Add("df", field)
		if q != "" {
			values.Add("q", q)
		}
	} else {
		if q != "" {
			values.Add("q", "title:"+q+"^2"+" OR content:"+q+"^0.2")
		}
	}
	logger.Infoln(selectUrl + values.Encode())
	resp, err := http.Get(selectUrl + values.Encode())
	if err != nil {
		logger.Errorln("search error:", err)
		return &model.ResponseBody{}, err
	}

	defer resp.Body.Close()

	var searchResponse model.SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResponse)
	if err != nil {
		logger.Errorln("parse response error:", err)
		return &model.ResponseBody{}, err
	}

	if len(searchResponse.Highlight) > 0 {
		for _, doc := range searchResponse.RespBody.Docs {
			highlighting, ok := searchResponse.Highlight[doc.Id]
			if ok {
				if len(highlighting.Title) > 0 {
					doc.HlTitle = highlighting.Title[0]
				}

				if len(highlighting.Content) > 0 {
					doc.HlContent = highlighting.Content[0]
				}
			}

			if doc.HlTitle == "" {
				doc.HlTitle = doc.Title
			}

			if doc.HlContent == "" && doc.Content != "" {
				utf8string := util.NewString(doc.Content)
				maxLen := utf8string.RuneCount() - 1
				if maxLen > searchContentLen {
					maxLen = searchContentLen
				}
				doc.HlContent = util.NewString(doc.Content).Slice(0, maxLen)
			}

			doc.HlContent += "..."
		}

	}

	if searchResponse.RespBody == nil {
		searchResponse.RespBody = &model.ResponseBody{}
	}

	return searchResponse.RespBody, nil
}

// SearchByField 搜索
func (this *SearcherLogic) SearchByField(field, value string, start, rows int, sorts ...string) (*model.ResponseBody, error) {
	selectUrl := this.engineUrl + "/select?"

	sort := "sort_time desc,cmtnum desc,viewnum desc"
	if len(sorts) > 0 {
		sort = sorts[0]
	}
	var values = url.Values{
		"wt":    []string{"json"},
		"start": []string{strconv.Itoa(start)},
		"rows":  []string{strconv.Itoa(rows)},
		"sort":  []string{sort},
		"fl":    []string{"objid,objtype,title,author,uid,pub_time,tags,viewnum,cmtnum,likenum,lastreplyuid,lastreplytime,updated_at,top,nid"},
	}

	values.Add("q", value)
	values.Add("df", field)

	logger.Infoln(selectUrl + values.Encode())

	resp, err := http.Get(selectUrl + values.Encode())
	if err != nil {
		logger.Errorln("search error:", err)
		return &model.ResponseBody{}, err
	}

	defer resp.Body.Close()

	var searchResponse model.SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResponse)
	if err != nil {
		logger.Errorln("parse response error:", err)
		return &model.ResponseBody{}, err
	}

	if searchResponse.RespBody == nil {
		searchResponse.RespBody = &model.ResponseBody{}
	}

	return searchResponse.RespBody, nil
}

func (this *SearcherLogic) FindAtomFeeds(rows int) (*model.ResponseBody, error) {
	selectUrl := this.engineUrl + "/select?"

	var values = url.Values{
		"q":     []string{"*:*"},
		"sort":  []string{"sort_time desc"},
		"wt":    []string{"json"},
		"start": []string{"0"},
		"rows":  []string{strconv.Itoa(rows)},
	}

	resp, err := http.Get(selectUrl + values.Encode())
	if err != nil {
		logger.Errorln("search error:", err)
		return &model.ResponseBody{}, err
	}

	defer resp.Body.Close()

	var searchResponse model.SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResponse)
	if err != nil {
		logger.Errorln("parse response error:", err)
		return &model.ResponseBody{}, err
	}

	if searchResponse.RespBody == nil {
		searchResponse.RespBody = &model.ResponseBody{}
	}

	return searchResponse.RespBody, nil
}

func (this *SearcherLogic) FillNodeAndUser(ctx context.Context, respBody *model.ResponseBody) (map[int]*model.User, map[int]*model.TopicNode) {
	if respBody.NumFound == 0 {
		return nil, nil
	}

	uidSet := set.New(set.NonThreadSafe)
	nidSet := set.New(set.NonThreadSafe)

	for _, doc := range respBody.Docs {
		if doc.Uid > 0 {
			uidSet.Add(doc.Uid)
		}
		if doc.Lastreplyuid > 0 {
			uidSet.Add(doc.Lastreplyuid)
		}
		if doc.Nid > 0 {
			nidSet.Add(doc.Nid)
		}
	}

	users := DefaultUser.FindUserInfos(nil, set.IntSlice(uidSet))
	nodes := GetNodesByNids(set.IntSlice(nidSet))

	return users, nodes
}

type SolrClient struct {
	addCommands []*model.AddCommand
	delCommands []*model.DelCommand
}

func NewSolrClient() *SolrClient {
	return &SolrClient{
		addCommands: make([]*model.AddCommand, 0, 100),
		delCommands: make([]*model.DelCommand, 0, 100),
	}
}

func (this *SolrClient) PushAdd(addCommand *model.AddCommand) {
	this.addCommands = append(this.addCommands, addCommand)
}

func (this *SolrClient) PushDel(delCommand *model.DelCommand) {
	this.delCommands = append(this.delCommands, delCommand)
}

func (this *SolrClient) Post() error {
	stringBuilder := goutils.NewBuffer().Append("{")

	needComma := false
	for _, addCommand := range this.addCommands {
		commandJson, err := json.Marshal(addCommand)
		if err != nil {
			continue
		}

		if stringBuilder.Len() == 1 {
			needComma = false
		} else {
			needComma = true
		}

		if needComma {
			stringBuilder.Append(",")
		}

		stringBuilder.Append(`"add":`).Append(commandJson)
	}

	for _, delCommand := range this.delCommands {
		commandJson, err := json.Marshal(delCommand)
		if err != nil {
			continue
		}

		if stringBuilder.Len() == 1 {
			needComma = false
		} else {
			needComma = true
		}

		if needComma {
			stringBuilder.Append(",")
		}

		stringBuilder.Append(`"delete":`).Append(commandJson)
	}

	if stringBuilder.Len() == 1 {
		logger.Errorln("post docs:no right addcommand")
		return errors.New("no right addcommand")
	}

	stringBuilder.Append("}")

	logger.Infoln("start post data to solr...")

	resp, err := http.Post(config.ConfigFile.MustValue("search", "engine_url")+"/update?wt=json&commit=true", "application/json", stringBuilder)
	if err != nil {
		logger.Errorln("post error:", err)
		return err
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		logger.Errorln("parse response error:", err)
		return err
	}

	logger.Infoln("post data result:", result)

	return nil
}
