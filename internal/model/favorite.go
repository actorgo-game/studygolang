// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

// 用户收藏（用户可以收藏文章、话题、资源等）
type Favorite struct {
	Uid     int    `json:"uid" bson:"uid"`
	Objtype int    `json:"objtype" bson:"objtype"`
	Objid   int    `json:"objid" bson:"objid"`
	Ctime   string `json:"ctime" bson:"ctime"`
}

func (*Favorite) CollectionName() string {
	return "favorites"
}
