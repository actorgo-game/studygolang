// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"errors"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/PuerkitoBio/goquery"
	"github.com/lunny/html2md"
	"github.com/polaris1119/config"
	"github.com/polaris1119/logger"
	"github.com/polaris1119/set"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProjectLogic struct{}

var DefaultProject = ProjectLogic{}

func (self ProjectLogic) Publish(ctx context.Context, user *model.Me, form url.Values) (err error) {
	objLog := GetLogger(ctx)

	id := form.Get("id")
	isModify := id != ""

	project := &model.OpenProject{}

	if isModify {
		idInt, _ := strconv.Atoi(id)
		err = db.GetCollection("open_project").FindOne(ctx, bson.M{"_id": idInt}).Decode(project)
		if err != nil {
			objLog.Errorln("Publish Project find error:", err)
			return
		}

		if !CanEdit(user, project) {
			err = NotModifyAuthorityErr
			return
		}

		err = schemaDecoder.Decode(project, form)
		if err != nil {
			objLog.Errorln("Publish Project schema decode error:", err)
			return
		}
	} else {
		err = schemaDecoder.Decode(project, form)
		if err != nil {
			objLog.Errorln("Publish Project schema decode error:", err)
			return
		}

		project.Username = user.Username
	}
	if project.Uri == "" {
		project.Uri = strings.Replace(project.Name, " ", "-", -1)
	}
	project.Uri = strings.ToLower(project.Uri)

	if !isModify && self.UriExists(ctx, form.Get("uri")) {
		err = errors.New("项目已存在")
		return
	}

	github := "github.com"
	pos := strings.Index(project.Src, github)
	if pos != -1 {
		project.Repo = project.Src[pos+len(github)+1:]
	}

	if !isModify {
		newId, idErr := db.NextID("open_project")
		if idErr != nil {
			objLog.Errorln("Publish Project NextID error:", idErr)
			err = idErr
			return
		}
		project.Id = newId
		_, err = db.GetCollection("open_project").InsertOne(ctx, project)
	} else {
		idInt, _ := strconv.Atoi(id)
		_, err = db.GetCollection("open_project").UpdateOne(ctx, bson.M{"_id": idInt}, bson.M{"$set": project})
	}

	if err != nil {
		objLog.Errorln("Publish Project error:", err)
		return
	}

	if isModify {
		go modifyObservable.NotifyObservers(user.Uid, model.TypeProject, project.Id)
	} else {
		go publishObservable.NotifyObservers(user.Uid, model.TypeProject, project.Id)
	}

	return
}

// UriExists 通过 uri 是否存在 project
func (ProjectLogic) UriExists(ctx context.Context, uri string) bool {
	total, err := db.GetCollection("open_project").CountDocuments(ctx, bson.M{"uri": uri})
	if err != nil || total == 0 {
		return false
	}

	return true
}

// Total 开源项目总数
func (ProjectLogic) Total() int64 {
	ctx := context.Background()
	total, err := db.GetCollection("open_project").CountDocuments(ctx, bson.M{})
	if err != nil {
		logger.Errorln("ProjectLogic Total error:", err)
	}
	return total
}

// FindBy 获取开源项目列表（分页）
func (ProjectLogic) FindBy(ctx context.Context, limit int, lastIds ...int) []*model.OpenProject {
	objLog := GetLogger(ctx)

	filter := bson.M{"status": bson.M{"$in": []int{model.ProjectStatusNew, model.ProjectStatusOnline}}}
	if len(lastIds) > 0 && lastIds[0] > 0 {
		filter["_id"] = bson.M{"$lt": lastIds[0]}
	}

	projectList := make([]*model.OpenProject, 0)
	opts := options.Find().SetSort(bson.D{{"_id", -1}}).SetLimit(int64(limit))
	cursor, err := db.GetCollection("open_project").Find(ctx, filter, opts)
	if err != nil {
		objLog.Errorln("ProjectLogic FindBy Error:", err)
		return nil
	}
	if err = cursor.All(ctx, &projectList); err != nil {
		objLog.Errorln("ProjectLogic FindBy cursor Error:", err)
		return nil
	}

	return projectList
}

// FindByIds 获取多个项目详细信息
func (ProjectLogic) FindByIds(ids []int) []*model.OpenProject {
	if len(ids) == 0 {
		return nil
	}

	ctx := context.Background()
	projects := make([]*model.OpenProject, 0)
	cursor, err := db.GetCollection("open_project").Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		logger.Errorln("ProjectLogic FindByIds error:", err)
		return nil
	}
	if err = cursor.All(ctx, &projects); err != nil {
		logger.Errorln("ProjectLogic FindByIds cursor error:", err)
		return nil
	}
	return projects
}

// findByIds 获取多个项目详细信息 包内使用
func (ProjectLogic) findByIds(ids []int) map[int]*model.OpenProject {
	if len(ids) == 0 {
		return nil
	}

	ctx := context.Background()
	projectList := make([]*model.OpenProject, 0)
	cursor, err := db.GetCollection("open_project").Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		logger.Errorln("ProjectLogic FindByIds error:", err)
		return nil
	}
	if err = cursor.All(ctx, &projectList); err != nil {
		logger.Errorln("ProjectLogic FindByIds cursor error:", err)
		return nil
	}

	projects := make(map[int]*model.OpenProject, len(projectList))
	for _, p := range projectList {
		projects[p.Id] = p
	}
	return projects
}

// FindOne 获取单个项目
func (ProjectLogic) FindOne(ctx context.Context, val interface{}) *model.OpenProject {
	objLog := GetLogger(ctx)

	filter := bson.M{
		"status": bson.M{"$in": []int{model.ProjectStatusNew, model.ProjectStatusOnline}},
	}

	switch v := val.(type) {
	case int:
		filter["_id"] = v
	case string:
		if _, err := strconv.Atoi(v); err != nil {
			filter["uri"] = v
		} else {
			filter["_id"] = v
		}
	}

	project := &model.OpenProject{}
	err := db.GetCollection("open_project").FindOne(ctx, filter).Decode(project)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			objLog.Errorln("project service FindProject error:", err)
		}
		return nil
	}

	project.User = DefaultUser.FindOne(ctx, "username", project.Username)

	return project
}

// FindRecent 获得某个用户最近发布的开源项目
func (ProjectLogic) FindRecent(ctx context.Context, username string) []*model.OpenProject {
	projectList := make([]*model.OpenProject, 0)
	opts := options.Find().SetSort(bson.D{{"_id", -1}}).SetLimit(5)
	cursor, err := db.GetCollection("open_project").Find(ctx, bson.M{"username": username}, opts)
	if err != nil {
		logger.Errorln("project logic FindRecent error:", err)
		return nil
	}
	if err = cursor.All(ctx, &projectList); err != nil {
		logger.Errorln("project logic FindRecent cursor error:", err)
		return nil
	}
	return projectList
}

// FindAll 支持多页翻看
func (self ProjectLogic) FindAll(ctx context.Context, paginator *Paginator, orderBy string, querystring string, args ...interface{}) []*model.OpenProject {
	objLog := GetLogger(ctx)

	projects := make([]*model.OpenProject, 0)

	sort := parseSort(orderBy)
	filter := buildFilter(querystring, args...)

	opts := options.Find().
		SetSort(sort).
		SetLimit(int64(paginator.PerPage())).
		SetSkip(int64(paginator.Offset()))
	cursor, err := db.GetCollection("open_project").Find(ctx, filter, opts)
	if err != nil {
		objLog.Errorln("ProjectLogic FindAll error:", err)
		return nil
	}
	if err = cursor.All(ctx, &projects); err != nil {
		objLog.Errorln("ProjectLogic FindAll cursor error:", err)
		return nil
	}

	self.fillUser(projects)

	return projects
}

func (ProjectLogic) Count(ctx context.Context, querystring string, args ...interface{}) int64 {
	objLog := GetLogger(ctx)

	filter := buildFilter(querystring, args...)
	total, err := db.GetCollection("open_project").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("ProjectLogic Count error:", err)
	}

	return total
}

func (ProjectLogic) fillUser(projects []*model.OpenProject) {
	usernameSet := set.New(set.NonThreadSafe)
	uidSet := set.New(set.NonThreadSafe)
	for _, project := range projects {
		usernameSet.Add(project.Username)

		if project.Lastreplyuid != 0 {
			uidSet.Add(project.Lastreplyuid)
		}
	}
	if !usernameSet.IsEmpty() {
		userMap := DefaultUser.FindUserInfos(nil, set.StringSlice(usernameSet))
		for _, project := range projects {
			for _, user := range userMap {
				if project.Username == user.Username {
					project.User = user
					break
				}
			}
		}
	}

	if !uidSet.IsEmpty() {
		replyUserMap := DefaultUser.FindUserInfos(nil, set.IntSlice(uidSet))
		for _, project := range projects {
			if project.Lastreplyuid == 0 {
				continue
			}

			project.LastReplyUser = replyUserMap[project.Lastreplyuid]
		}
	}
}

// getOwner 通过objid获得 project 的所有者
func (ProjectLogic) getOwner(ctx context.Context, id int) int {
	project := &model.OpenProject{}
	err := db.GetCollection("open_project").FindOne(ctx, bson.M{"_id": id}).Decode(project)
	if err != nil {
		logger.Errorln("project logic getOwner Error:", err)
		return 0
	}

	user := DefaultUser.FindOne(ctx, "username", project.Username)
	return user.Uid
}

// ParseProjectList 解析其他网站的开源项目
func (self ProjectLogic) ParseProjectList(pUrl string) error {
	pUrl = strings.TrimSpace(pUrl)
	if !strings.HasPrefix(pUrl, "http") {
		pUrl = "http://" + pUrl
	}

	var (
		doc *goquery.Document
		err error
	)

	if doc, err = goquery.NewDocument(pUrl); err != nil {
		logger.Errorln("goquery opensource project newdocument error:", err)
		return err
	}

	// 最后面的先入库处理
	projectsSelection := doc.Find("#projectList .list-container").Children()

	for i := projectsSelection.Length() - 1; i >= 0; i-- {
		contentSelection := goquery.NewDocumentFromNode(projectsSelection.Get(i)).Selection
		projectUrl, ok := contentSelection.Find(".content .header a").First().Attr("href")

		if !ok || projectUrl == "" {
			logger.Errorln("project url is empty")
			continue
		}
		go func(projectUrl string) {
			err := self.ParseOneProject(projectUrl)

			if err != nil {
				logger.Errorln(err)
			}
		}(projectUrl)
	}

	return err
}

const OsChinaDomain = "https://www.oschina.net"

// ProjectLogoPrefix 开源项目 logo 前缀
const ProjectLogoPrefix = "plogo"

var PresetUsernames = config.ConfigFile.MustValueArray("crawl", "preset_users", ",")

// ParseOneProject 处理单个 project
func (ProjectLogic) ParseOneProject(projectUrl string) error {
	if !strings.HasPrefix(projectUrl, "http") {
		projectUrl = OsChinaDomain + projectUrl
	}

	var (
		doc *goquery.Document
		err error
	)

	// 加上 ?fromerr=xfwefs，否则页面有 js 重定向
	if doc, err = goquery.NewDocument(projectUrl + "?fromerr=xfwefs"); err != nil {
		return errors.New("goquery fetch " + projectUrl + " error:" + err.Error())
	}

	// 标题
	category := strings.TrimSpace(doc.Find(".detail-header h1 .project-title").Text())
	name := strings.TrimSpace(doc.Find(".detail-header h1 .project-name").Text())
	if category == "" && name == "" {
		return errors.New("projectUrl:" + projectUrl + " category and name are empty")
	}

	tmpIndex := strings.LastIndex(category, name)
	if tmpIndex != -1 {
		category = category[:tmpIndex]
	}

	// uri
	uri := projectUrl[strings.LastIndex(projectUrl, "/")+1:]

	ctx := context.Background()
	project := &model.OpenProject{}

	err = db.GetCollection("open_project").FindOne(ctx, bson.M{"uri": uri}).Decode(project)
	// 已经存在
	if project.Id != 0 {
		logger.Infoln("url", projectUrl, "has exists!")
		return nil
	}

	logoSelection := doc.Find(".detail-header .logo-wrap img")
	if logoSelection.AttrOr("alt", "") != "" {
		project.Logo = logoSelection.AttrOr("src", "")

		if !strings.HasPrefix(project.Logo, "http") {
			project.Logo = ""
		} else {
			project.Logo, err = DefaultUploader.TransferUrl(nil, project.Logo, ProjectLogoPrefix)
			if err != nil {
				logger.Errorln("project logo upload error:", err)
			}
		}
	}

	// 获取项目相关链接
	doc.Find(".related-links a").Each(func(i int, aSelection *goquery.Selection) {
		uri := aSelection.AttrOr("href", "")
		switch aSelection.Text() {
		case "软件首页":
			project.Home = uri
		case "软件文档":
			project.Doc = uri
		case "官方下载":
			project.Download = uri
		}
	})

	doc.Find(".info-list .box .info-item").Each(func(i int, liSelection *goquery.Selection) {
		aSelection := liSelection.Find("span")
		txt := strings.TrimSpace(aSelection.Text())
		if i == 0 {
			project.Licence = txt
			if txt == "未知" {
				project.Licence = "其他"
			}
		} else if i == 1 {
			txt = liSelection.Find("span a:first-child").Text()
			project.Lang = txt
		} else if i == 2 {
			project.Os = txt
		}
	})

	project.Name = name
	project.Category = strings.TrimSpace(category)
	project.Uri = uri
	project.Src = project.Download

	if strings.HasPrefix(project.Src, "https://github.com/") {
		project.Repo = project.Src[len("https://github.com/"):]
	} else if strings.HasPrefix(project.Src, "https://gitee.com/") {
		project.Repo = project.Src[len("https://gitee.com/"):]
	} else {
		return nil
	}

	pos := strings.Index(project.Repo, "/")
	if pos > -1 {
		project.Author = project.Repo[:pos]
	} else {
		project.Author = "网友"
	}

	if project.Doc == "" {
		project.Doc = "https://godoc.org/" + project.Src[8:]
	}

	desc := ""
	doc.Find(".project-body").Children().Each(func(i int, domSelection *goquery.Selection) {
		if domSelection.HasClass("ad-wrap") {
			return
		}
		doc.FindSelection(domSelection).WrapHtml(`<div id="tmp` + strconv.Itoa(i) + `"></div>`)
		domHtml, _ := doc.Find("#tmp" + strconv.Itoa(i)).Html()
		if domSelection.Is("pre") {
			desc += domHtml + "\n\n"
		} else {
			desc += html2md.Convert(domHtml) + "\n\n"
		}
	})

	project.Desc = strings.TrimSpace(desc)
	project.Username = PresetUsernames[rand.Intn(len(PresetUsernames))]
	project.Status = model.ProjectStatusOnline
	project.Ctime = model.OftenTime(time.Now())

	newId, idErr := db.NextID("open_project")
	if idErr != nil {
		return errors.New("NextID open project error:" + idErr.Error())
	}
	project.Id = newId

	_, err = db.GetCollection("open_project").InsertOne(ctx, project)
	if err != nil {
		return errors.New("insert into open project error:" + err.Error())
	}

	return nil
}

// 项目评论
type ProjectComment struct{}

// 更新该项目的评论信息
// cid：评论id；objid：被评论对象id；uid：评论者；cmttime：评论时间
func (self ProjectComment) UpdateComment(cid, objid, uid int, cmttime time.Time) {
	ctx := context.Background()
	_, err := db.GetCollection("open_project").UpdateOne(ctx, bson.M{"_id": objid}, bson.M{
		"$inc": bson.M{"cmtnum": 1},
		"$set": bson.M{
			"lastreplyuid":  uid,
			"lastreplytime": cmttime,
		},
	})
	if err != nil {
		logger.Errorln("更新项目评论数失败：", err)
		return
	}
}

func (self ProjectComment) String() string {
	return "project"
}

// 实现 CommentObjecter 接口
func (self ProjectComment) SetObjinfo(ids []int, commentMap map[int][]*model.Comment) {
	projects := DefaultProject.FindByIds(ids)
	if len(projects) == 0 {
		return
	}

	for _, project := range projects {
		objinfo := make(map[string]interface{})
		objinfo["title"] = project.Category + project.Name
		objinfo["uri"] = model.PathUrlMap[model.TypeProject]
		objinfo["type_name"] = model.TypeNameMap[model.TypeProject]

		for _, comment := range commentMap[project.Id] {
			comment.Objinfo = objinfo
		}
	}
}

// 项目喜欢
type ProjectLike struct{}

// 更新该项目的喜欢数
// objid：被喜欢对象id；num: 喜欢数(负数表示取消喜欢)
func (self ProjectLike) UpdateLike(objid, num int) {
	ctx := context.Background()
	_, err := db.GetCollection("open_project").UpdateOne(ctx, bson.M{"_id": objid}, bson.M{"$inc": bson.M{"likenum": num}})
	if err != nil {
		logger.Errorln("更新项目喜欢数失败：", err)
	}
}

func (self ProjectLike) String() string {
	return "project"
}

// parseSort converts an SQL-style ORDER BY clause to a MongoDB sort document
func parseSort(orderBy string) bson.D {
	if orderBy == "" {
		return bson.D{{"_id", -1}}
	}
	parts := strings.Fields(orderBy)
	field := parts[0]
	if field == "id" {
		field = "_id"
	}
	dir := 1
	if len(parts) > 1 && strings.EqualFold(parts[1], "DESC") {
		dir = -1
	}
	return bson.D{{field, dir}}
}

// buildFilter converts a simple SQL WHERE clause to a MongoDB filter
func buildFilter(querystring string, args ...interface{}) bson.M {
	if querystring == "" || len(args) == 0 {
		return bson.M{}
	}

	filter := bson.M{}
	argIdx := 0
	conditions := strings.Split(querystring, " AND ")
	for _, cond := range conditions {
		cond = strings.TrimSpace(cond)
		if strings.Contains(cond, " IN(") {
			field := cond[:strings.Index(cond, " ")]
			if field == "id" {
				field = "_id"
			}
			numPlaceholders := strings.Count(cond, "?")
			vals := make([]interface{}, 0, numPlaceholders)
			for i := 0; i < numPlaceholders && argIdx < len(args); i++ {
				vals = append(vals, args[argIdx])
				argIdx++
			}
			filter[field] = bson.M{"$in": vals}
		} else if idx := strings.Index(cond, "=?"); idx != -1 {
			field := strings.TrimSpace(cond[:idx])
			if field == "id" {
				field = "_id"
			}
			if argIdx < len(args) {
				filter[field] = args[argIdx]
				argIdx++
			}
		} else if idx := strings.Index(cond, "<?"); idx != -1 {
			field := strings.TrimSpace(cond[:idx])
			if field == "id" {
				field = "_id"
			}
			if argIdx < len(args) {
				filter[field] = bson.M{"$lt": args[argIdx]}
				argIdx++
			}
		} else if idx := strings.Index(cond, ">?"); idx != -1 {
			field := strings.TrimSpace(cond[:idx])
			if field == "id" {
				field = "_id"
			}
			if argIdx < len(args) {
				filter[field] = bson.M{"$gt": args[argIdx]}
				argIdx++
			}
		}
	}

	return filter
}
