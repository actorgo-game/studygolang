// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"sync"
	"time"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"
	"github.com/studygolang/studygolang/util"

	"github.com/fatih/structs"
	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"github.com/polaris1119/set"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

type TopicLogic struct{}

var DefaultTopic = TopicLogic{}

// Publish 发布主题。入topics和topics_ex库
func (self TopicLogic) Publish(ctx context.Context, me *model.Me, form url.Values) (tid int, err error) {
	objLog := GetLogger(ctx)

	tid = goutils.MustInt(form.Get("tid"))
	if tid != 0 {
		topic := &model.Topic{}
		err = db.GetCollection("topics").FindOne(ctx, bson.M{"_id": tid}).Decode(topic)
		if err != nil {
			objLog.Errorln("Publish Topic find error:", err)
			return
		}

		if !CanEdit(me, topic) {
			err = NotModifyAuthorityErr
			return
		}

		_, err = self.Modify(ctx, me, form)
		if err != nil {
			objLog.Errorln("Publish Topic modify error:", err)
			return
		}

		nid := goutils.MustInt(form.Get("nid"))

		go func() {
			if topic.Uid != me.Uid && topic.Nid != nid {
				node := DefaultNode.FindOne(nid)
				award := -500
				if node.ShowIndex {
					award = -30
				}
				desc := fmt.Sprintf(`主题节点被管理员调整为 <a href="/go/%s">%s</a>`, node.Ename, node.Name)
				user := DefaultUser.FindOne(ctx, "uid", topic.Uid)
				DefaultUserRich.IncrUserRich(user, model.MissionTypeModify, award, desc)
			}

			if nid != topic.Nid {
				DefaultFeed.modifyTopicNode(tid, nid)
			}
		}()
	} else {
		usernames := form.Get("usernames")
		form.Del("usernames")

		topic := &model.Topic{}
		err = schemaDecoder.Decode(topic, form)
		if err != nil {
			objLog.Errorln("TopicLogic Publish decode error:", err)
			return
		}
		topic.Uid = me.Uid
		topic.Lastreplytime = model.NewOftenTime()

		newID, idErr := db.NextID("topics")
		if idErr != nil {
			err = idErr
			objLog.Errorln("TopicLogic Publish NextID error:", err)
			return
		}
		topic.Tid = newID

		session, sessErr := db.GetClient().StartSession()
		if sessErr != nil {
			err = sessErr
			objLog.Errorln("TopicLogic Publish StartSession error:", err)
			return
		}
		defer session.EndSession(ctx)

		_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
			_, insertErr := db.GetCollection("topics").InsertOne(sc, topic)
			if insertErr != nil {
				return nil, insertErr
			}

			topicEx := &model.TopicEx{
				Tid: topic.Tid,
			}
			_, insertErr = db.GetCollection("topics_ex").InsertOne(sc, topicEx)
			if insertErr != nil {
				return nil, insertErr
			}
			return nil, nil
		})
		if err != nil {
			objLog.Errorln("TopicLogic Publish transaction error:", err)
			return
		}

		go func() {
			bgCtx := context.Background()
			topicNum, countErr := db.GetCollection("topics").CountDocuments(bgCtx, bson.M{
				"uid":   me.Uid,
				"ctime": bson.M{"$gt": time.Now().Format("2006-01-02 00:00:00")},
			})
			if countErr != nil {
				logger.Errorln("find today topic num error:", countErr)
				return
			}

			if topicNum > 3 {
				node := DefaultNode.FindOne(topic.Nid)
				if node.ShowIndex {
					return
				}

				award := -1000

				desc := fmt.Sprintf(`一天发布推广过多或 Spam 扣除铜币 %d 个`, -award)
				user := DefaultUser.FindOne(ctx, "uid", me.Uid)
				DefaultUserRich.IncrUserRich(user, model.MissionTypeSpam, award, desc)

				DefaultRank.GenDAURank(me.Uid, -1000)
			}
		}()

		topicEx := &model.TopicEx{Tid: topic.Tid}
		DefaultFeed.publish(topic, topicEx, me)

		ext := map[string]interface{}{
			"objid":   topic.Tid,
			"objtype": model.TypeTopic,
			"uid":     me.Uid,
			"msgtype": model.MsgtypePublishAtMe,
		}
		go DefaultMessage.SendSysMsgAtUsernames(ctx, usernames, ext, 0)

		go publishObservable.NotifyObservers(me.Uid, model.TypeTopic, topic.Tid)

		tid = topic.Tid
	}

	return
}

// Modify 修改主题
func (TopicLogic) Modify(ctx context.Context, user *model.Me, form url.Values) (errMsg string, err error) {
	objLog := GetLogger(ctx)

	change := bson.M{
		"editor_uid": user.Uid,
	}

	fields := []string{"title", "content", "nid", "permission"}
	for _, field := range fields {
		change[field] = form.Get(field)
	}

	tid := goutils.MustInt(form.Get("tid"))
	_, err = db.GetCollection("topics").UpdateOne(ctx, bson.M{"_id": tid}, bson.M{"$set": change})
	if err != nil {
		objLog.Errorf("更新主题 【%d】 信息失败：%s\n", tid, err)
		errMsg = "对不起，服务器内部错误，请稍后再试！"
		return
	}

	go modifyObservable.NotifyObservers(user.Uid, model.TypeTopic, tid)

	return
}

// Append 主题附言
func (self TopicLogic) Append(ctx context.Context, uid, tid int, content string) error {
	objLog := GetLogger(ctx)

	num, err := db.GetCollection("topic_append").CountDocuments(ctx, bson.M{"tid": tid})
	if err != nil {
		objLog.Errorln("TopicLogic Append error:", err)
		return err
	}

	if num >= model.AppendMaxNum {
		return errors.New("不允许再发附言！")
	}

	newID, err := db.NextID("topic_append")
	if err != nil {
		objLog.Errorln("TopicLogic Append NextID error:", err)
		return err
	}

	topicAppend := &model.TopicAppend{
		Id:      newID,
		Tid:     tid,
		Content: content,
	}
	_, err = db.GetCollection("topic_append").InsertOne(ctx, topicAppend)

	if err != nil {
		objLog.Errorln("TopicLogic Append insert error:", err)
		return err
	}

	go appendObservable.NotifyObservers(uid, model.TypeTopic, tid)

	return nil
}

// SetTop 置顶
func (self TopicLogic) SetTop(ctx context.Context, me *model.Me, tid int) error {
	objLog := GetLogger(ctx)

	if !me.IsAdmin {
		topic := self.findByTid(tid)
		if topic.Tid == 0 || topic.Uid != me.Uid {
			return NotFoundErr
		}
	}

	_, err := db.GetCollection("topics").UpdateOne(ctx, bson.M{"_id": tid}, bson.M{"$set": bson.M{
		"top":      1,
		"top_time": time.Now().Unix(),
	}})
	if err != nil {
		objLog.Errorln("TopicLogic SetTop error:", err)
		return err
	}

	err = DefaultFeed.setTop(ctx, tid, model.TypeTopic, 1)
	if err != nil {
		objLog.Errorln("TopicLogic SetTop feed error:", err)
		return err
	}

	go topObservable.NotifyObservers(me.Uid, model.TypeTopic, tid)

	return nil
}

// UnsetTop 取消置顶
func (self TopicLogic) UnsetTop(ctx context.Context, tid int) error {
	objLog := GetLogger(ctx)

	_, err := db.GetCollection("topics").UpdateOne(ctx, bson.M{"_id": tid}, bson.M{"$set": bson.M{
		"top": 0,
	}})
	if err != nil {
		objLog.Errorln("TopicLogic UnsetTop error:", err)
		return err
	}

	err = DefaultFeed.setTop(ctx, tid, model.TypeTopic, 0)
	if err != nil {
		objLog.Errorln("TopicLogic UnsetTop feed error:", err)
		return err
	}

	return nil
}

// AutoUnsetTop 自动取消置顶
func (self TopicLogic) AutoUnsetTop() error {
	ctx := context.Background()
	cursor, err := db.GetCollection("topics").Find(ctx, bson.M{"top": 1})
	if err != nil {
		logger.Errorln("TopicLogic AutoUnsetTop error:", err)
		return err
	}
	defer cursor.Close(ctx)

	topics := make([]*model.Topic, 0)
	if err = cursor.All(ctx, &topics); err != nil {
		logger.Errorln("TopicLogic AutoUnsetTop decode error:", err)
		return err
	}

	for _, topic := range topics {
		if topic.TopTime == 0 || topic.TopTime+86400 > time.Now().Unix() {
			continue
		}

		self.UnsetTop(ctx, topic.Tid)
	}

	return nil
}

// FindAll 支持多页翻看
func (self TopicLogic) FindAll(ctx context.Context, paginator *Paginator, orderBy string, querystring string, args ...interface{}) []map[string]interface{} {
	objLog := GetLogger(ctx)

	filter := bson.M{}
	if querystring != "" {
		filter = buildFilter(querystring, args...)
	}
	filter = self.addFlagFilter(filter)

	total, err := db.GetCollection("topics").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("TopicLogic FindAll count error:", err)
		return nil
	}
	paginator.SetTotal(total)

	findOpts := options.Find().
		SetSort(buildTopicSort(orderBy)).
		SetSkip(int64(paginator.Offset())).
		SetLimit(int64(paginator.PerPage()))

	cursor, err := db.GetCollection("topics").Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("TopicLogic FindAll error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	topics := make([]*model.Topic, 0)
	if err = cursor.All(ctx, &topics); err != nil {
		objLog.Errorln("TopicLogic FindAll decode error:", err)
		return nil
	}

	tids := make([]int, len(topics))
	for i, t := range topics {
		tids[i] = t.Tid
	}

	topicExMap := make(map[int]*model.TopicEx)
	if len(tids) > 0 {
		exCursor, exErr := db.GetCollection("topics_ex").Find(ctx, bson.M{"tid": bson.M{"$in": tids}})
		if exErr == nil {
			defer exCursor.Close(ctx)
			exList := make([]*model.TopicEx, 0)
			if exCursor.All(ctx, &exList) == nil {
				for _, ex := range exList {
					topicExMap[ex.Tid] = ex
				}
			}
		}
	}

	topicInfos := make([]*model.TopicInfo, 0, len(topics))
	for _, t := range topics {
		info := &model.TopicInfo{Topic: *t}
		if ex, ok := topicExMap[t.Tid]; ok {
			info.TopicEx = *ex
		}
		topicInfos = append(topicInfos, info)
	}

	return self.fillDataForTopicInfo(topicInfos)
}

func (TopicLogic) FindLastList(beginTime string, limit int) ([]*model.Topic, error) {
	ctx := context.Background()
	filter := bson.M{
		"ctime": bson.M{"$gt": beginTime},
		"flag":  bson.M{"$in": []uint8{model.FlagNoAudit, model.FlagNormal}},
	}
	findOpts := options.Find().
		SetSort(bson.M{"_id": -1}).
		SetLimit(int64(limit))

	cursor, err := db.GetCollection("topics").Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	topics := make([]*model.Topic, 0)
	err = cursor.All(ctx, &topics)
	return topics, err
}

// FindRecent 获得最近的主题(uids[0]，则获取某个用户最近的主题)
func (self TopicLogic) FindRecent(limit int, uids ...int) []*model.Topic {
	ctx := context.Background()
	filter := self.addFlagFilter(bson.M{})
	if len(uids) > 0 {
		filter["uid"] = uids[0]
	}

	findOpts := options.Find().
		SetSort(bson.M{"ctime": -1}).
		SetLimit(int64(limit))

	cursor, err := db.GetCollection("topics").Find(ctx, filter, findOpts)
	if err != nil {
		logger.Errorln("TopicLogic FindRecent error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	topics := make([]*model.Topic, 0)
	if err = cursor.All(ctx, &topics); err != nil {
		logger.Errorln("TopicLogic FindRecent decode error:", err)
		return nil
	}

	for _, topic := range topics {
		topic.Node = GetNodeName(topic.Nid)
	}
	return topics
}

// FindByNid 获得某个节点下的主题列表（侧边栏推荐）
func (TopicLogic) FindByNid(ctx context.Context, nid, curTid string) []*model.Topic {
	objLog := GetLogger(ctx)

	filter := bson.M{
		"nid":  goutils.MustInt(nid),
		"_id":  bson.M{"$ne": goutils.MustInt(curTid)},
		"flag": bson.M{"$lt": model.FlagAuditDelete},
	}
	findOpts := options.Find().SetLimit(10)

	cursor, err := db.GetCollection("topics").Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("TopicLogic FindByNid Error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	topics := make([]*model.Topic, 0)
	if err = cursor.All(ctx, &topics); err != nil {
		objLog.Errorln("TopicLogic FindByNid decode error:", err)
	}

	return topics
}

// FindByTids 获取多个主题详细信息
func (TopicLogic) FindByTids(tids []int) []*model.Topic {
	if len(tids) == 0 {
		return nil
	}

	ctx := context.Background()
	cursor, err := db.GetCollection("topics").Find(ctx, bson.M{"_id": bson.M{"$in": tids}})
	if err != nil {
		logger.Errorln("TopicLogic FindByTids error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	topics := make([]*model.Topic, 0)
	if err = cursor.All(ctx, &topics); err != nil {
		logger.Errorln("TopicLogic FindByTids decode error:", err)
		return nil
	}
	return topics
}

func (self TopicLogic) FindFullinfoByTids(tids []int) []map[string]interface{} {
	if len(tids) == 0 {
		return nil
	}

	ctx := context.Background()

	topicCursor, err := db.GetCollection("topics").Find(ctx, bson.M{"_id": bson.M{"$in": tids}})
	if err != nil {
		logger.Errorln("TopicLogic FindFullinfoByTids topics error:", err)
		return nil
	}
	defer topicCursor.Close(ctx)

	topicList := make([]*model.Topic, 0)
	if err = topicCursor.All(ctx, &topicList); err != nil {
		logger.Errorln("TopicLogic FindFullinfoByTids topics decode error:", err)
		return nil
	}
	topicMap := make(map[int]*model.Topic, len(topicList))
	for _, t := range topicList {
		topicMap[t.Tid] = t
	}

	exCursor, err := db.GetCollection("topics_ex").Find(ctx, bson.M{"tid": bson.M{"$in": tids}})
	if err != nil {
		logger.Errorln("TopicLogic FindFullinfoByTids topics_ex error:", err)
		return nil
	}
	defer exCursor.Close(ctx)

	exList := make([]*model.TopicEx, 0)
	if err = exCursor.All(ctx, &exList); err != nil {
		logger.Errorln("TopicLogic FindFullinfoByTids topics_ex decode error:", err)
		return nil
	}
	exMap := make(map[int]*model.TopicEx, len(exList))
	for _, ex := range exList {
		exMap[ex.Tid] = ex
	}

	topicInfos := make([]*model.TopicInfo, 0, len(tids))
	for _, tid := range tids {
		t, ok := topicMap[tid]
		if !ok {
			continue
		}
		if t.Flag > model.FlagNormal {
			continue
		}
		info := &model.TopicInfo{Topic: *t}
		if ex, exOk := exMap[tid]; exOk {
			info.TopicEx = *ex
		}
		topicInfos = append(topicInfos, info)
	}

	return self.fillDataForTopicInfo(topicInfos)
}

// FindByTid 获得主题详细信息（包括详细回复）
func (self TopicLogic) FindByTid(ctx context.Context, tid int) (topicMap map[string]interface{}, replies []map[string]interface{}, err error) {
	objLog := GetLogger(ctx)

	topic := &model.Topic{}
	err = db.GetCollection("topics").FindOne(ctx, bson.M{"_id": tid}).Decode(topic)
	if err != nil {
		objLog.Errorln("TopicLogic FindByTid get error:", err)
		return
	}

	if topic.Tid == 0 {
		err = errors.New("The topic of tid is not exists")
		objLog.Errorln("TopicLogic FindByTid get error:", err)
		return
	}

	if topic.Flag > model.FlagNormal {
		err = errors.New("The topic of tid is not exists or delete")
		return
	}

	topicEx := &model.TopicEx{}
	_ = db.GetCollection("topics_ex").FindOne(ctx, bson.M{"tid": tid}).Decode(topicEx)

	topicMap = make(map[string]interface{})
	structs.FillMap(topic, topicMap)
	structs.FillMap(topicEx, topicMap)

	topicMap["content"] = self.decodeTopicContent(ctx, topic)

	topicMap["node"] = GetNode(topic.Nid)

	replies, owerUser, lastReplyUser := DefaultComment.FindObjComments(ctx, topic.Tid, model.TypeTopic, topic.Uid, topic.Lastreplyuid)
	topicMap["user"] = owerUser
	if topic.Lastreplyuid != 0 {
		topicMap["lastreplyusername"] = lastReplyUser.Username
	}

	if topic.EditorUid != 0 {
		editorUser := DefaultUser.FindOne(ctx, "uid", topic.EditorUid)
		topicMap["editor_username"] = editorUser.Username
	}

	return
}

// FindByPage 获取列表（分页）：后台用
func (TopicLogic) FindByPage(ctx context.Context, conds map[string]string, curPage, limit int) ([]*model.Topic, int) {
	objLog := GetLogger(ctx)

	filter := bson.M{}
	for k, v := range conds {
		filter[k] = v
	}

	total, err := db.GetCollection("topics").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("find count error:", err)
		return nil, 0
	}

	offset := (curPage - 1) * limit
	findOpts := options.Find().
		SetSort(bson.M{"_id": -1}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := db.GetCollection("topics").Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("find error:", err)
		return nil, 0
	}
	defer cursor.Close(ctx)

	topicList := make([]*model.Topic, 0)
	if err = cursor.All(ctx, &topicList); err != nil {
		objLog.Errorln("find decode error:", err)
		return nil, 0
	}

	return topicList, int(total)
}

func (TopicLogic) FindAppend(ctx context.Context, tid int) []*model.TopicAppend {
	objLog := GetLogger(ctx)

	cursor, err := db.GetCollection("topic_append").Find(ctx, bson.M{"tid": tid})
	if err != nil {
		objLog.Errorln("TopicLogic FindAppend error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	topicAppends := make([]*model.TopicAppend, 0)
	if err = cursor.All(ctx, &topicAppends); err != nil {
		objLog.Errorln("TopicLogic FindAppend decode error:", err)
	}

	return topicAppends
}

func (TopicLogic) findByTid(tid int) *model.Topic {
	topic := &model.Topic{}
	ctx := context.Background()
	err := db.GetCollection("topics").FindOne(ctx, bson.M{"_id": tid}).Decode(topic)
	if err != nil {
		logger.Errorln("TopicLogic findByTid error:", err)
	}
	return topic
}

// findByTids 获取多个主题详细信息 包内用
func (TopicLogic) findByTids(tids []int) map[int]*model.Topic {
	if len(tids) == 0 {
		return nil
	}

	ctx := context.Background()
	cursor, err := db.GetCollection("topics").Find(ctx, bson.M{"_id": bson.M{"$in": tids}})
	if err != nil {
		logger.Errorln("TopicLogic findByTids error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	topicList := make([]*model.Topic, 0)
	if err = cursor.All(ctx, &topicList); err != nil {
		logger.Errorln("TopicLogic findByTids decode error:", err)
		return nil
	}

	topicMap := make(map[int]*model.Topic, len(topicList))
	for _, t := range topicList {
		topicMap[t.Tid] = t
	}
	return topicMap
}

func (TopicLogic) fillDataForTopicInfo(topicInfos []*model.TopicInfo) []map[string]interface{} {
	uidSet := set.New(set.NonThreadSafe)
	nidSet := set.New(set.NonThreadSafe)
	for _, topicInfo := range topicInfos {
		uidSet.Add(topicInfo.Uid)
		if topicInfo.Lastreplyuid != 0 {
			uidSet.Add(topicInfo.Lastreplyuid)
		}
		nidSet.Add(topicInfo.Nid)
	}

	usersMap := DefaultUser.FindUserInfos(nil, set.IntSlice(uidSet))
	nodes := GetNodesByNids(set.IntSlice(nidSet))

	data := make([]map[string]interface{}, len(topicInfos))

	for i, topicInfo := range topicInfos {
		dest := make(map[string]interface{})

		if topicInfo.Lastreplyuid != 0 {
			if user, ok := usersMap[topicInfo.Lastreplyuid]; ok {
				dest["lastreplyusername"] = user.Username
			}
		}

		structs.FillMap(topicInfo.Topic, dest)
		structs.FillMap(topicInfo.TopicEx, dest)

		dest["user"] = usersMap[topicInfo.Uid]
		dest["node"] = nodes[topicInfo.Nid]

		data[i] = dest
	}

	return data
}

var (
	hotNodesCache  []map[string]interface{}
	hotNodesBegin  time.Time
	hotNodesLocker sync.Mutex
)

// FindHotNodes 获得热门节点
func (TopicLogic) FindHotNodes(ctx context.Context) []map[string]interface{} {
	hotNodesLocker.Lock()
	defer hotNodesLocker.Unlock()
	if !hotNodesBegin.IsZero() && hotNodesBegin.Add(1*time.Hour).Before(time.Now()) {
		return hotNodesCache
	}

	objLog := GetLogger(ctx)

	hotNum := 10

	lastWeek := time.Now().Add(-7 * 24 * time.Hour)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"ctime": bson.M{"$gte": lastWeek}}}},
		{{Key: "$group", Value: bson.M{"_id": "$nid", "topicnum": bson.M{"$sum": 1}}}},
		{{Key: "$sort", Value: bson.M{"topicnum": -1}}},
		{{Key: "$limit", Value: 15}},
	}

	cursor, err := db.GetCollection("topics").Aggregate(ctx, pipeline)
	if err != nil {
		objLog.Errorln("TopicLogic FindHotNodes error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	type aggResult struct {
		Nid      int `bson:"_id"`
		TopicNum int `bson:"topicnum"`
	}

	nids := make([]int, 0, 15)
	for cursor.Next(ctx) {
		var result aggResult
		if err = cursor.Decode(&result); err != nil {
			objLog.Errorln("FindHotNodes decode error:", err)
			continue
		}
		nids = append(nids, result.Nid)
	}

	nodes := make([]map[string]interface{}, 0, hotNum)

	topicNodes := GetNodesByNids(nids)
	for _, nid := range nids {
		topicNode := topicNodes[nid]
		if !topicNode.ShowIndex {
			continue
		}

		node := map[string]interface{}{
			"name":  topicNode.Name,
			"ename": topicNode.Ename,
			"nid":   topicNode.Nid,
		}
		nodes = append(nodes, node)
		if len(nodes) == hotNum {
			break
		}
	}

	hotNodesCache = nodes
	hotNodesBegin = time.Now()

	return nodes
}

// Total 话题总数
func (TopicLogic) Total() int64 {
	ctx := context.Background()
	total, err := db.GetCollection("topics").CountDocuments(ctx, bson.M{})
	if err != nil {
		logger.Errorln("TopicLogic Total error:", err)
	}
	return total
}

// JSEscape 安全过滤
func (TopicLogic) JSEscape(topics []*model.Topic) []*model.Topic {
	for i, topic := range topics {
		topics[i].Title = template.JSEscapeString(topic.Title)
		topics[i].Content = template.JSEscapeString(topic.Content)
	}
	return topics
}

func (TopicLogic) Count(ctx context.Context, querystring string, args ...interface{}) int64 {
	objLog := GetLogger(ctx)

	filter := bson.M{"flag": bson.M{"$lt": model.FlagAuditDelete}}
	if querystring != "" {
		extra := buildFilter(querystring, args...)
		for k, v := range extra {
			filter[k] = v
		}
	}

	total, err := db.GetCollection("topics").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("TopicLogic Count error:", err)
	}

	return total
}

// getOwner 通过tid获得话题的所有者
func (TopicLogic) getOwner(tid int) int {
	topic := &model.Topic{}
	ctx := context.Background()
	err := db.GetCollection("topics").FindOne(ctx, bson.M{"_id": tid}).Decode(topic)
	if err != nil {
		logger.Errorln("topic logic getOwner Error:", err)
		return 0
	}
	return topic.Uid
}

func (TopicLogic) decodeTopicContent(ctx context.Context, topic *model.Topic) string {
	content := util.EmbedWide(topic.Content)
	return parseAtUser(ctx, content)
}

func (TopicLogic) addFlagFilter(filter bson.M) bson.M {
	filter["flag"] = bson.M{"$lt": model.FlagAuditDelete}
	return filter
}

// buildTopicSort 将 "field DESC" 形式的排序转为 bson.M
func buildTopicSort(orderBy string) bson.M {
	if orderBy == "" {
		return bson.M{"_id": -1}
	}
	sort := bson.M{}
	parts := splitOrderBy(orderBy)
	for _, p := range parts {
		field := p[0]
		if field == "tid" {
			field = "_id"
		}
		dir := 1
		if len(p) > 1 && (p[1] == "DESC" || p[1] == "desc") {
			dir = -1
		}
		sort[field] = dir
	}
	return sort
}

func splitOrderBy(s string) [][]string {
	var result [][]string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := trimSpace(s[start:i])
			if part != "" {
				result = append(result, splitSpace(part))
			}
			start = i + 1
		}
	}
	return result
}

func trimSpace(s string) string {
	i, j := 0, len(s)
	for i < j && s[i] == ' ' {
		i++
	}
	for j > i && s[j-1] == ' ' {
		j--
	}
	return s[i:j]
}

func splitSpace(s string) []string {
	var parts []string
	start := -1
	for i, c := range s {
		if c == ' ' || c == '\t' {
			if start >= 0 {
				parts = append(parts, s[start:i])
				start = -1
			}
		} else {
			if start < 0 {
				start = i
			}
		}
	}
	if start >= 0 {
		parts = append(parts, s[start:])
	}
	return parts
}

// 话题回复（评论）
type TopicComment struct{}

// UpdateComment 更新该主题的回复信息
func (self TopicComment) UpdateComment(cid, objid, uid int, cmttime time.Time) {
	ctx := context.Background()

	_, err := db.GetCollection("topics").UpdateOne(ctx, bson.M{"_id": objid}, bson.M{
		"$set": bson.M{
			"lastreplyuid":  uid,
			"lastreplytime": cmttime,
		},
	})
	if err != nil {
		logger.Errorln("更新主题最后回复人信息失败：", err)
		return
	}

	_, err = db.GetCollection("topics_ex").UpdateOne(ctx, bson.M{"tid": objid}, bson.M{
		"$inc": bson.M{"reply": 1},
	})
	if err != nil {
		logger.Errorln("更新主题回复数失败：", err)
		return
	}
}

func (self TopicComment) String() string {
	return "topic"
}

// SetObjinfo 实现 CommentObjecter 接口
func (self TopicComment) SetObjinfo(ids []int, commentMap map[int][]*model.Comment) {

	topics := DefaultTopic.FindByTids(ids)
	if len(topics) == 0 {
		return
	}

	for _, topic := range topics {
		if topic.Flag > model.FlagNormal {
			continue
		}
		objinfo := make(map[string]interface{})
		objinfo["title"] = topic.Title
		objinfo["uri"] = model.PathUrlMap[model.TypeTopic]
		objinfo["type_name"] = model.TypeNameMap[model.TypeTopic]

		for _, comment := range commentMap[topic.Tid] {
			comment.Objinfo = objinfo
		}
	}
}

// 主题喜欢
type TopicLike struct{}

// UpdateLike 更新该主题的喜欢数
func (self TopicLike) UpdateLike(objid, num int) {
	ctx := context.Background()
	_, err := db.GetCollection("topics_ex").UpdateOne(ctx, bson.M{"tid": objid}, bson.M{
		"$inc": bson.M{"like": num},
	})
	if err != nil {
		logger.Errorln("更新主题喜欢数失败：", err)
	}
}

func (self TopicLike) String() string {
	return "topic"
}
