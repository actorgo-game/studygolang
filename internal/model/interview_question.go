// Copyright 2022 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import (
	"strconv"
	"time"
)

// Go 面试题
type InterviewQuestion struct {
	Id        int       `json:"id" bson:"_id"`
	Sn        int64     `json:"sn" bson:"sn"`
	ShowSn    string    `json:"show_sn" bson:"-"`
	Question  string    `json:"question" bson:"question"`
	Answer    string    `json:"answer" bson:"answer"`
	Level     int       `json:"level" bson:"level"`
	Viewnum   int       `json:"viewnum" bson:"viewnum"`
	Cmtnum    int       `json:"cmtnum" bson:"cmtnum"`
	Likenum   int       `json:"likenum" bson:"likenum"`
	Source    string    `json:"source" bson:"source"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

func (iq *InterviewQuestion) AfterLoad() {
	iq.ShowSn = strconv.FormatInt(iq.Sn, 32)
}

func (iq *InterviewQuestion) AfterInsert() {
	iq.ShowSn = strconv.FormatInt(iq.Sn, 32)
}
