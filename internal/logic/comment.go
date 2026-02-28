// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"fmt"
	"html/template"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/fatih/structs"
	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"github.com/polaris1119/set"
	"github.com/polaris1119/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

type CommentLogic struct{}

var DefaultComment = CommentLogic{}

// FindObjComments 获得某个对象的所有评论
// owner: 被评论对象属主
func (self CommentLogic) FindObjComments(ctx context.Context, objid, objtype int, owner, lastCommentUid int) (comments []map[string]interface{}, ownerUser, lastReplyUser *model.User) {
	objLog := GetLogger(ctx)

	coll := db.GetCollection("comments")
	filter := bson.M{"objid": objid, "objtype": objtype}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		objLog.Errorln("CommentLogic FindObjComments Error:", err)
		return
	}
	defer cursor.Close(ctx)

	commentList := make([]*model.Comment, 0)
	if err = cursor.All(ctx, &commentList); err != nil {
		objLog.Errorln("CommentLogic FindObjComments decode Error:", err)
		return
	}

	uids := slices.StructsIntSlice(commentList, "Uid")

	// 避免某些情况下最后回复人没在回复列表中
	uids = append(uids, owner, lastCommentUid)

	userMap := DefaultUser.FindUserInfos(ctx, uids)
	ownerUser = userMap[owner]
	if lastCommentUid != 0 {
		lastReplyUser = userMap[lastCommentUid]
	}
	comments = make([]map[string]interface{}, 0, len(commentList))
	for _, comment := range commentList {
		tmpMap := structs.Map(comment)
		tmpMap["content"] = template.HTML(self.decodeCmtContent(ctx, comment))
		tmpMap["user"] = userMap[comment.Uid]
		comments = append(comments, tmpMap)
	}
	return
}

const CommentPerNum = 50

// FindObjectComments 获得某个对象的所有评论（新版）
func (self CommentLogic) FindObjectComments(ctx context.Context, objid, objtype, p int) (commentList []*model.Comment, replyComments []*model.Comment, pageNum int, err error) {
	objLog := GetLogger(ctx)

	coll := db.GetCollection("comments")
	filter := bson.M{"objid": objid, "objtype": objtype}

	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("comment logic FindObjectComments count Error:", err)
		return
	}

	pageNum = int(math.Ceil(float64(total) / CommentPerNum))
	if p == 0 {
		p = pageNum
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}).
		SetSkip(int64((p - 1) * CommentPerNum)).
		SetLimit(CommentPerNum)

	cursor, err := coll.Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("comment logic FindObjectComments Error:", err)
		return
	}
	defer cursor.Close(ctx)

	commentList = make([]*model.Comment, 0)
	if err = cursor.All(ctx, &commentList); err != nil {
		objLog.Errorln("comment logic FindObjectComments decode Error:", err)
		return
	}

	floors := make([]interface{}, 0, len(commentList))
	for _, comment := range commentList {
		self.decodeCmtContentForShow(ctx, comment, true)

		if comment.ReplyFloor > 0 {
			floors = append(floors, comment.ReplyFloor)
		}
	}

	if len(floors) > 0 {
		replyFilter := bson.M{"objid": objid, "objtype": objtype, "floor": bson.M{"$in": floors}}
		replyCursor, replyErr := coll.Find(ctx, replyFilter)
		if replyErr != nil {
			err = replyErr
			return
		}
		defer replyCursor.Close(ctx)

		replyComments = make([]*model.Comment, 0)
		err = replyCursor.All(ctx, &replyComments)
	}

	return
}

// FindComment 获得评论和额外两个评论
func (self CommentLogic) FindComment(ctx context.Context, cid, objid, objtype int) (*model.Comment, []*model.Comment) {
	objLog := GetLogger(ctx)

	coll := db.GetCollection("comments")

	comment := &model.Comment{}
	err := coll.FindOne(ctx, bson.M{"_id": cid}).Decode(comment)
	if err != nil {
		objLog.Errorln("CommentLogic FindComment error:", err)
		return comment, nil
	}
	self.decodeCmtContentForShow(ctx, comment, false)

	findOpts := options.Find().SetLimit(2)
	cursor, err := coll.Find(ctx, bson.M{"objid": objid, "objtype": objtype, "_id": bson.M{"$ne": cid}}, findOpts)
	if err != nil {
		objLog.Errorln("CommentLogic FindComment Find more error:", err)
		return comment, nil
	}
	defer cursor.Close(ctx)

	comments := make([]*model.Comment, 0)
	if err = cursor.All(ctx, &comments); err != nil {
		objLog.Errorln("CommentLogic FindComment decode more error:", err)
		return comment, nil
	}
	for _, cmt := range comments {
		self.decodeCmtContentForShow(ctx, cmt, false)
	}

	return comment, comments
}

// Total 评论总数(objtypes[0] 取某一类型的评论总数)
func (CommentLogic) Total(objtypes ...int) int64 {
	filter := bson.M{}
	if len(objtypes) > 0 {
		filter["objtype"] = objtypes[0]
	}

	total, err := db.GetCollection("comments").CountDocuments(context.Background(), filter)
	if err != nil {
		logger.Errorln("CommentLogic Total error:", err)
	}
	return total
}

// FindRecent 获得最近的评论
// 如果 uid!=0，表示获取某人的评论；
// 如果 objtype!=-1，表示获取某类型的评论；
func (self CommentLogic) FindRecent(ctx context.Context, uid, objtype, limit int) []*model.Comment {
	filter := bson.M{}
	if uid != 0 {
		filter["uid"] = uid
	}
	if objtype != -1 {
		filter["objtype"] = objtype
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := db.GetCollection("comments").Find(ctx, filter, findOpts)
	if err != nil {
		logger.Errorln("CommentLogic FindRecent error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	comments := make([]*model.Comment, 0)
	if err = cursor.All(ctx, &comments); err != nil {
		logger.Errorln("CommentLogic FindRecent decode error:", err)
		return nil
	}

	cmtMap := make(map[int][]*model.Comment, len(model.PathUrlMap))
	for _, comment := range comments {
		self.decodeCmtContent(ctx, comment)

		if _, ok := cmtMap[comment.Objtype]; !ok {
			cmtMap[comment.Objtype] = make([]*model.Comment, 0, 10)
		}

		cmtMap[comment.Objtype] = append(cmtMap[comment.Objtype], comment)
	}

	cmtObjs := []CommentObjecter{
		model.TypeTopic:     TopicComment{},
		model.TypeArticle:   ArticleComment{},
		model.TypeResource:  ResourceComment{},
		model.TypeWiki:      nil,
		model.TypeProject:   ProjectComment{},
		model.TypeBook:      BookComment{},
		model.TypeInterview: InterviewComment{},
	}
	for cmtType, cmts := range cmtMap {
		self.fillObjinfos(cmts, cmtObjs[cmtType])
	}

	return self.filterDelObjectCmt(comments)
}

// Publish 发表评论（或回复）。
// objid 注册的评论对象
// uid 评论人
func (self CommentLogic) Publish(ctx context.Context, uid, objid int, form url.Values) (*model.Comment, error) {
	objLog := GetLogger(ctx)

	objtype := goutils.MustInt(form.Get("objtype"))
	comment := &model.Comment{
		Objid:   objid,
		Objtype: objtype,
		Uid:     uid,
		Content: form.Get("content"),
	}

	coll := db.GetCollection("comments")

	// 暂时只是从数据库中取出最后的评论楼层
	tmpCmt := &model.Comment{}
	findOpts := options.FindOne().SetSort(bson.D{{Key: "floor", Value: -1}})
	err := coll.FindOne(ctx, bson.M{"objid": objid, "objtype": objtype}, findOpts).Decode(tmpCmt)
	if err != nil && err != mongo.ErrNoDocuments {
		objLog.Errorln("post comment find last floor error:", err)
		return nil, err
	}

	comment.Floor = tmpCmt.Floor + 1

	if tmpCmt.Uid == comment.Uid && tmpCmt.Content == comment.Content {
		objLog.Infof("had post comment: %+v", *comment)
		return tmpCmt, nil
	}

	// 入评论库
	cid, idErr := db.NextID("comments")
	if idErr != nil {
		objLog.Errorln("post comment NextID error:", idErr)
		return nil, idErr
	}
	comment.Cid = cid

	_, err = coll.InsertOne(ctx, comment)
	if err != nil {
		objLog.Errorln("post comment service error:", err)
		return nil, err
	}
	self.decodeCmtContentForShow(ctx, comment, true)

	if commenter, ok := commenters[objtype]; ok {
		now := time.Now()

		objLog.Debugf("评论[objid:%d] [objtype:%d] [uid:%d] 成功，通知被评论者更新", objid, objtype, uid)
		go commenter.UpdateComment(comment.Cid, objid, uid, now)

		DefaultFeed.updateComment(objid, objtype, uid, now)
	}

	go commentObservable.NotifyObservers(uid, objtype, comment.Cid)

	go self.sendSystemMsg(ctx, uid, objid, objtype, comment.Cid, form)

	return comment, nil
}

func (CommentLogic) sendSystemMsg(ctx context.Context, uid, objid, objtype, cid int, form url.Values) {
	ext := map[string]interface{}{
		"objid":   objid,
		"objtype": objtype,
		"cid":     cid,
		"uid":     uid,
	}

	to := 0
	switch objtype {
	case model.TypeTopic:
		to = DefaultTopic.getOwner(objid)
	case model.TypeArticle:
		to = DefaultArticle.getOwner(objid)
	case model.TypeResource:
		to = DefaultResource.getOwner(objid)
	case model.TypeWiki:
		to = DefaultWiki.getOwner(objid)
	case model.TypeProject:
		to = DefaultProject.getOwner(ctx, objid)
	}

	DefaultMessage.SendSystemMsgTo(ctx, to, objtype, ext)

	// @某人 发系统消息
	DefaultMessage.SendSysMsgAtUids(ctx, form.Get("uid"), ext, to)
	DefaultMessage.SendSysMsgAtUsernames(ctx, form.Get("usernames"), ext, to)
}

// Modify 修改评论信息
func (CommentLogic) Modify(ctx context.Context, cid int, content string) (errMsg string, err error) {
	objLog := GetLogger(ctx)

	_, err = db.GetCollection("comments").UpdateOne(ctx, bson.M{"_id": cid}, bson.M{"$set": bson.M{"content": content}})
	if err != nil {
		objLog.Errorf("更新评论内容 【%d】 失败：%s", cid, err)
		errMsg = "对不起，服务器内部错误，请稍后再试！"
		return
	}

	return
}

// fillObjinfos 填充评论对应的主体信息
func (CommentLogic) fillObjinfos(comments []*model.Comment, cmtObj CommentObjecter) {
	if len(comments) == 0 {
		return
	}
	count := len(comments)
	commentMap := make(map[int][]*model.Comment, count)
	idSet := set.New(set.NonThreadSafe)
	for _, comment := range comments {
		if _, ok := commentMap[comment.Objid]; !ok {
			commentMap[comment.Objid] = make([]*model.Comment, 0, count)
		}
		commentMap[comment.Objid] = append(commentMap[comment.Objid], comment)
		idSet.Add(comment.Objid)
	}
	cmtObj.SetObjinfo(set.IntSlice(idSet), commentMap)
}

// findByIds 提供给其他service调用（包内）
func (CommentLogic) findByIds(cids []int) map[int]*model.Comment {
	if len(cids) == 0 {
		return nil
	}

	coll := db.GetCollection("comments")
	cursor, err := coll.Find(context.Background(), bson.M{"_id": bson.M{"$in": cids}})
	if err != nil {
		return nil
	}
	defer cursor.Close(context.Background())

	commentList := make([]*model.Comment, 0)
	if err = cursor.All(context.Background(), &commentList); err != nil {
		return nil
	}

	comments := make(map[int]*model.Comment, len(commentList))
	for _, c := range commentList {
		comments[c.Cid] = c
	}
	return comments
}

func (CommentLogic) FindById(cid int) (*model.Comment, error) {
	comment := &model.Comment{}
	err := db.GetCollection("comments").FindOne(context.Background(), bson.M{"_id": cid}).Decode(comment)
	if err != nil {
		logger.Errorln("CommentLogic findById error:", err)
	}

	return comment, err
}

func (CommentLogic) decodeCmtContent(ctx context.Context, comment *model.Comment) string {
	// 安全过滤
	content := template.HTMLEscapeString(comment.Content)
	// @别人
	content = parseAtUser(ctx, content)

	// 回复某一楼层
	reg := regexp.MustCompile(`#(\d+)楼`)
	url := fmt.Sprintf("%s%d#comment", model.PathUrlMap[comment.Objtype], comment.Objid)
	content = reg.ReplaceAllString(content, `<a href="`+url+`$1" title="$1">#$1<span>楼</span></a>`)

	comment.Content = content

	return content
}

// decodeCmtContentForShow 采用引用的方式显示对其他楼层的回复
func (CommentLogic) decodeCmtContentForShow(ctx context.Context, comment *model.Comment, isEscape bool) {
	// 安全过滤
	content := template.HTMLEscapeString(comment.Content)

	// 回复某一楼层
	reg := regexp.MustCompile(`#(\d+)楼 @([a-zA-Z0-9_-]+)`)
	matches := reg.FindStringSubmatch(content)
	if len(matches) > 2 {
		comment.ReplyFloor = goutils.MustInt(matches[1])
		content = strings.TrimSpace(content[len(matches[0]):])
	}

	// @别人
	content = parseAtUser(ctx, content)

	comment.Content = content
}

// CommentObjecter 填充 Comment 对象的 Objinfo 成员接口
type CommentObjecter interface {
	SetObjinfo(ids []int, commentMap map[int][]*model.Comment)
}

// FindAll 支持多页翻看
func (self CommentLogic) FindAll(ctx context.Context, paginator *Paginator, orderBy string, querystring string, args ...interface{}) []*model.Comment {
	objLog := GetLogger(ctx)

	coll := db.GetCollection("comments")
	filter := bson.M{}
	if querystring != "" {
		filter = buildFilter(querystring, args...)
	}

	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("CommentLogical FindAll count error:", err)
		return nil
	}
	paginator.SetTotal(total)

	findOpts := options.Find().
		SetSort(buildCommentSort(orderBy)).
		SetSkip(int64(paginator.Offset())).
		SetLimit(int64(paginator.PerPage()))

	cursor, err := coll.Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("CommentLogical FindAll error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	comments := make([]*model.Comment, 0)
	if err = cursor.All(ctx, &comments); err != nil {
		objLog.Errorln("CommentLogical FindAll decode error:", err)
		return nil
	}

	cmtMap := make(map[int][]*model.Comment, len(model.PathUrlMap))
	for _, comment := range comments {
		self.decodeCmtContent(ctx, comment)
		if _, ok := cmtMap[comment.Objtype]; !ok {
			cmtMap[comment.Objtype] = make([]*model.Comment, 0, 10)
		}

		cmtMap[comment.Objtype] = append(cmtMap[comment.Objtype], comment)
	}

	cmtObjs := []CommentObjecter{
		model.TypeTopic:     TopicComment{},
		model.TypeArticle:   ArticleComment{},
		model.TypeResource:  ResourceComment{},
		model.TypeWiki:      nil,
		model.TypeProject:   ProjectComment{},
		model.TypeBook:      BookComment{},
		model.TypeInterview: InterviewComment{},
	}
	for cmtType, cmts := range cmtMap {
		self.fillObjinfos(cmts, cmtObjs[cmtType])
	}

	return self.filterDelObjectCmt(comments)
}

// EnrichWithUsers 为评论列表填充用户信息
func (CommentLogic) EnrichWithUsers(ctx context.Context, comments []*model.Comment) []map[string]interface{} {
	if len(comments) == 0 {
		return nil
	}
	uids := make([]int, 0, len(comments))
	for _, c := range comments {
		uids = append(uids, c.Uid)
	}
	userMap := DefaultUser.FindUserInfos(ctx, uids)
	result := make([]map[string]interface{}, len(comments))
	for i, c := range comments {
		m := structs.Map(c)
		m["user"] = userMap[c.Uid]
		result[i] = m
	}
	return result
}

// Count 获取用户全部评论数
func (CommentLogic) Count(ctx context.Context, querystring string, args ...interface{}) int64 {
	objLog := GetLogger(ctx)

	filter := bson.M{}
	if querystring != "" {
		filter = buildFilter(querystring, args...)
	}

	total, err := db.GetCollection("comments").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("CommentLogic Count error:", err)
	}

	return total
}

func (CommentLogic) filterDelObjectCmt(comments []*model.Comment) []*model.Comment {
	resultCmts := make([]*model.Comment, 0, len(comments))
	for _, comment := range comments {
		if comment.Objinfo != nil && len(comment.Objinfo) > 0 {
			resultCmts = append(resultCmts, comment)
		}
	}
	return resultCmts
}

// CommentLike 回复赞（喜欢）
type CommentLike struct{}

// UpdateLike 更新该回复的赞
// objid：被喜欢对象id；num: 喜欢数(负数表示取消喜欢)
func (self CommentLike) UpdateLike(objid, num int) {
	_, err := db.GetCollection("comments").UpdateOne(
		context.Background(),
		bson.M{"_id": objid},
		bson.M{"$inc": bson.M{"likenum": num}},
	)
	if err != nil {
		logger.Errorln("更新回复喜欢数失败：", err)
	}
}

func (self CommentLike) String() string {
	return "comment"
}

func buildCommentSort(orderBy string) bson.D {
	if orderBy == "" {
		return bson.D{{Key: "_id", Value: -1}}
	}
	var sort bson.D
	parts := splitOrderBy(orderBy)
	for _, p := range parts {
		field := p[0]
		if field == "cid" {
			field = "_id"
		}
		dir := 1
		if len(p) > 1 && (p[1] == "DESC" || p[1] == "desc") {
			dir = -1
		}
		sort = append(sort, bson.E{Key: field, Value: dir})
	}
	return sort
}
