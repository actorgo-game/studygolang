// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

type ViewSource struct {
	Id        int       `bson:"_id"`
	Objid     int       `bson:"objid"`
	Objtype   int       `bson:"objtype"`
	Google    int       `bson:"google"`
	Baidu     int       `bson:"baidu"`
	Bing      int       `bson:"bing"`
	Sogou     int       `bson:"sogou"`
	So        int       `bson:"so"`
	Other     int       `bson:"other"`
	UpdatedAt OftenTime `bson:"updated_at"`
}
