// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

// 权限信息
type Authority struct {
	Aid    int       `json:"aid" bson:"_id"`
	Name   string    `json:"name" bson:"name"`
	Menu1  int       `json:"menu1" bson:"menu1"`
	Menu2  int       `json:"menu2" bson:"menu2"`
	Route  string    `json:"route" bson:"route"`
	OpUser string    `json:"op_user" bson:"op_user"`
	Ctime  OftenTime `json:"ctime" bson:"ctime"`
	Mtime  OftenTime `json:"mtime" bson:"mtime"`
}
