// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/fatih/structs"
	"github.com/polaris1119/logger"
	"github.com/polaris1119/set"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

type ResourceLogic struct{}

var DefaultResource = ResourceLogic{}

// Publish 增加（修改）资源
func (ResourceLogic) Publish(ctx context.Context, me *model.Me, form url.Values) (err error) {
	objLog := GetLogger(ctx)

	uid := me.Uid
	resource := &model.Resource{}

	if form.Get("id") != "" {
		id, _ := strconv.Atoi(form.Get("id"))
		err = db.GetCollection("resource").FindOne(ctx, bson.M{"_id": id}).Decode(resource)
		if err != nil {
			logger.Errorln("ResourceLogic Publish find error:", err)
			return
		}

		if !CanEdit(me, resource) {
			err = NotModifyAuthorityErr
			return
		}

		if form.Get("form") == model.LinkForm {
			form.Set("content", "")
		} else {
			form.Set("url", "")
		}

		err = schemaDecoder.Decode(resource, form)
		if err != nil {
			objLog.Errorln("ResourceLogic Publish decode error:", err)
			return
		}
		_, err = db.GetCollection("resource").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": resource})
		if err != nil {
			objLog.Errorf("更新资源 【%d】 信息失败：%s\n", id, err)
			return
		}

		go modifyObservable.NotifyObservers(uid, model.TypeResource, resource.Id)

	} else {

		err = schemaDecoder.Decode(resource, form)
		if err != nil {
			objLog.Errorln("ResourceLogic Publish decode error:", err)
			return
		}

		resource.Uid = uid
		resource.BeforeInsert()

		newID, idErr := db.NextID("resource")
		if idErr != nil {
			err = idErr
			objLog.Errorln("ResourceLogic Publish NextID error:", err)
			return
		}
		resource.Id = newID

		session, sessErr := db.GetClient().StartSession()
		if sessErr != nil {
			err = sessErr
			objLog.Errorln("ResourceLogic Publish StartSession error:", err)
			return
		}
		defer session.EndSession(ctx)

		_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
			_, insertErr := db.GetCollection("resource").InsertOne(sc, resource)
			if insertErr != nil {
				return nil, insertErr
			}

			resourceEx := &model.ResourceEx{
				Id: resource.Id,
			}
			_, insertErr = db.GetCollection("resource_ex").InsertOne(sc, resourceEx)
			if insertErr != nil {
				return nil, insertErr
			}
			return nil, nil
		})
		if err != nil {
			objLog.Errorln("ResourceLogic Publish transaction error:", err)
			return
		}

		// 发布动态
		resourceEx := &model.ResourceEx{Id: resource.Id}
		DefaultFeed.publish(resource, resourceEx, me)

		// 给 被@用户 发系统消息
		ext := map[string]interface{}{
			"objid":   resource.Id,
			"objtype": model.TypeResource,
			"uid":     uid,
			"msgtype": model.MsgtypePublishAtMe,
		}
		go DefaultMessage.SendSysMsgAtUsernames(ctx, form.Get("usernames"), ext, 0)

		go publishObservable.NotifyObservers(uid, model.TypeResource, resource.Id)
	}

	return
}

// Total 资源总数
func (ResourceLogic) Total() int64 {
	ctx := context.Background()
	total, err := db.GetCollection("resource").CountDocuments(ctx, bson.M{})
	if err != nil {
		logger.Errorln("CommentLogic Total error:", err)
	}
	return total
}

// FindBy 获取资源列表（分页）
func (ResourceLogic) FindBy(ctx context.Context, limit int, lastIds ...int) []*model.Resource {
	objLog := GetLogger(ctx)

	filter := bson.M{}
	if len(lastIds) > 0 && lastIds[0] > 0 {
		filter["_id"] = bson.M{"$lt": lastIds[0]}
	}

	findOpts := options.Find().
		SetSort(bson.M{"_id": -1}).
		SetLimit(int64(limit))

	cursor, err := db.GetCollection("resource").Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("ResourceLogic FindBy Error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	resourceList := make([]*model.Resource, 0)
	if err = cursor.All(ctx, &resourceList); err != nil {
		objLog.Errorln("ResourceLogic FindBy decode error:", err)
		return nil
	}

	return resourceList
}

// FindAll 获得资源列表（完整信息），分页
func (self ResourceLogic) FindAll(ctx context.Context, paginator *Paginator, orderBy, querystring string, args ...interface{}) (resources []map[string]interface{}, total int64) {
	objLog := GetLogger(ctx)

	count := paginator.PerPage()

	filter := bson.M{}
	if querystring != "" {
		filter = buildFilter(querystring, args...)
	}

	total = self.Count(ctx, querystring, args...)

	findOpts := options.Find().
		SetSort(buildResourceSort(orderBy)).
		SetSkip(int64(paginator.Offset())).
		SetLimit(int64(count))

	cursor, err := db.GetCollection("resource").Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("ResourceLogic FindAll error:", err)
		return
	}
	defer cursor.Close(ctx)

	resourceList := make([]*model.Resource, 0)
	if err = cursor.All(ctx, &resourceList); err != nil {
		objLog.Errorln("ResourceLogic FindAll decode error:", err)
		return
	}

	ids := make([]int, len(resourceList))
	for i, r := range resourceList {
		ids[i] = r.Id
	}

	resourceExMap := make(map[int]*model.ResourceEx)
	if len(ids) > 0 {
		exCursor, exErr := db.GetCollection("resource_ex").Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
		if exErr == nil {
			defer exCursor.Close(ctx)
			exList := make([]*model.ResourceEx, 0)
			if exCursor.All(ctx, &exList) == nil {
				for _, ex := range exList {
					resourceExMap[ex.Id] = ex
				}
			}
		}
	}

	resourceInfos := make([]*model.ResourceInfo, 0, len(resourceList))
	for _, r := range resourceList {
		info := &model.ResourceInfo{Resource: *r}
		if ex, ok := resourceExMap[r.Id]; ok {
			info.ResourceEx = *ex
		}
		resourceInfos = append(resourceInfos, info)
	}

	uidSet := set.New(set.NonThreadSafe)
	for _, resourceInfo := range resourceInfos {
		uidSet.Add(resourceInfo.Uid)
	}

	usersMap := DefaultUser.FindUserInfos(ctx, set.IntSlice(uidSet))

	resources = make([]map[string]interface{}, len(resourceInfos))

	for i, resourceInfo := range resourceInfos {
		dest := make(map[string]interface{})

		structs.FillMap(resourceInfo.Resource, dest)
		structs.FillMap(resourceInfo.ResourceEx, dest)

		dest["user"] = usersMap[resourceInfo.Uid]

		if resourceInfo.Form == model.LinkForm {
			urlObj, err := url.Parse(resourceInfo.Url)
			if err == nil {
				dest["host"] = urlObj.Host
			}
		} else {
			dest["url"] = "/resources/" + strconv.Itoa(resourceInfo.Resource.Id)
		}

		resources[i] = dest
	}

	return
}

func (ResourceLogic) Count(ctx context.Context, querystring string, args ...interface{}) int64 {
	objLog := GetLogger(ctx)

	filter := bson.M{}
	if querystring != "" {
		filter = buildFilter(querystring, args...)
	}

	total, err := db.GetCollection("resource").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("ResourceLogic Count error:", err)
	}

	return total
}

// FindByCatid 获得某个分类的资源列表，分页
func (ResourceLogic) FindByCatid(ctx context.Context, paginator *Paginator, catid int) (resources []map[string]interface{}, total int64) {
	objLog := GetLogger(ctx)

	count := paginator.PerPage()
	filter := bson.M{"catid": catid}

	total, err := db.GetCollection("resource").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("ResourceLogic FindByCatid count error:", err)
		return
	}

	findOpts := options.Find().
		SetSort(bson.M{"mtime": -1}).
		SetSkip(int64(paginator.Offset())).
		SetLimit(int64(count))

	cursor, err := db.GetCollection("resource").Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("ResourceLogic FindByCatid error:", err)
		return
	}
	defer cursor.Close(ctx)

	resourceList := make([]*model.Resource, 0)
	if err = cursor.All(ctx, &resourceList); err != nil {
		objLog.Errorln("ResourceLogic FindByCatid decode error:", err)
		return
	}

	ids := make([]int, len(resourceList))
	for i, r := range resourceList {
		ids[i] = r.Id
	}

	resourceExMap := make(map[int]*model.ResourceEx)
	if len(ids) > 0 {
		exCursor, exErr := db.GetCollection("resource_ex").Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
		if exErr == nil {
			defer exCursor.Close(ctx)
			exList := make([]*model.ResourceEx, 0)
			if exCursor.All(ctx, &exList) == nil {
				for _, ex := range exList {
					resourceExMap[ex.Id] = ex
				}
			}
		}
	}

	resourceInfos := make([]*model.ResourceInfo, 0, len(resourceList))
	for _, r := range resourceList {
		info := &model.ResourceInfo{Resource: *r}
		if ex, ok := resourceExMap[r.Id]; ok {
			info.ResourceEx = *ex
		}
		resourceInfos = append(resourceInfos, info)
	}

	uidSet := set.New(set.NonThreadSafe)
	for _, resourceInfo := range resourceInfos {
		uidSet.Add(resourceInfo.Uid)
	}

	usersMap := DefaultUser.FindUserInfos(ctx, set.IntSlice(uidSet))

	resources = make([]map[string]interface{}, len(resourceInfos))

	for i, resourceInfo := range resourceInfos {
		dest := make(map[string]interface{})

		structs.FillMap(resourceInfo.Resource, dest)
		structs.FillMap(resourceInfo.ResourceEx, dest)

		dest["user"] = usersMap[resourceInfo.Uid]

		if resourceInfo.Form == model.LinkForm {
			urlObj, err := url.Parse(resourceInfo.Url)
			if err == nil {
				dest["host"] = urlObj.Host
			}
		} else {
			dest["url"] = "/resources/" + strconv.Itoa(resourceInfo.Resource.Id)
		}

		resources[i] = dest
	}

	return
}

// FindByIds 获取多个资源详细信息
func (ResourceLogic) FindByIds(ids []int) []*model.Resource {
	if len(ids) == 0 {
		return nil
	}

	ctx := context.Background()
	cursor, err := db.GetCollection("resource").Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		logger.Errorln("ResourceLogic FindByIds error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	resources := make([]*model.Resource, 0)
	if err = cursor.All(ctx, &resources); err != nil {
		logger.Errorln("ResourceLogic FindByIds decode error:", err)
		return nil
	}
	return resources
}

func (ResourceLogic) findById(id int) *model.Resource {
	resource := &model.Resource{}
	ctx := context.Background()
	err := db.GetCollection("resource").FindOne(ctx, bson.M{"_id": id}).Decode(resource)
	if err != nil {
		logger.Errorln("ResourceLogic findById error:", err)
	}
	return resource
}

// findByIds 获取多个资源详细信息 包内使用
func (ResourceLogic) findByIds(ids []int) map[int]*model.Resource {
	if len(ids) == 0 {
		return nil
	}

	ctx := context.Background()
	cursor, err := db.GetCollection("resource").Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		logger.Errorln("ResourceLogic FindByIds error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	resourceList := make([]*model.Resource, 0)
	if err = cursor.All(ctx, &resourceList); err != nil {
		logger.Errorln("ResourceLogic FindByIds decode error:", err)
		return nil
	}

	resources := make(map[int]*model.Resource, len(resourceList))
	for _, r := range resourceList {
		resources[r.Id] = r
	}
	return resources
}

// FindById 获得资源详细信息
func (ResourceLogic) FindById(ctx context.Context, id int) (resourceMap map[string]interface{}, comments []map[string]interface{}) {
	objLog := GetLogger(ctx)

	resource := &model.Resource{}
	err := db.GetCollection("resource").FindOne(ctx, bson.M{"_id": id}).Decode(resource)
	if err != nil {
		objLog.Errorln("ResourceLogic FindById error:", err)
		return
	}

	if resource.Id == 0 {
		objLog.Errorln("ResourceLogic FindById get error:", err)
		return
	}

	resourceEx := &model.ResourceEx{}
	_ = db.GetCollection("resource_ex").FindOne(ctx, bson.M{"_id": id}).Decode(resourceEx)

	resourceMap = make(map[string]interface{})
	structs.FillMap(resource, resourceMap)
	structs.FillMap(resourceEx, resourceMap)

	resourceMap["catname"] = GetCategoryName(resource.Catid)
	if resource.Form == model.LinkForm {
		urlObj, err := url.Parse(resource.Url)
		if err == nil {
			resourceMap["host"] = urlObj.Host
		}
	} else {
		resourceMap["url"] = "/resources/" + strconv.Itoa(resource.Id)
	}

	// 评论信息
	comments, ownerUser, _ := DefaultComment.FindObjComments(ctx, id, model.TypeResource, resource.Uid, 0)
	resourceMap["user"] = ownerUser
	return
}

// FindResource 获取单个 Resource 信息（用于编辑）
func (ResourceLogic) FindResource(ctx context.Context, id int) *model.Resource {
	objLog := GetLogger(ctx)

	resource := &model.Resource{}
	err := db.GetCollection("resource").FindOne(ctx, bson.M{"_id": id}).Decode(resource)
	if err != nil {
		objLog.Errorf("ResourceLogic FindResource [%d] error：%s\n", id, err)
	}

	return resource
}

// FindRecent 获得某个用户最近的资源
func (ResourceLogic) FindRecent(ctx context.Context, uid int) []*model.Resource {
	filter := bson.M{"uid": uid}
	findOpts := options.Find().
		SetSort(bson.M{"_id": -1}).
		SetLimit(5)

	cursor, err := db.GetCollection("resource").Find(ctx, filter, findOpts)
	if err != nil {
		logger.Errorln("resource logic FindRecent error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	resourceList := make([]*model.Resource, 0)
	if err = cursor.All(ctx, &resourceList); err != nil {
		logger.Errorln("resource logic FindRecent decode error:", err)
		return nil
	}

	return resourceList
}

// getOwner 通过id获得资源的所有者
func (ResourceLogic) getOwner(id int) int {
	resource := &model.Resource{}
	ctx := context.Background()
	err := db.GetCollection("resource").FindOne(ctx, bson.M{"_id": id}).Decode(resource)
	if err != nil {
		logger.Errorln("resource logic getOwner Error:", err)
		return 0
	}
	return resource.Uid
}

// buildResourceSort 将 "field DESC" 形式的排序转为 bson.M
func buildResourceSort(orderBy string) bson.M {
	if orderBy == "" {
		return bson.M{"_id": -1}
	}
	sort := bson.M{}
	parts := splitOrderBy(orderBy)
	for _, p := range parts {
		field := p[0]
		if field == "id" {
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

// 资源评论
type ResourceComment struct{}

// UpdateComment 更新该资源的评论信息
// cid：评论id；objid：被评论对象id；uid：评论者；cmttime：评论时间
func (self ResourceComment) UpdateComment(cid, objid, uid int, cmttime time.Time) {
	ctx := context.Background()

	_, err := db.GetCollection("resource").UpdateOne(ctx, bson.M{"_id": objid}, bson.M{
		"$set": bson.M{
			"lastreplyuid":  uid,
			"lastreplytime": cmttime,
		},
	})
	if err != nil {
		logger.Errorln("更新最后回复人信息失败：", err)
		return
	}

	_, err = db.GetCollection("resource_ex").UpdateOne(ctx, bson.M{"_id": objid}, bson.M{
		"$inc": bson.M{"cmtnum": 1},
	})
	if err != nil {
		logger.Errorln("更新资源评论数失败：", err)
		return
	}
}

func (self ResourceComment) String() string {
	return "resource"
}

// SetObjinfo 实现 CommentObjecter 接口
func (self ResourceComment) SetObjinfo(ids []int, commentMap map[int][]*model.Comment) {
	resources := DefaultResource.FindByIds(ids)
	if len(resources) == 0 {
		return
	}

	for _, resource := range resources {
		objinfo := make(map[string]interface{})
		objinfo["title"] = resource.Title
		objinfo["uri"] = model.PathUrlMap[model.TypeResource]
		objinfo["type_name"] = model.TypeNameMap[model.TypeResource]

		for _, comment := range commentMap[resource.Id] {
			comment.Objinfo = objinfo
		}
	}
}

// 资源喜欢
type ResourceLike struct{}

// UpdateLike 更新该主题的喜欢数
// objid：被喜欢对象id；num: 喜欢数(负数表示取消喜欢)
func (self ResourceLike) UpdateLike(objid, num int) {
	ctx := context.Background()
	_, err := db.GetCollection("resource_ex").UpdateOne(ctx, bson.M{"_id": objid}, bson.M{
		"$inc": bson.M{"likenum": num},
	})
	if err != nil {
		logger.Errorln("更新资源喜欢数失败：", err)
	}
}

func (self ResourceLike) String() string {
	return "resource"
}

func (ResourceLogic) Delete(ctx context.Context, id, uid int, isRoot bool) error {
	resource := &model.Resource{}
	err := db.GetCollection("resource").FindOne(ctx, bson.M{"_id": id}).Decode(resource)
	if err != nil {
		return errors.New("资源不存在")
	}
	if resource.Uid != uid && !isRoot {
		return errors.New("无权删除")
	}
	_, err = db.GetCollection("resource").DeleteOne(ctx, bson.M{"_id": id})
	return err
}
