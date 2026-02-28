// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/polaris1119/config"
	"github.com/polaris1119/logger"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/dao/cache"
	"github.com/studygolang/studygolang/internal/model"
	"github.com/studygolang/studygolang/util"

	"github.com/polaris1119/set"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FeedLogic struct{}

var DefaultFeed = FeedLogic{}

func (self FeedLogic) GetTotalCount(ctx context.Context) int64 {
	objLog := GetLogger(ctx)
	count, err := db.GetCollection("feed").CountDocuments(ctx, bson.M{"state": 0})
	if err != nil {
		objLog.Errorln("FeedLogic Count error:", err)
		return 0
	}
	return count
}

func (self FeedLogic) FindRecentWithPaginator(ctx context.Context, paginator *Paginator, tab string) []*model.Feed {
	objLog := GetLogger(ctx)

	feeds := cache.Feed.GetList(ctx, paginator.curPage)
	if len(feeds) > 0 {
		return feeds
	}

	feeds = make([]*model.Feed, 0)
	sort := bson.D{{"updated_at", -1}}
	if tab == model.TabRecommend {
		sort = bson.D{{"seq", -1}, {"updated_at", -1}}
	}
	opts := options.Find().
		SetSort(sort).
		SetLimit(int64(paginator.PerPage())).
		SetSkip(int64(paginator.Offset()))
	cursor, err := db.GetCollection("feed").Find(ctx, bson.M{}, opts)
	if err != nil {
		objLog.Errorln("FeedLogic FindRecent error:", err)
		return nil
	}
	if err = cursor.All(ctx, &feeds); err != nil {
		objLog.Errorln("FeedLogic FindRecent cursor error:", err)
		return nil
	}

	feeds = self.fillOtherInfo(ctx, feeds, true)
	if len(feeds) > 0 {
		cache.Feed.SetList(ctx, paginator.curPage, feeds)
	}
	return feeds
}

func (self FeedLogic) FindRecent(ctx context.Context, num int) []*model.Feed {
	objLog := GetLogger(ctx)

	feeds := make([]*model.Feed, 0)
	opts := options.Find().SetSort(bson.D{{"updated_at", -1}}).SetLimit(int64(num))
	cursor, err := db.GetCollection("feed").Find(ctx, bson.M{}, opts)
	if err != nil {
		objLog.Errorln("FeedLogic FindRecent error:", err)
		return nil
	}
	if err = cursor.All(ctx, &feeds); err != nil {
		objLog.Errorln("FeedLogic FindRecent cursor error:", err)
		return nil
	}

	return self.fillOtherInfo(ctx, feeds, true)
}

func (self FeedLogic) FindTop(ctx context.Context) []*model.Feed {
	objLog := GetLogger(ctx)

	feeds := cache.Feed.GetTop(ctx)
	if feeds != nil {
		return feeds
	}

	feeds = make([]*model.Feed, 0)
	opts := options.Find().SetSort(bson.D{{"updated_at", -1}})
	cursor, err := db.GetCollection("feed").Find(ctx, bson.M{"top": 1}, opts)
	if err != nil {
		objLog.Errorln("FeedLogic FindRecent error:", err)
		return nil
	}
	if err = cursor.All(ctx, &feeds); err != nil {
		objLog.Errorln("FeedLogic FindRecent cursor error:", err)
		return nil
	}

	feeds = self.fillOtherInfo(ctx, feeds, false)
	cache.Feed.SetTop(ctx, feeds)
	return feeds
}

// AutoUpdateSeq 自动更新动态的排序（校准）
func (self FeedLogic) AutoUpdateSeq() {
	curHour := time.Now().Hour()
	if curHour < 7 {
		return
	}

	feedDay := config.ConfigFile.MustInt("feed", "day", 3)
	cmtWeight := config.ConfigFile.MustInt("feed", "cmt_weight", 80)
	viewWeight := config.ConfigFile.MustInt("feed", "view_weight", 80)

	ctx := context.Background()
	offset, limit := int64(0), int64(100)
	for {
		feeds := make([]*model.Feed, 0)
		opts := options.Find().SetSkip(offset).SetLimit(limit)
		cursor, err := db.GetCollection("feed").Find(ctx, bson.M{"seq": bson.M{"$gt": 0}}, opts)
		if err != nil {
			return
		}
		if err = cursor.All(ctx, &feeds); err != nil || len(feeds) == 0 {
			return
		}

		offset += limit

		for _, feed := range feeds {
			if feed.State == model.FeedOffline {
				continue
			}

			// 当天（不到24小时）发布的，不降
			elapse := int(time.Now().Sub(time.Time(feed.CreatedAt)).Hours())
			if elapse < 24 {
				continue
			}

			if feed.Uid > 0 {
				user := DefaultUser.FindOne(nil, "uid", feed.Uid)
				if DefaultUser.IsAdmin(user) {
					elapse = int(time.Now().Sub(time.Time(feed.UpdatedAt)).Hours())
				}
			}

			seq := 0
			if elapse <= feedDay*24 {
				seq = self.calcChangeSeq(feed, cmtWeight, viewWeight)
			}

			db.GetCollection("feed").UpdateOne(ctx, bson.M{"_id": feed.Id}, bson.M{"$set": bson.M{
				"updated_at": time.Time(feed.UpdatedAt),
				"seq":        seq,
			}})
		}
	}
}

func (self FeedLogic) calcChangeSeq(feed *model.Feed, cmtWeight int, viewWeight int) int {
	seq := 0
	ctx := context.Background()

	// 最近有评论（时间更新）的，降 1/10 个评论数
	if int(time.Now().Sub(time.Time(feed.UpdatedAt)).Hours()) < 1 {
		seq = feed.Seq - cmtWeight/10
	} else {
		// 最近有没有其他变动（赞、阅读等）
		var updatedAt time.Time
		switch feed.Objtype {
		case model.TypeTopic:
			topicEx := &model.TopicEx{}
			db.GetCollection("topics_ex").FindOne(ctx, bson.M{"_id": feed.Objid}).Decode(topicEx)
			updatedAt = topicEx.Mtime
		case model.TypeArticle:
			article := &model.Article{}
			db.GetCollection("articles").FindOne(ctx, bson.M{"_id": feed.Objid}).Decode(article)
			updatedAt = time.Time(article.Mtime)
		case model.TypeResource:
			resourceEx := &model.ResourceEx{}
			db.GetCollection("resource_ex").FindOne(ctx, bson.M{"_id": feed.Objid}).Decode(resourceEx)
			updatedAt = resourceEx.Mtime
		case model.TypeProject:
			project := &model.OpenProject{}
			db.GetCollection("open_project").FindOne(ctx, bson.M{"_id": feed.Objid}).Decode(project)
			updatedAt = time.Time(project.Mtime)
		case model.TypeBook:
			book := &model.Book{}
			db.GetCollection("book").FindOne(ctx, bson.M{"_id": feed.Objid}).Decode(book)
			updatedAt = time.Time(book.UpdatedAt)
		}

		dynamicElapse := int(time.Now().Sub(updatedAt).Hours())

		if dynamicElapse < 1 {
			seq = feed.Seq - viewWeight*10
		} else {
			seq = feed.Seq / 2
		}
	}

	if seq < 20 {
		seq = 20
	}

	return seq
}

func (FeedLogic) fillOtherInfo(ctx context.Context, feeds []*model.Feed, filterTop bool) []*model.Feed {
	newFeeds := make([]*model.Feed, 0, len(feeds))

	uidSet := set.New(set.NonThreadSafe)
	nidSet := set.New(set.NonThreadSafe)
	for _, feed := range feeds {
		if feed.State == model.FeedOffline {
			continue
		}

		if filterTop && feed.Top == 1 {
			continue
		}

		newFeeds = append(newFeeds, feed)

		if feed.Uid > 0 {
			uidSet.Add(feed.Uid)
		}
		if feed.Lastreplyuid > 0 {
			uidSet.Add(feed.Lastreplyuid)
		}
		if feed.Objtype == model.TypeTopic {
			nidSet.Add(feed.Nid)
		} else if feed.Objtype == model.TypeResource {
			feed.Node = map[string]interface{}{
				"name": GetCategoryName(feed.Nid),
			}
		}

		feed.Uri = model.PathUrlMap[feed.Objtype] + strconv.Itoa(feed.Objid)
	}

	usersMap := DefaultUser.FindUserInfos(ctx, set.IntSlice(uidSet))
	nodesMap := GetNodesByNids(set.IntSlice(nidSet))
	for _, feed := range newFeeds {
		if _, ok := usersMap[feed.Uid]; ok {
			feed.User = usersMap[feed.Uid]
		}
		if _, ok := usersMap[feed.Lastreplyuid]; ok {
			feed.Lastreplyuser = usersMap[feed.Lastreplyuid]
		}

		if feed.Objtype == model.TypeTopic {
			if _, ok := nodesMap[feed.Nid]; ok {
				feed.Node = map[string]interface{}{}
				util.Struct2Map(feed.Node, nodesMap[feed.Nid])
			}
		}
	}

	return newFeeds
}

// publish 发布动态
func (FeedLogic) publish(object interface{}, objectExt interface{}, me *model.Me) {
	go model.PublishFeed(object, objectExt, me)
}

func (self FeedLogic) updateSeq(objid, objtype, cmtnum, likenum, viewnum int) {
	cmtWeight := config.ConfigFile.MustInt("feed", "cmt_weight", 80)
	likeWeight := config.ConfigFile.MustInt("feed", "like_weight", 60)
	viewWeight := config.ConfigFile.MustInt("feed", "view_weight", 5)

	go func() {
		ctx := context.Background()
		feed := &model.Feed{}
		err := db.GetCollection("feed").FindOne(ctx, bson.M{"objid": objid, "objtype": objtype}).Decode(feed)
		if err != nil {
			return
		}

		if feed.State == model.FeedOffline {
			return
		}

		feedDay := config.ConfigFile.MustInt("feed", "day", 3)
		elapse := int(time.Now().Sub(time.Time(feed.CreatedAt)).Hours())

		if feed.Uid > 0 {
			user := DefaultUser.FindOne(nil, "uid", feed.Uid)
			if DefaultUser.IsAdmin(user) {
				elapse = int(time.Now().Sub(time.Time(feed.UpdatedAt)).Hours())
			}
		}

		seq := 0

		if elapse > feedDay*24 {
			if feed.Seq == 0 {
				return
			}
		} else {
			if feed.Seq == 0 {
				seq = feedDay*24 - elapse + (feed.Cmtnum+cmtnum)*cmtWeight + likenum*likeWeight + viewnum*viewWeight
			} else {
				seq = feed.Seq + cmtnum*cmtWeight + likenum*likeWeight + viewnum*viewWeight
			}
		}

		_, err = db.GetCollection("feed").UpdateOne(ctx, bson.M{"objid": objid, "objtype": objtype}, bson.M{"$set": bson.M{
			"updated_at": time.Time(feed.UpdatedAt),
			"seq":        seq,
		}})

		if err != nil {
			logger.Errorln("update feed seq error:", err)
			return
		}
	}()
}

// setTop 置顶或取消置顶
func (FeedLogic) setTop(ctx context.Context, objid, objtype int, top int) error {
	_, err := db.GetCollection("feed").UpdateOne(ctx, bson.M{"objid": objid, "objtype": objtype}, bson.M{"$set": bson.M{
		"top": top,
	}})

	return err
}

// updateComment 更新动态评论数据
func (self FeedLogic) updateComment(objid, objtype, uid int, cmttime time.Time) {
	go func() {
		ctx := context.Background()
		db.GetCollection("feed").UpdateOne(ctx, bson.M{"objid": objid, "objtype": objtype}, bson.M{
			"$inc": bson.M{"cmtnum": 1},
			"$set": bson.M{
				"lastreplyuid":  uid,
				"lastreplytime": cmttime,
			},
		})

		self.updateSeq(objid, objtype, 1, 0, 0)
	}()
}

// updateLike 更新动态赞数据
func (self FeedLogic) updateLike(objid, objtype, uid, num int) {
	go func() {
		ctx := context.Background()
		db.GetCollection("feed").UpdateOne(ctx, bson.M{"objid": objid, "objtype": objtype}, bson.M{
			"$inc": bson.M{"likenum": num},
		})
	}()
	self.updateSeq(objid, objtype, 0, num, 0)
}

func (self FeedLogic) modifyTopicNode(tid, nid int) {
	go func() {
		ctx := context.Background()
		change := bson.M{
			"nid": nid,
		}

		node := &model.TopicNode{}
		err := db.GetCollection("topics_node").FindOne(ctx, bson.M{"_id": nid}).Decode(node)
		if err == nil && !node.ShowIndex {
			change["state"] = model.FeedOffline
		}
		db.GetCollection("feed").UpdateOne(ctx, bson.M{"objid": tid, "objtype": model.TypeTopic}, bson.M{"$set": change})
	}()
}
