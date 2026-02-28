// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import (
	"strings"
	"time"
)

const (
	RtypeGo   = iota // Go技术晨读
	RtypeComp        // 综合技术晨读
)

// 技术晨读
type MorningReading struct {
	Id       int       `json:"id" bson:"_id"`
	Content  string    `json:"content" bson:"content"`
	Rtype    int       `json:"rtype" bson:"rtype"`
	Inner    int       `json:"inner" bson:"inner"`
	Url      string    `json:"url" bson:"url"`
	Moreurls string    `json:"moreurls" bson:"moreurls"`
	Username string    `json:"username" bson:"username"`
	Clicknum int       `json:"clicknum,omitempty" bson:"clicknum"`
	Ctime    OftenTime `json:"ctime" bson:"ctime"`

	// 晨读日期，从 ctime 中提取
	Rdate string `json:"rdate,omitempty" bson:"-"`

	Urls []string `json:"urls" bson:"-"`
}

func (this *MorningReading) AfterLoad() {
	this.Rdate = time.Time(this.Ctime).Format("2006-01-02")
	if this.Moreurls != "" {
		this.Urls = strings.Split(this.Moreurls, ",")
	}
}
