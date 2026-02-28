// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

const (
	GCTTRoleTranslator = iota
	GCTTRoleLeader
	GCTTRoleSelecter // 选题
	GCTTRoleChecker  // 校对
	GCTTRoleCore     // 核心成员
)

const (
	IssueOpened = iota
	IssueClosed
)

const (
	LabelUnClaim = "待认领"
	LabelClaimed = "已认领"
)

var roleMap = map[int]string{
	GCTTRoleTranslator: "译者",
	GCTTRoleLeader:     "组长",
	GCTTRoleSelecter:   "选题",
	GCTTRoleChecker:    "校对",
	GCTTRoleCore:       "核心成员",
}

var faMap = map[int]string{
	GCTTRoleTranslator: "fa-user",
	GCTTRoleLeader:     "fa-graduation-cap",
	GCTTRoleSelecter:   "fa-user-circle",
	GCTTRoleChecker:    "fa-user-secret",
	GCTTRoleCore:       "fa-heart",
}

type GCTTUser struct {
	Id        int       `bson:"_id"`
	Username  string    `bson:"username"`
	Avatar    string    `bson:"avatar"`
	Uid       int       `bson:"uid"`
	JoinedAt  int64     `bson:"joined_at"`
	LastAt    int64     `bson:"last_at"`
	Num       int       `bson:"num"`
	Words     int       `bson:"words"`
	AvgTime   int       `bson:"avg_time"`
	Role      int       `bson:"role"`
	CreatedAt time.Time `bson:"created_at"`

	RoleName string `bson:"-"`
	Fa       string `bson:"-"`
}

func (this *GCTTUser) AfterLoad() {
	this.RoleName = roleMap[this.Role]
	this.Fa = faMap[this.Role]
}

func (*GCTTUser) CollectionName() string {
	return "gctt_user"
}

type GCTTGit struct {
	Id            int       `bson:"_id"`
	Username      string    `bson:"username"`
	Md5           string    `bson:"md5"`
	Title         string    `bson:"title"`
	PR            int       `bson:"pr"`
	TranslatingAt int64     `bson:"translating_at"`
	TranslatedAt  int64     `bson:"translated_at"`
	Words         int       `bson:"words"`
	ArticleId     int       `bson:"article_id"`
	CreatedAt     time.Time `bson:"created_at"`
}

func (*GCTTGit) CollectionName() string {
	return "gctt_git"
}

type GCTTIssue struct {
	Id            int       `bson:"_id"`
	Translator    string    `bson:"translator"`
	Email         string    `bson:"email"`
	Title         string    `bson:"title"`
	TranslatingAt int64     `bson:"translating_at"`
	TranslatedAt  int64     `bson:"translated_at"`
	Label         string    `bson:"label"`
	State         uint8     `bson:"state"`
	CreatedAt     time.Time `bson:"created_at"`
}

func (*GCTTIssue) CollectionName() string {
	return "gctt_issue"
}

type GCTTTimeLine struct {
	Id        int       `bson:"_id"`
	Content   string    `bson:"content"`
	CreatedAt time.Time `bson:"created_at"`
}

func (*GCTTTimeLine) CollectionName() string {
	return "gctt_timeline"
}
