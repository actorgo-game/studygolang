// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

type ViewRecord struct {
	Id        int       `json:"id" bson:"_id"`
	Objid     int       `json:"objid" bson:"objid"`
	Objtype   int       `json:"objtype" bson:"objtype"`
	Uid       int       `json:"uid" bson:"uid"`
	CreatedAt OftenTime `json:"created_at" bson:"created_at"`
}
