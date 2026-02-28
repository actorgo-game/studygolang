// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"errors"
	"sync"

	"github.com/polaris1119/logger"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 常驻内存数据（多实例部署时，数据同步会有问题）

var (
	authLocker  sync.RWMutex
	Authorities []*model.Authority

	roleAuthLocker  sync.RWMutex
	RoleAuthorities map[int][]int

	roleLocker sync.RWMutex
	Roles      []*model.Role // 相应的 roleid-1 为索引

	nodeRWMutex sync.RWMutex
	// 节点信息
	AllNode []map[string]interface{}
	// 推荐节点
	AllRecommendNodes []map[string][]map[string]interface{}

	catRWMutex sync.RWMutex
	// 资源分类
	AllCategory []*model.ResourceCat

	WebsiteSetting = model.WebsiteSetting

	DefaultAvatars []string

	userSettingLocker sync.RWMutex
	UserSetting       map[string]int
)

// 将所有 权限 加载到内存中；后台修改权限时，重新加载一次
func LoadAuthorities() error {
	ctx := context.Background()
	authorities := make([]*model.Authority, 0)
	cursor, err := db.GetCollection("authority").Find(ctx, bson.M{})
	if err != nil {
		logger.Errorln("LoadAuthorities authority read fail:", err)
		return err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &authorities); err != nil {
		logger.Errorln("LoadAuthorities authority decode fail:", err)
		return err
	}

	authLocker.Lock()
	defer authLocker.Unlock()

	Authorities = authorities

	logger.Infoln("LoadAuthorities successfully!")

	return nil
}

// 将所有 角色拥有的权限 加载到内存中；后台修改时，重新加载一次
func LoadRoleAuthorities() error {
	ctx := context.Background()
	roleAuthorities := make([]*model.RoleAuthority, 0)
	cursor, err := db.GetCollection("role_authority").Find(ctx, bson.M{})
	if err != nil {
		logger.Errorln("LoadRoleAuthorities role_authority read fail:", err)
		return err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &roleAuthorities); err != nil {
		logger.Errorln("LoadRoleAuthorities role_authority decode fail:", err)
		return err
	}

	roleAuthLocker.Lock()
	defer roleAuthLocker.Unlock()

	RoleAuthorities = make(map[int][]int)

	for _, roleAuth := range roleAuthorities {
		roleId := roleAuth.Roleid

		if authorities, ok := RoleAuthorities[roleId]; ok {
			RoleAuthorities[roleId] = append(authorities, roleAuth.Aid)
		} else {
			RoleAuthorities[roleId] = []int{roleAuth.Aid}
		}
	}

	logger.Infoln("LoadRoleAuthorities successfully!")

	return nil
}

// 将所有 角色 加载到内存中；后台修改角色时，重新加载一次
func LoadRoles() error {
	ctx := context.Background()
	roles := make([]*model.Role, 0)
	cursor, err := db.GetCollection("role").Find(ctx, bson.M{})
	if err != nil {
		logger.Errorln("LoadRoles role read fail:", err)
		return err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &roles); err != nil {
		logger.Errorln("LoadRoles role decode fail:", err)
		return err
	}

	if len(roles) == 0 {
		logger.Errorln("LoadRoles role read fail: num is 0")
		return errors.New("no role")
	}

	roleLocker.Lock()
	defer roleLocker.Unlock()

	maxRoleid := roles[len(roles)-1].Roleid
	Roles = make([]*model.Role, maxRoleid)

	// 由于角色不多，而且一般角色id是连续自增的，因此这里以角色id当slice的index
	for _, role := range roles {
		Roles[role.Roleid-1] = role
	}

	logger.Infoln("LoadRoles successfully!")

	return nil
}

// 将所有 节点信息 加载到内存中：后台修改节点时，重新加载一次
func LoadNodes() error {
	// 如果有 推荐 节点，加载推荐节点
	hadRecommend := loadRecommendNodes()
	if hadRecommend {
		return nil
	}

	ctx := context.Background()
	nodeList := make([]*model.TopicNode, 0)
	cursor, err := db.GetCollection("topics_node").Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"seq": 1}))
	if err != nil {
		logger.Errorln("LoadNodes node read fail:", err)
		return err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &nodeList); err != nil {
		logger.Errorln("LoadNodes node decode fail:", err)
		return err
	}

	nodeNum := len(nodeList)
	tmpNodeList := make(map[int]*model.TopicNode, nodeNum)
	for _, node := range nodeList {
		tmpNodeList[node.Nid] = node
	}

	nodeRWMutex.Lock()
	defer nodeRWMutex.Unlock()

	AllNode = make([]map[string]interface{}, nodeNum)
	for i, node := range nodeList {
		nodeMap := make(map[string]interface{}, 5)
		nodeMap["pid"] = node.Parent
		if node.Parent == 0 {
			nodeMap["parent"] = "根节点"
		} else {
			nodeMap["parent"] = tmpNodeList[node.Parent].Name
		}
		nodeMap["nid"] = node.Nid
		nodeMap["logo"] = node.Logo
		nodeMap["name"] = node.Name
		nodeMap["ename"] = node.Ename
		nodeMap["intro"] = node.Intro
		nodeMap["show_index"] = node.ShowIndex
		nodeMap["ctime"] = node.Ctime
		AllNode[i] = nodeMap
	}

	logger.Infoln("LoadNodes successfully!")

	return nil
}

func LoadWebsiteSetting() error {
	ctx := context.Background()
	err := db.GetCollection("website_setting").FindOne(ctx, bson.M{}).Decode(WebsiteSetting)
	if err != nil {
		logger.Errorln("LoadWebsiteSetting read fail:", err)
		return err
	}
	WebsiteSetting.AfterLoad()

	logger.Infoln("LoadWebsiteSetting successfully!")

	return nil
}

func LoadUserSetting() error {
	ctx := context.Background()
	userSettings := make([]*model.UserSetting, 0)
	cursor, err := db.GetCollection("user_setting").Find(ctx, bson.M{})
	if err != nil {
		logger.Errorln("LoadUserSetting Find fail:", err)
		return err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &userSettings); err != nil {
		logger.Errorln("LoadUserSetting decode fail:", err)
		return err
	}

	userSettingLocker.Lock()
	defer userSettingLocker.Unlock()

	UserSetting = make(map[string]int)
	for _, userSetting := range userSettings {
		UserSetting[userSetting.Key] = userSetting.Value
	}

	logger.Infoln("LoadUserSetting successfully!")

	return nil
}

func LoadDefaultAvatar() error {
	ctx := context.Background()
	defaultAvatars := make([]*model.DefaultAvatar, 0)
	cursor, err := db.GetCollection("default_avatar").Find(ctx, bson.M{})
	if err != nil {
		logger.Errorln("LoadDefaultAvatar Find fail:", err)
		return err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &defaultAvatars); err != nil {
		logger.Errorln("LoadDefaultAvatar decode fail:", err)
		return err
	}

	DefaultAvatars = make([]string, len(defaultAvatars))
	for i, defaultAvatar := range defaultAvatars {
		DefaultAvatars[i] = defaultAvatar.Filename
	}

	logger.Infoln("LoadDefaultAvatar successfully!")

	return nil
}

// 获得单个节点名
func GetNodeName(nid int) string {
	if len(AllRecommendNodes) > 0 {
		return DefaultNode.FindOne(nid).Name
	}

	nodeRWMutex.RLock()
	defer nodeRWMutex.RUnlock()
	for _, node := range AllNode {
		if node["nid"].(int) == nid {
			return node["name"].(string)
		}
	}
	return ""
}

// 通过 ename 获得单个节点
func GetNodeByEname(ename string) map[string]interface{} {
	if len(AllRecommendNodes) > 0 {
		node := DefaultNode.FindByEname(ename)
		return map[string]interface{}{
			"ename":      node.Ename,
			"name":       node.Name,
			"pid":        node.Parent,
			"nid":        node.Nid,
			"logo":       node.Logo,
			"show_index": node.ShowIndex,
			"intro":      node.Intro,
		}
	}

	nodeRWMutex.RLock()
	defer nodeRWMutex.RUnlock()
	for _, node := range AllNode {
		if node["ename"].(string) == ename {
			return node
		}
	}
	return nil
}

// 通过 ename 获得 nid
func GetNidByEname(ename string) int {
	if len(AllRecommendNodes) > 0 {
		return DefaultNode.FindByEname(ename).Nid
	}

	nodeRWMutex.RLock()
	defer nodeRWMutex.RUnlock()
	for _, node := range AllNode {
		if node["ename"].(string) == ename {
			return node["nid"].(int)
		}
	}
	return 0
}

// 获得单个节点信息
func GetNode(nid int) map[string]interface{} {
	if len(AllRecommendNodes) > 0 {
		node := DefaultNode.FindOne(nid)
		return map[string]interface{}{
			"ename":      node.Ename,
			"pid":        node.Parent,
			"name":       node.Name,
			"nid":        node.Nid,
			"logo":       node.Logo,
			"intro":      node.Intro,
			"show_index": node.ShowIndex,
		}
	}

	nodeRWMutex.RLock()
	defer nodeRWMutex.RUnlock()
	for _, node := range AllNode {
		if node["nid"].(int) == nid {
			return node
		}
	}
	return nil
}

// 获得多个节点
func GetNodesByNids(nids []int) map[int]*model.TopicNode {
	if len(AllRecommendNodes) > 0 {
		return DefaultNode.FindByNids(nids)
	}

	nodes := make(map[int]*model.TopicNode, len(nids))
	nodeRWMutex.RLock()
	defer nodeRWMutex.RUnlock()
	for _, nid := range nids {
		for _, node := range AllNode {
			if node["nid"].(int) == nid {
				nodes[nid] = &model.TopicNode{
					Nid:       nid,
					Name:      node["name"].(string),
					Ename:     node["ename"].(string),
					ShowIndex: node["show_index"].(bool),
				}
			}
		}
	}
	return nodes
}

// GetChildrenNode 获取某个父节点下最多 num 个子节点
func GetChildrenNode(parentId, num int) []interface{} {
	nids := make([]interface{}, 0, num)

	if len(AllRecommendNodes) > 0 {
		nodeList := DefaultNode.FindByParent(parentId, num)

		for _, node := range nodeList {
			nids = append(nids, node.Nid)
		}

		return nids
	}

	for _, node := range AllNode {
		if node["pid"].(int) == parentId {
			nids = append(nids, node["nid"])
			if len(nids) == num {
				break
			}
		}
	}

	return nids
}

// 将 node 组织成一定结构，方便前端展示
func GenNodes() []map[string][]map[string]interface{} {
	if len(AllRecommendNodes) > 0 {
		return AllRecommendNodes
	}

	sameParent := make(map[string][]map[string]interface{})
	allParentNodes := make([]string, 0, 8)
	for _, node := range AllNode {
		if node["pid"].(int) != 0 {
			if len(sameParent[node["parent"].(string)]) == 0 {
				sameParent[node["parent"].(string)] = []map[string]interface{}{node}
			} else {
				sameParent[node["parent"].(string)] = append(sameParent[node["parent"].(string)], node)
			}
		} else {
			allParentNodes = append(allParentNodes, node["name"].(string))
		}
	}
	nodes := make([]map[string][]map[string]interface{}, 0, len(allParentNodes))
	for _, parent := range allParentNodes {
		tmpMap := make(map[string][]map[string]interface{})
		tmpMap[parent] = sameParent[parent]
		nodes = append(nodes, tmpMap)
	}
	logger.Debugf("%v\n", nodes)
	return nodes
}

// 将所有 资源分类信息 加载到内存中：后台修改节点时，重新加载一次
func LoadCategories() (err error) {
	ctx := context.Background()
	categories := make([]*model.ResourceCat, 0)
	cursor, err := db.GetCollection("resource_category").Find(ctx, bson.M{})
	if err != nil {
		logger.Errorln("LoadCategories category read fail:", err)
		return
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &categories); err != nil {
		logger.Errorln("LoadCategories category decode fail:", err)
		return
	}

	catRWMutex.Lock()
	defer catRWMutex.Unlock()

	AllCategory = categories

	logger.Infoln("LoadCategories successfully!")

	return
}

// 获得分类名
func GetCategoryName(catid int) string {
	catRWMutex.RLock()
	defer catRWMutex.RUnlock()
	for _, cat := range AllCategory {
		if cat.Catid == catid {
			return cat.Name
		}
	}
	return ""
}

func GetCurIndexNav(tab string) *model.IndexNav {
	for _, indexNav := range WebsiteSetting.IndexNavs {
		if indexNav.Tab == tab {
			return indexNav
		}
	}
	return nil
}

func loadRecommendNodes() bool {
	ctx := context.Background()

	recommendNodes := make([]*model.RecommendNode, 0)
	cursor, err := db.GetCollection("recommend_node").Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"seq": 1}))
	if err != nil {
		logger.Errorln("loadRecommendNodes recommend_node read fail:", err)
		return false
	}
	if err = cursor.All(ctx, &recommendNodes); err != nil {
		cursor.Close(ctx)
		logger.Errorln("loadRecommendNodes recommend_node decode fail:", err)
		return false
	}
	cursor.Close(ctx)

	if len(recommendNodes) == 0 {
		return false
	}

	nids := make([]int, 0, len(recommendNodes))
	for _, rn := range recommendNodes {
		if rn.Nid > 0 {
			nids = append(nids, rn.Nid)
		}
	}

	topicNodeMap := make(map[int]*model.TopicNode)
	if len(nids) > 0 {
		topicCursor, terr := db.GetCollection("topics_node").Find(ctx, bson.M{"_id": bson.M{"$in": nids}})
		if terr == nil {
			topicNodes := make([]*model.TopicNode, 0)
			if terr = topicCursor.All(ctx, &topicNodes); terr == nil {
				for _, tn := range topicNodes {
					topicNodeMap[tn.Nid] = tn
				}
			}
			topicCursor.Close(ctx)
		}
	}

	nodeList := make([]*model.NodeInfo, len(recommendNodes))
	for i, rn := range recommendNodes {
		ni := &model.NodeInfo{RecommendNode: *rn}
		if tn, ok := topicNodeMap[rn.Nid]; ok {
			ni.TopicNode = *tn
		}
		nodeList[i] = ni
	}

	parentMap := make(map[int]string)
	parentSlice := make([]string, 0, 20)
	sameParent := make(map[string][]map[string]interface{})

	for _, node := range nodeList {
		if node.RecommendNode.Parent == 0 {
			parentName := node.RecommendNode.Name
			parentMap[node.Id] = parentName
			parentSlice = append(parentSlice, parentName)
		} else {
			parentName := parentMap[node.RecommendNode.Parent]
			sameParent[parentName] = append(sameParent[parentName], map[string]interface{}{
				"name":  node.TopicNode.Name,
				"ename": node.Ename,
			})
		}
	}

	AllRecommendNodes = make([]map[string][]map[string]interface{}, len(parentSlice))

	for i, name := range parentSlice {
		children := sameParent[name]
		AllRecommendNodes[i] = map[string][]map[string]interface{}{
			name: children,
		}
	}

	logger.Infoln("loadRecommendNodes successfully!")

	return true
}
