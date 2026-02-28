// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

type LearningMaterial struct {
	Id        int       `json:"-" bson:"_id"`
	Title     string    `json:"title" bson:"title"`
	Url       string    `json:"url" bson:"url"`
	Type      int       `json:"type" bson:"type"`
	Seq       int       `json:"-" bson:"seq"`
	FirstUrl  string    `json:"first_url" bson:"first_url"`
	CreatedAt time.Time `json:"-" bson:"created_at"`
}
