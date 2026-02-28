// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

type FriendLink struct {
	Id        int       `json:"-" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	Url       string    `json:"url" bson:"url"`
	Logo      string    `json:"logo" bson:"logo"`
	Seq       int       `json:"-" bson:"seq"`
	CreatedAt time.Time `json:"-" bson:"created_at"`
}
