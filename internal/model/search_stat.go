// Copyright 2014 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

// 搜索词统计
type SearchStat struct {
	Id      int       `json:"id" bson:"_id"`
	Keyword string    `json:"keyword" bson:"keyword"`
	Times   int       `json:"times" bson:"times"`
	Ctime   time.Time `json:"ctime" bson:"ctime"`
}
