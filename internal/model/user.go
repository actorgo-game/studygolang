// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/polaris1119/goutils"
)

// 用户登录信息
type UserLogin struct {
	Uid       int       `json:"uid" bson:"_id"`
	Username  string    `json:"username" bson:"username"`
	Passcode  string    `json:"passcode" bson:"passcode"`
	Passwd    string    `json:"passwd" bson:"passwd"`
	Email     string    `json:"email" bson:"email"`
	LoginIp   string    `json:"login_ip" bson:"login_ip"`
	LoginTime time.Time `json:"login_time" bson:"login_time"`
}

func (this *UserLogin) CollectionName() string {
	return "user_login"
}

// 生成加密密码
func (this *UserLogin) GenMd5Passwd() error {
	if this.Passwd == "" {
		return errors.New("password is empty!")
	}
	this.Passcode = fmt.Sprintf("%x", rand.Int31())
	// 密码经过md5(passwd+passcode)加密保存
	this.Passwd = goutils.Md5(this.Passwd + this.Passcode)
	return nil
}

const (
	UserStatusNoAudit = iota
	UserStatusAudit   // 已激活
	UserStatusRefuse
	UserStatusFreeze // 冻结
	UserStatusOutage // 停用
)

const (
	// 用户拥有的权限设置
	DauAuthTopic = 1 << iota
	DauAuthArticle
	DauAuthResource
	DauAuthWiki
	DauAuthProject
	DauAuthBook
	DauAuthComment // 评论
	DauAuthTop     // 置顶
)

const DefaultAuth = DauAuthTopic | DauAuthArticle | DauAuthResource | DauAuthProject | DauAuthComment

// 用户基本信息
type User struct {
	Uid         int       `json:"uid" bson:"_id"`
	Username    string    `json:"username" bson:"username" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	Email       string    `json:"email" bson:"email"`
	Open        int       `json:"open" bson:"open"`
	Name        string    `json:"name" bson:"name"`
	Avatar      string    `json:"avatar" bson:"avatar"`
	City        string    `json:"city" bson:"city"`
	Company     string    `json:"company" bson:"company"`
	Github      string    `json:"github" bson:"github"`
	Gitea       string    `json:"gitea" bson:"gitea"`
	Weibo       string    `json:"weibo" bson:"weibo"`
	Website     string    `json:"website" bson:"website"`
	Monlog      string    `json:"monlog" bson:"monlog"`
	Introduce   string    `json:"introduce" bson:"introduce"`
	Unsubscribe int       `json:"unsubscribe" bson:"unsubscribe"`
	Balance     int       `json:"balance" bson:"balance"`
	IsThird     int       `json:"is_third" bson:"is_third"`
	DauAuth     int       `json:"dau_auth" bson:"dau_auth"`
	IsVip       bool      `json:"is_vip" bson:"is_vip"`
	VipExpire   int       `json:"vip_expire" bson:"vip_expire"`
	Status      int       `json:"status" bson:"status"`
	IsRoot      bool      `json:"is_root" bson:"is_root"`
	Ctime       OftenTime `json:"ctime" bson:"ctime"`
	Mtime       time.Time `json:"mtime" bson:"mtime"`

	// 非用户表中的信息，为了方便放在这里
	Roleids   []int    `bson:"-"`
	Rolenames []string `bson:"-"`

	// 活跃度
	Weight int `json:"weight" bson:"-"`
	Gold   int `json:"gold" bson:"-"`
	Silver int `json:"silver" bson:"-"`
	Copper int `json:"copper" bson:"-"`

	IsOnline bool `json:"is_online" bson:"-"`
}

func (this *User) CollectionName() string {
	return "user_info"
}

func (this *User) String() string {
	buffer := goutils.NewBuffer()
	buffer.Append(this.Username).Append(" ").
		Append(this.Email).Append(" ").
		Append(this.Uid).Append(" ").
		Append(this.Mtime)

	return buffer.String()
}

func (this *User) AfterLoad() {
	this.Gold = this.Balance / 10000
	balance := this.Balance % 10000

	this.Silver = balance / 100
	this.Copper = balance % 100
}

// Me 代表当前用户
type Me struct {
	Uid       int       `json:"uid" bson:"uid"`
	Username  string    `json:"username" bson:"username"`
	Name      string    `json:"name" bson:"name"`
	Monlog    string    `json:"monlog" bson:"monlog"`
	Email     string    `json:"email" bson:"email"`
	Avatar    string    `json:"avatar" bson:"avatar"`
	Status    int       `json:"status" bson:"status"`
	MsgNum    int       `json:"msgnum" bson:"msgnum"`
	IsAdmin   bool      `json:"isadmin" bson:"isadmin"`
	IsRoot    bool      `json:"is_root" bson:"is_root"`
	DauAuth   int       `json:"dau_auth" bson:"dau_auth"`
	IsVip     bool      `json:"is_vip" bson:"is_vip"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`

	Balance int `json:"balance" bson:"balance"`
	Gold    int `json:"gold" bson:"gold"`
	Silver  int `json:"silver" bson:"silver"`
	Copper  int `json:"copper" bson:"copper"`

	RoleIds []int `json:"-" bson:"-"`
}

// 活跃用户信息
// 活跃度规则：
//
//	1、注册成功后 +2
//	2、登录一次 +1
//	3、修改资料 +1
//	4、发帖子 + 10
//	5、评论 +5
//	6、创建Wiki页 +10
type UserActive struct {
	Uid      int       `json:"uid" bson:"_id"`
	Username string    `json:"username" bson:"username"`
	Email    string    `json:"email" bson:"email"`
	Avatar   string    `json:"avatar" bson:"avatar"`
	Weight   int       `json:"weight" bson:"weight"`
	Mtime    time.Time `json:"mtime" bson:"mtime"`
}

// 用户角色信息
type UserRole struct {
	Uid    int    `json:"uid" bson:"uid"`
	Roleid int    `json:"roleid" bson:"roleid"`
	ctime  string `bson:"-"`
}

const (
	BindTypeGithub = iota
	BindTypeGitea
)

type BindUser struct {
	Id           int       `json:"id" bson:"_id"`
	Uid          int       `json:"uid" bson:"uid"`
	Type         int       `json:"type" bson:"type"`
	Email        string    `json:"email" bson:"email"`
	Tuid         int       `json:"tuid" bson:"tuid"`
	Username     string    `json:"username" bson:"username"`
	Name         string    `json:"name" bson:"name"`
	AccessToken  string    `json:"access_token" bson:"access_token"`
	RefreshToken string    `json:"refresh_token" bson:"refresh_token"`
	Expire       int       `json:"expire" bson:"expire"`
	Avatar       string    `json:"avatar" bson:"avatar"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}
