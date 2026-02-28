// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/global"
	"github.com/studygolang/studygolang/internal/model"
	"github.com/studygolang/studygolang/util"

	"github.com/polaris1119/goutils"
	"github.com/polaris1119/set"
	"github.com/polaris1119/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SubjectLogic struct{}

var DefaultSubject = SubjectLogic{}

func (self SubjectLogic) FindBy(ctx context.Context, paginator *Paginator) []*model.Subject {
	objLog := GetLogger(ctx)

	subjects := make([]*model.Subject, 0)
	opts := options.Find().
		SetSort(bson.D{{"article_num", -1}}).
		SetLimit(int64(paginator.PerPage())).
		SetSkip(int64(paginator.Offset()))
	cursor, err := db.GetCollection("subject").Find(ctx, bson.M{}, opts)
	if err != nil {
		objLog.Errorln("SubjectLogic FindBy error:", err)
		return subjects
	}
	if err = cursor.All(ctx, &subjects); err != nil {
		objLog.Errorln("SubjectLogic FindBy cursor error:", err)
		return subjects
	}

	if len(subjects) > 0 {
		uidSet := set.New(set.NonThreadSafe)
		for _, subject := range subjects {
			uidSet.Add(subject.Uid)
		}
		usersMap := DefaultUser.FindUserInfos(ctx, set.IntSlice(uidSet))
		for _, subject := range subjects {
			subject.User = usersMap[subject.Uid]
		}
	}

	return subjects
}

func (self SubjectLogic) FindOne(ctx context.Context, sid int) *model.Subject {
	objLog := GetLogger(ctx)

	subject := &model.Subject{}
	err := db.GetCollection("subject").FindOne(ctx, bson.M{"_id": sid}).Decode(subject)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			objLog.Errorln("SubjectLogic FindOne get error:", err)
		}
	}

	if subject.Uid > 0 {
		subject.User = DefaultUser.findUser(ctx, subject.Uid)
	}

	return subject
}

func (self SubjectLogic) findByIds(ids []int) map[int]*model.Subject {
	if len(ids) == 0 {
		return nil
	}

	ctx := context.Background()
	subjectList := make([]*model.Subject, 0)
	cursor, err := db.GetCollection("subject").Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil
	}
	if err = cursor.All(ctx, &subjectList); err != nil {
		return nil
	}

	subjects := make(map[int]*model.Subject, len(subjectList))
	for _, s := range subjectList {
		subjects[s.Id] = s
	}
	return subjects
}

func (self SubjectLogic) FindArticles(ctx context.Context, sid int, paginator *Paginator, orderBy string) []*model.Article {
	objLog := GetLogger(ctx)

	// Step 1: Find subject_articles
	saFilter := bson.M{"sid": sid, "state": model.ContributeStateOnline}
	saOpts := options.Find().SetSort(bson.D{{"created_at", -1}})

	subjectArticles := make([]*model.SubjectArticle, 0)
	cursor, err := db.GetCollection("subject_article").Find(ctx, saFilter, saOpts)
	if err != nil {
		objLog.Errorln("SubjectLogic FindArticles Find subject_article error:", err)
		return nil
	}
	if err = cursor.All(ctx, &subjectArticles); err != nil {
		objLog.Errorln("SubjectLogic FindArticles cursor error:", err)
		return nil
	}

	if len(subjectArticles) == 0 {
		return nil
	}

	// Step 2: Collect article IDs
	articleIds := make([]int, 0, len(subjectArticles))
	for _, sa := range subjectArticles {
		articleIds = append(articleIds, sa.ArticleId)
	}

	// Step 3: Fetch articles
	artOpts := options.Find()
	if orderBy == "commented_at" {
		artOpts.SetSort(bson.D{{"lastreplytime", -1}})
	}
	artOpts.SetLimit(int64(paginator.PerPage())).SetSkip(int64(paginator.Offset()))

	articleList := make([]*model.Article, 0)
	cursor, err = db.GetCollection("articles").Find(ctx, bson.M{"_id": bson.M{"$in": articleIds}}, artOpts)
	if err != nil {
		objLog.Errorln("SubjectLogic FindArticles find articles error:", err)
		return nil
	}
	if err = cursor.All(ctx, &articleList); err != nil {
		objLog.Errorln("SubjectLogic FindArticles articles cursor error:", err)
		return nil
	}

	articles := make([]*model.Article, 0, len(articleList))
	for _, article := range articleList {
		if article.Status != model.ArticleStatusOffline {
			articles = append(articles, article)
		}
	}

	DefaultArticle.fillUser(articles)
	return articles
}

// FindArticleTotal 专栏收录的文章数
func (self SubjectLogic) FindArticleTotal(ctx context.Context, sid int) int64 {
	objLog := GetLogger(ctx)

	total, err := db.GetCollection("subject_article").CountDocuments(ctx, bson.M{"sid": sid})
	if err != nil {
		objLog.Errorln("SubjectLogic FindArticleTotal error:", err)
	}

	return total
}

// FindFollowers 专栏关注的用户
func (self SubjectLogic) FindFollowers(ctx context.Context, sid int) []*model.SubjectFollower {
	objLog := GetLogger(ctx)

	followers := make([]*model.SubjectFollower, 0)
	opts := options.Find().SetSort(bson.D{{"_id", -1}}).SetLimit(8)
	cursor, err := db.GetCollection("subject_follower").Find(ctx, bson.M{"sid": sid}, opts)
	if err != nil {
		objLog.Errorln("SubjectLogic FindFollowers error:", err)
		return followers
	}
	if err = cursor.All(ctx, &followers); err != nil {
		objLog.Errorln("SubjectLogic FindFollowers cursor error:", err)
		return followers
	}

	if len(followers) == 0 {
		return followers
	}

	uids := slices.StructsIntSlice(followers, "Uid")
	usersMap := DefaultUser.FindUserInfos(ctx, uids)
	for _, follower := range followers {
		follower.User = usersMap[follower.Uid]
		follower.TimeAgo = util.TimeAgo(follower.CreatedAt)
	}

	return followers
}

func (self SubjectLogic) findFollowersBySid(sid int) []*model.SubjectFollower {
	ctx := context.Background()
	followers := make([]*model.SubjectFollower, 0)
	cursor, err := db.GetCollection("subject_follower").Find(ctx, bson.M{"sid": sid})
	if err == nil {
		cursor.All(ctx, &followers)
	}
	return followers
}

// FindFollowerTotal 专栏关注的用户数
func (self SubjectLogic) FindFollowerTotal(ctx context.Context, sid int) int64 {
	objLog := GetLogger(ctx)

	total, err := db.GetCollection("subject_follower").CountDocuments(ctx, bson.M{"sid": sid})
	if err != nil {
		objLog.Errorln("SubjectLogic FindFollowerTotal error:", err)
	}

	return total
}

// Follow 关注或取消关注
func (self SubjectLogic) Follow(ctx context.Context, sid int, me *model.Me) (err error) {
	objLog := GetLogger(ctx)

	follower := &model.SubjectFollower{}
	err = db.GetCollection("subject_follower").FindOne(ctx, bson.M{"sid": sid, "uid": me.Uid}).Decode(follower)
	if err != nil && err != mongo.ErrNoDocuments {
		objLog.Errorln("SubjectLogic Follow Get error:", err)
	}

	if follower.Id > 0 {
		_, err = db.GetCollection("subject_follower").DeleteOne(ctx, bson.M{"sid": sid, "uid": me.Uid})
		if err != nil {
			objLog.Errorln("SubjectLogic Follow Delete error:", err)
		}
		return
	}

	follower.Sid = sid
	follower.Uid = me.Uid

	id, err := db.NextID("subject_follower")
	if err != nil {
		objLog.Errorln("SubjectLogic Follow NextID error:", err)
		return
	}
	follower.Id = id

	_, err = db.GetCollection("subject_follower").InsertOne(ctx, follower)
	if err != nil {
		objLog.Errorln("SubjectLogic Follow insert error:", err)
	}
	return
}

func (self SubjectLogic) HadFollow(ctx context.Context, sid int, me *model.Me) bool {
	objLog := GetLogger(ctx)

	num, err := db.GetCollection("subject_follower").CountDocuments(ctx, bson.M{"sid": sid, "uid": me.Uid})
	if err != nil {
		objLog.Errorln("SubjectLogic Follow insert error:", err)
	}

	return num > 0
}

// Contribute 投稿
func (self SubjectLogic) Contribute(ctx context.Context, me *model.Me, sid, articleId int) error {
	objLog := GetLogger(ctx)

	subject := self.FindOne(ctx, sid)
	if subject.Id == 0 {
		return errors.New("该专栏不存在")
	}

	count, _ := db.GetCollection("subject_article").CountDocuments(ctx, bson.M{"article_id": articleId})
	if count >= 5 {
		return errors.New("该文超过 5 次投稿")
	}

	subjectArticle := &model.SubjectArticle{
		Sid:       sid,
		ArticleId: articleId,
		State:     model.ContributeStateNew,
	}

	if subject.Uid == me.Uid {
		subjectArticle.State = model.ContributeStateOnline
	} else {
		if !subject.Contribute {
			return errors.New("不允许投稿")
		}
		if !subject.Audit {
			subjectArticle.State = model.ContributeStateOnline
		}
	}

	session, err := db.GetClient().StartSession()
	if err != nil {
		objLog.Errorln("SubjectLogic Contribute start session error:", err)
		return errors.New("投稿失败:" + err.Error())
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		id, err := db.NextID("subject_article")
		if err != nil {
			return nil, err
		}
		subjectArticle.Id = id

		_, err = db.GetCollection("subject_article").InsertOne(sessCtx, subjectArticle)
		if err != nil {
			return nil, err
		}

		_, err = db.GetCollection("subject").UpdateOne(sessCtx, bson.M{"_id": sid}, bson.M{"$inc": bson.M{"article_num": 1}})
		return nil, err
	})

	if err != nil {
		objLog.Errorln("SubjectLogic Contribute error:", err)
		return errors.New("投稿失败:" + err.Error())
	}

	go self.sendMsgForFollower(ctx, subject, sid, articleId)

	return nil
}

// sendMsgForFollower 专栏投稿发送消息给关注者
func (self SubjectLogic) sendMsgForFollower(ctx context.Context, subject *model.Subject, sid, articleId int) {
	followers := self.findFollowersBySid(sid)
	for _, f := range followers {
		DefaultMessage.SendSystemMsgTo(ctx, f.Uid, model.MsgtypeSubjectContribute, map[string]interface{}{
			"uid":   subject.Uid,
			"objid": articleId,
			"sid":   sid,
		})
	}
}

// RemoveContribute 删除投稿
func (self SubjectLogic) RemoveContribute(ctx context.Context, sid, articleId int) error {
	objLog := GetLogger(ctx)

	session, err := db.GetClient().StartSession()
	if err != nil {
		objLog.Errorln("SubjectLogic RemoveContribute start session error:", err)
		return errors.New("删除投稿失败:" + err.Error())
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		_, err := db.GetCollection("subject_article").DeleteOne(sessCtx, bson.M{"sid": sid, "article_id": articleId})
		if err != nil {
			return nil, err
		}

		_, err = db.GetCollection("subject").UpdateOne(sessCtx, bson.M{"_id": sid}, bson.M{"$inc": bson.M{"article_num": -1}})
		return nil, err
	})

	if err != nil {
		objLog.Errorln("SubjectLogic RemoveContribute error:", err)
		return errors.New("删除投稿失败:" + err.Error())
	}

	return nil
}

func (self SubjectLogic) ExistByName(name string) bool {
	ctx := context.Background()
	count, err := db.GetCollection("subject").CountDocuments(ctx, bson.M{"name": name})
	return err == nil && count > 0
}

// Publish 发布专栏。
func (self SubjectLogic) Publish(ctx context.Context, me *model.Me, form url.Values) (sid int, err error) {
	objLog := GetLogger(ctx)

	sid = goutils.MustInt(form.Get("sid"))
	if sid != 0 {
		subject := &model.Subject{}
		err = db.GetCollection("subject").FindOne(ctx, bson.M{"_id": sid}).Decode(subject)
		if err != nil {
			objLog.Errorln("Publish Subject find error:", err)
			return
		}

		_, err = self.Modify(ctx, me, form)
		if err != nil {
			objLog.Errorln("Publish Subject modify error:", err)
			return
		}

	} else {
		subject := &model.Subject{}
		err = schemaDecoder.Decode(subject, form)
		if err != nil {
			objLog.Errorln("SubjectLogic Publish decode error:", err)
			return
		}
		subject.Uid = me.Uid

		newId, idErr := db.NextID("subject")
		if idErr != nil {
			objLog.Errorln("SubjectLogic Publish NextID error:", idErr)
			err = idErr
			return
		}
		subject.Id = newId

		_, err = db.GetCollection("subject").InsertOne(ctx, subject)
		if err != nil {
			objLog.Errorln("SubjectLogic Publish insert error:", err)
			return
		}
		sid = subject.Id
	}
	return
}

// Modify 修改专栏
func (SubjectLogic) Modify(ctx context.Context, user *model.Me, form url.Values) (errMsg string, err error) {
	objLog := GetLogger(ctx)

	change := map[string]interface{}{}

	fields := []string{"name", "description", "cover", "contribute", "audit"}
	for _, field := range fields {
		change[field] = form.Get(field)
	}

	sid := form.Get("sid")
	sidInt, _ := strconv.Atoi(sid)
	_, err = db.GetCollection("subject").UpdateOne(ctx, bson.M{"_id": sidInt}, bson.M{"$set": change})
	if err != nil {
		objLog.Errorf("更新专栏 【%s】 信息失败：%s\n", sid, err)
		errMsg = "对不起，服务器内部错误，请稍后再试！"
		return
	}

	return
}

func (self SubjectLogic) FindArticleSubjects(ctx context.Context, articleId int) []*model.Subject {
	objLog := GetLogger(ctx)

	subjectArticles := make([]*model.SubjectArticle, 0)
	cursor, err := db.GetCollection("subject_article").Find(ctx, bson.M{"article_id": articleId})
	if err != nil {
		objLog.Errorln("SubjectLogic FindArticleSubjects find error:", err)
		return nil
	}
	if err = cursor.All(ctx, &subjectArticles); err != nil {
		objLog.Errorln("SubjectLogic FindArticleSubjects cursor error:", err)
		return nil
	}

	subjectLen := len(subjectArticles)
	if subjectLen == 0 {
		return nil
	}

	sids := make([]int, subjectLen)
	for i, subjectArticle := range subjectArticles {
		sids[i] = subjectArticle.Sid
	}

	subjects := make([]*model.Subject, 0)
	cursor, err = db.GetCollection("subject").Find(ctx, bson.M{"_id": bson.M{"$in": sids}})
	if err != nil {
		objLog.Errorln("SubjectLogic FindArticleSubjects find subject error:", err)
		return nil
	}
	if err = cursor.All(ctx, &subjects); err != nil {
		objLog.Errorln("SubjectLogic FindArticleSubjects subject cursor error:", err)
		return nil
	}

	return subjects
}

// FindMine 获取我管理的专栏列表
func (self SubjectLogic) FindMine(ctx context.Context, me *model.Me, articleId int, kw string) []map[string]interface{} {
	objLog := GetLogger(ctx)

	// 先是我创建的专栏
	subjects := make([]*model.Subject, 0)
	filter := bson.M{"uid": me.Uid}
	if kw != "" {
		filter["name"] = bson.M{"$regex": kw}
	}
	cursor, err := db.GetCollection("subject").Find(ctx, filter)
	if err != nil {
		objLog.Errorln("SubjectLogic FindMine find subject error:", err)
		return nil
	}
	if err = cursor.All(ctx, &subjects); err != nil {
		objLog.Errorln("SubjectLogic FindMine subject cursor error:", err)
		return nil
	}

	// 获取我管理的专栏
	adminSubjects := make([]*model.Subject, 0)
	saAdmins := make([]*model.SubjectAdmin, 0)
	saCursor, saErr := db.GetCollection("subject_admin").Find(ctx, bson.M{"uid": me.Uid})
	if saErr == nil {
		saCursor.All(ctx, &saAdmins)
	}
	if len(saAdmins) > 0 {
		sids := make([]int, len(saAdmins))
		for i, sa := range saAdmins {
			sids[i] = sa.Sid
		}
		adminFilter := bson.M{"_id": bson.M{"$in": sids}}
		if kw != "" {
			adminFilter["name"] = bson.M{"$regex": kw}
		}
		adminCursor, adminErr := db.GetCollection("subject").Find(ctx, adminFilter)
		if adminErr == nil {
			adminCursor.All(ctx, &adminSubjects)
		}
	}

	subjectArticles := make([]*model.SubjectArticle, 0)
	saCursor2, saErr2 := db.GetCollection("subject_article").Find(ctx, bson.M{"article_id": articleId})
	if saErr2 == nil {
		saCursor2.All(ctx, &subjectArticles)
	}
	subjectArticleMap := make(map[int]struct{})
	for _, sa := range subjectArticles {
		subjectArticleMap[sa.Sid] = struct{}{}
	}

	uidSet := set.New(set.NonThreadSafe)
	for _, subject := range subjects {
		uidSet.Add(subject.Uid)
	}
	for _, subject := range adminSubjects {
		uidSet.Add(subject.Uid)
	}
	usersMap := DefaultUser.FindUserInfos(ctx, set.IntSlice(uidSet))

	subjectMapSlice := make([]map[string]interface{}, 0, len(subjects)+len(adminSubjects))

	for _, subject := range subjects {
		self.genSubjectMapSlice(subject, &subjectMapSlice, subjectArticleMap, usersMap)
	}

	for _, subject := range adminSubjects {
		self.genSubjectMapSlice(subject, &subjectMapSlice, subjectArticleMap, usersMap)
	}

	return subjectMapSlice
}

func (self SubjectLogic) genSubjectMapSlice(subject *model.Subject, subjectMapSlice *[]map[string]interface{}, subjectArticleMap map[int]struct{}, usersMap map[int]*model.User) {
	hadAdd := 0
	if _, ok := subjectArticleMap[subject.Id]; ok {
		hadAdd = 1
	}

	cover := subject.Cover
	if cover == "" {
		user := usersMap[subject.Uid]
		cover = util.Gravatar(user.Avatar, user.Email, 48, true)
	} else if !strings.HasPrefix(cover, "http") {
		cdnDomain := global.App.CanonicalCDN(true)
		cover = cdnDomain + subject.Cover
	}

	*subjectMapSlice = append(*subjectMapSlice, map[string]interface{}{
		"id":       subject.Id,
		"name":     subject.Name,
		"cover":    cover,
		"username": usersMap[subject.Uid].Username,
		"had_add":  hadAdd,
	})
}
