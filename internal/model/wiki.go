// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

type Wiki struct {
	Id      int       `json:"id" bson:"_id"`
	Title   string    `json:"title" bson:"title"`
	Content string    `json:"content" bson:"content"`
	Uri     string    `json:"uri" bson:"uri"`
	Uid     int       `json:"uid" bson:"uid"`
	Cuid    string    `json:"cuid" bson:"cuid"`
	Viewnum int       `json:"viewnum" bson:"viewnum"`
	Tags    string    `json:"tags" bson:"tags"`
	Ctime   OftenTime `json:"ctime" bson:"ctime"`
	Mtime   time.Time `json:"mtime" bson:"mtime"`

	Users map[int]*User `bson:"-"`
}

func (this *Wiki) BeforeInsert() {
	if this.Tags == "" {
		this.Tags = AutoTag(this.Title, this.Content, 4)
	}
}
