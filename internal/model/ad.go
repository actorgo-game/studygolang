// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

type Advertisement struct {
	Id        int       `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	AdType    int       `json:"ad_type" bson:"ad_type"`
	Code      string    `json:"code" bson:"code"`
	Source    string    `json:"source" bson:"source"`
	IsOnline  bool      `json:"is_online" bson:"is_online"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type PageAd struct {
	Id        int       `json:"id" bson:"_id"`
	Path      string    `json:"path" bson:"path"`
	AdId      int       `json:"ad_id" bson:"ad_id"`
	Position  string    `json:"position" bson:"position"`
	IsOnline  bool      `json:"is_online" bson:"is_online"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
