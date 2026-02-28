// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

// 动态（go动态；本站动态等）
type Dynamic struct {
	Id      int       `json:"id" bson:"_id"`
	Content string    `json:"content" bson:"content"`
	Dmtype  int       `json:"dmtype" bson:"dmtype"`
	Url     string    `json:"url" bson:"url"`
	Seq     int       `json:"seq" bson:"seq"`
	Ctime   time.Time `json:"ctime" bson:"ctime"`
}
