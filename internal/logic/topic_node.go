// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"net/url"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TopicNodeLogic struct{}

var DefaultNode = TopicNodeLogic{}

func (self TopicNodeLogic) FindOne(nid int) *model.TopicNode {
	topicNode := &model.TopicNode{}
	err := db.GetCollection("topics_node").FindOne(context.Background(), bson.M{"_id": nid}).Decode(topicNode)
	if err != nil {
		logger.Errorln("TopicNodeLogic FindOne error:", err, "nid:", nid)
	}

	return topicNode
}

func (self TopicNodeLogic) FindByEname(ename string) *model.TopicNode {
	topicNode := &model.TopicNode{}
	err := db.GetCollection("topics_node").FindOne(context.Background(), bson.M{"ename": ename}).Decode(topicNode)
	if err != nil {
		logger.Errorln("TopicNodeLogic FindByEname error:", err, "ename:", ename)
	}

	return topicNode
}

func (self TopicNodeLogic) FindByNids(nids []int) map[int]*model.TopicNode {
	if len(nids) == 0 {
		return nil
	}
	ctx := context.Background()
	nodeList := make(map[int]*model.TopicNode)
	nodes := make([]*model.TopicNode, 0)

	cursor, err := db.GetCollection("topics_node").Find(ctx, bson.M{"_id": bson.M{"$in": nids}})
	if err != nil {
		logger.Errorln("TopicNodeLogic FindByNids error:", err, "nids:", nids)
		return nodeList
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &nodes)

	for _, node := range nodes {
		nodeList[node.Nid] = node
	}

	return nodeList
}

func (self TopicNodeLogic) FindByParent(pid, num int) []*model.TopicNode {
	ctx := context.Background()
	nodeList := make([]*model.TopicNode, 0)
	opts := options.Find().SetLimit(int64(num))
	cursor, err := db.GetCollection("topics_node").Find(ctx, bson.M{"parent": pid}, opts)
	if err != nil {
		logger.Errorln("TopicNodeLogic FindByParent error:", err, "parent:", pid)
		return nodeList
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &nodeList)

	return nodeList
}

func (self TopicNodeLogic) FindAll(ctx context.Context) []*model.TopicNode {
	nodeList := make([]*model.TopicNode, 0)
	opts := options.Find().SetSort(bson.D{{Key: "seq", Value: 1}})
	cursor, err := db.GetCollection("topics_node").Find(ctx, bson.M{}, opts)
	if err != nil {
		logger.Errorln("TopicNodeLogic FindAll error:", err)
		return nodeList
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &nodeList)

	return nodeList
}

func (self TopicNodeLogic) Modify(ctx context.Context, form url.Values) error {
	objLog := GetLogger(ctx)

	node := &model.TopicNode{}
	err := schemaDecoder.Decode(node, form)
	if err != nil {
		objLog.Errorln("TopicNodeLogic Modify decode error:", err)
		return err
	}

	nid := goutils.MustInt(form.Get("nid"))
	if nid == 0 {
		var id int
		id, err = db.NextID("topics_node")
		if err != nil {
			objLog.Errorln("TopicNodeLogic Modify NextID error:", err)
			return err
		}
		node.Nid = id
		_, err = db.GetCollection("topics_node").InsertOne(ctx, node)
		if err != nil {
			objLog.Errorln("TopicNodeLogic Modify insert error:", err)
		}
		return err
	}

	change := bson.M{}
	fields := []string{"parent", "logo", "name", "ename", "intro", "seq", "show_index"}
	for _, field := range fields {
		change[field] = form.Get(field)
	}

	_, err = db.GetCollection("topics_node").UpdateOne(ctx, bson.M{"_id": nid}, bson.M{"$set": change})
	if err != nil {
		objLog.Errorln("TopicNodeLogic Modify update error:", err)
	}
	return err
}

func (self TopicNodeLogic) ModifySeq(ctx context.Context, nid, seq int) error {
	_, err := db.GetCollection("topics_node").UpdateOne(ctx, bson.M{"_id": nid}, bson.M{"$set": bson.M{"seq": seq}})
	return err
}

func (self TopicNodeLogic) FindParallelTree(ctx context.Context) []*model.TopicNode {
	nodeList := make([]*model.TopicNode, 0)
	opts := options.Find().SetSort(bson.D{{Key: "parent", Value: 1}, {Key: "seq", Value: 1}})
	cursor, err := db.GetCollection("topics_node").Find(ctx, bson.M{}, opts)
	if err != nil {
		logger.Errorln("TopicNodeLogic FindTreeList error:", err)
		return nil
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &nodeList)

	showNodeList := make([]*model.TopicNode, 0, len(nodeList))
	self.tileNodes(&showNodeList, nodeList, 0, 1, 3, 0)

	return showNodeList
}

func (self TopicNodeLogic) tileNodes(showNodeList *[]*model.TopicNode, nodeList []*model.TopicNode, parentId, curLevel, showLevel, pos int) {
	for num := len(nodeList); pos < num; pos++ {
		node := nodeList[pos]

		if node.Parent == parentId {
			*showNodeList = append(*showNodeList, node)

			if node.Level == 0 {
				node.Level = curLevel
			}

			if curLevel <= showLevel {
				self.tileNodes(showNodeList, nodeList, node.Nid, curLevel+1, showLevel, pos+1)
			}
		}

		if node.Parent > parentId {
			break
		}
	}
}
