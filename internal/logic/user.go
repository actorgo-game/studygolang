// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"errors"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"
	"github.com/studygolang/studygolang/util"

	"github.com/polaris1119/times"

	"github.com/polaris1119/slices"

	"github.com/go-validator/validator"
	"github.com/polaris1119/config"
	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserLogic struct{}

var DefaultUser = UserLogic{}

// CreateUser 创建用户
func (self UserLogic) CreateUser(ctx context.Context, form url.Values) (errMsg string, err error) {
	objLog := GetLogger(ctx)

	if self.UserExists(ctx, "email", form.Get("email")) {
		errMsg = "该邮箱已注册过"
		err = errors.New(errMsg)
		return
	}
	if self.UserExists(ctx, "username", form.Get("username")) {
		errMsg = "用户名已存在"
		err = errors.New(errMsg)
		return
	}

	user := &model.User{}
	err = schemaDecoder.Decode(user, form)
	if err != nil {
		objLog.Errorln("user schema Decode error:", err)
		errMsg = err.Error()
		return
	}

	if err = validator.Validate(user); err != nil {
		objLog.Errorf("validate user error:%#v", err)

		if errMap, ok := err.(validator.ErrorMap); ok {
			if _, ok = errMap["Username"]; ok {
				errMsg = "用户名不合法！"
			}
		} else {
			errMsg = err.Error()
		}
		return
	}

	if config.ConfigFile.MustBool("account", "verify_email", true) {
		if !user.IsRoot {
			user.Status = model.UserStatusNoAudit
		}
	} else {
		user.Status = model.UserStatusAudit
	}

	session, sessionErr := db.GetClient().StartSession()
	if sessionErr != nil {
		errMsg = "内部服务错误！"
		err = sessionErr
		objLog.Errorln("start session error:", err)
		return
	}
	defer session.EndSession(ctx)

	_, txErr := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		txErr := self.doCreateUser(sessCtx, user, form.Get("passwd"))
		if txErr != nil {
			return nil, txErr
		}

		if form.Get("id") != "" {
			id := goutils.MustInt(form.Get("id"))
			_, txErr = DefaultWechat.Bind(sessCtx, id, user.Uid, form.Get("userInfo"))
			if txErr != nil {
				return nil, txErr
			}
		}
		return nil, nil
	})

	if txErr != nil {
		errMsg = "内部服务错误！"
		err = txErr
		objLog.Errorln("create user error:", err)
		return
	}

	return
}

// Update 更新用户信息
func (self UserLogic) Update(ctx context.Context, me *model.Me, form url.Values) (errMsg string, err error) {
	objLog := GetLogger(ctx)

	if form.Get("open") != "1" {
		form.Set("open", "0")
	}

	user := &model.User{}
	err = schemaDecoder.Decode(user, form)
	if err != nil {
		objLog.Errorln("userlogic update, schema decode error:", err)
		errMsg = "服务内部错误"
		return
	}

	updateFields := bson.M{
		"name":      user.Name,
		"open":      user.Open,
		"city":      user.City,
		"company":   user.Company,
		"github":    user.Github,
		"weibo":     user.Weibo,
		"website":   user.Website,
		"monlog":    user.Monlog,
		"introduce": user.Introduce,
	}

	if user.Email != me.Email {
		updateFields["email"] = user.Email
		updateFields["status"] = model.UserStatusNoAudit
	}

	session, sessionErr := db.GetClient().StartSession()
	if sessionErr != nil {
		errMsg = "对不起，服务器内部错误，请稍后再试！"
		err = sessionErr
		return
	}
	defer session.EndSession(ctx)

	_, txErr := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		_, txErr := db.GetCollection("user_info").UpdateOne(sessCtx, bson.M{"_id": me.Uid}, bson.M{"$set": updateFields})
		if txErr != nil {
			return nil, txErr
		}

		_, txErr = db.GetCollection("user_login").UpdateOne(sessCtx, bson.M{"uid": me.Uid}, bson.M{"$set": bson.M{"email": me.Email}})
		if txErr != nil {
			return nil, txErr
		}
		return nil, nil
	})

	if txErr != nil {
		objLog.Errorf("更新用户 【%d】 信息失败：%s", me.Uid, txErr)
		if strings.Contains(txErr.Error(), "duplicate key") {
			errMsg = "该邮箱地址被其他账号注册了"
		} else {
			errMsg = "对不起，服务器内部错误，请稍后再试！"
		}
		err = txErr
		return
	}

	go self.IncrUserWeight("uid", me.Uid, 1)

	return
}

// UpdateUserStatus 更新用户状态
func (UserLogic) UpdateUserStatus(ctx context.Context, uid, status int) error {
	objLog := GetLogger(ctx)

	_, err := db.GetCollection("user_info").UpdateOne(ctx, bson.M{"_id": uid}, bson.M{"$set": bson.M{"status": status}})
	if err != nil {
		objLog.Errorf("更新用户 【%d】 状态失败：%s", uid, err)
	}

	return err
}

// ChangeAvatar 更换头像
func (UserLogic) ChangeAvatar(ctx context.Context, uid int, avatar string) (err error) {
	changeData := bson.M{"avatar": avatar}
	_, err = db.GetCollection("user_info").UpdateOne(ctx, bson.M{"_id": uid}, bson.M{"$set": changeData})
	if err == nil {
		_, err = db.GetCollection("user_active").UpdateOne(ctx, bson.M{"_id": uid}, bson.M{"$set": changeData})
	}

	return
}

// UserExists 判断用户是否存在
func (UserLogic) UserExists(ctx context.Context, field, val string) bool {
	objLog := GetLogger(ctx)

	userLogin := &model.UserLogin{}
	err := db.GetCollection("user_login").FindOne(ctx, bson.M{field: val}).Decode(userLogin)
	if err != nil || userLogin.Uid == 0 {
		if err != nil && err != mongo.ErrNoDocuments {
			objLog.Errorln("user logic UserExists error:", err)
		}
		return false
	}
	return true
}

// EmailOrUsernameExists 判断指定的邮箱（email）或用户名是否存在
func (UserLogic) EmailOrUsernameExists(ctx context.Context, email, username string) bool {
	objLog := GetLogger(ctx)

	userLogin := &model.UserLogin{}
	filter := bson.M{"$or": []bson.M{{"email": email}, {"username": username}}}
	err := db.GetCollection("user_login").FindOne(ctx, filter).Decode(userLogin)
	if err != nil || userLogin.Uid == 0 {
		if err != nil && err != mongo.ErrNoDocuments {
			objLog.Errorln("user logic EmailOrUsernameExists error:", err)
		}
		return false
	}
	return true
}

// FindUserInfos 获得用户信息，uniq 可能是 uid slice 或 username slice
func (self UserLogic) FindUserInfos(ctx context.Context, uniq interface{}) map[int]*model.User {
	objLog := GetLogger(ctx)

	field := "uid"
	if uids, ok := uniq.([]int); ok {
		if len(uids) == 0 {
			return nil
		}
		field = "_id"
	} else if usernames, ok := uniq.([]string); ok {
		if len(usernames) == 0 {
			return nil
		}
		field = "username"
	}

	filter := bson.M{field: bson.M{"$in": uniq}}
	cursor, err := db.GetCollection("user_info").Find(ctx, filter)
	if err != nil {
		objLog.Errorln("user logic FindUserInfos not record found:", err)
		return nil
	}
	defer cursor.Close(ctx)

	users := make([]*model.User, 0)
	if err = cursor.All(ctx, &users); err != nil {
		objLog.Errorln("user logic FindUserInfos decode error:", err)
		return nil
	}

	usersMap := make(map[int]*model.User, len(users))
	for _, u := range users {
		u.AfterLoad()
		usersMap[u.Uid] = u
	}
	return usersMap
}

func (self UserLogic) FindOne(ctx context.Context, field string, val interface{}) *model.User {
	objLog := GetLogger(ctx)

	user := &model.User{}
	err := db.GetCollection("user_info").FindOne(ctx, bson.M{field: val}).Decode(user)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			objLog.Errorln("user logic FindOne error:", err)
		}
	}

	if user.Uid != 0 {
		user.AfterLoad()

		if user.IsRoot {
			user.Roleids = []int{0}
			user.Rolenames = []string{"站长"}
			return user
		}

		userRoleList := make([]*model.UserRole, 0)
		cursor, err := db.GetCollection("user_role").Find(ctx, bson.M{"uid": user.Uid}, options.Find().SetSort(bson.M{"roleid": 1}))
		if err != nil {
			objLog.Errorf("获取用户 %s 角色 信息失败：%s", val, err)
			return nil
		}
		defer cursor.Close(ctx)
		if err = cursor.All(ctx, &userRoleList); err != nil {
			objLog.Errorf("获取用户 %s 角色 信息解码失败：%s", val, err)
			return nil
		}

		if roleNum := len(userRoleList); roleNum > 0 {
			user.Roleids = make([]int, roleNum)
			user.Rolenames = make([]string, roleNum)

			for i, userRole := range userRoleList {
				user.Roleids[i] = userRole.Roleid
				user.Rolenames[i] = Roles[userRole.Roleid-1].Name
			}
		}
	}
	return user
}

// 获取当前登录用户信息（常用信息）
func (self UserLogic) FindCurrentUser(ctx context.Context, username interface{}) *model.Me {
	objLog := GetLogger(ctx)

	user := &model.User{}
	var filter bson.M

	if uid, ok := username.(int); ok {
		filter = bson.M{"_id": uid, "status": bson.M{"$lte": model.UserStatusAudit}}
	} else {
		filter = bson.M{"username": username, "status": bson.M{"$lte": model.UserStatusAudit}}
	}

	err := db.GetCollection("user_info").FindOne(ctx, filter).Decode(user)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			objLog.Errorf("获取用户 %q 信息失败：%s", username, err)
		}
		return &model.Me{}
	}
	if user.Uid == 0 {
		logger.Infof("用户 %q 不存在或状态不正常！", username)
		return &model.Me{}
	}

	user.AfterLoad()

	isVip := user.IsVip
	if user.VipExpire < goutils.MustInt(times.Format("Ymd")) {
		isVip = false
	}

	me := &model.Me{
		Uid:       user.Uid,
		Username:  user.Username,
		Name:      user.Name,
		Monlog:    user.Monlog,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Status:    user.Status,
		IsRoot:    user.IsRoot,
		MsgNum:    DefaultMessage.FindNotReadMsgNum(ctx, user.Uid),
		DauAuth:   user.DauAuth,
		IsVip:     isVip,
		CreatedAt: time.Time(user.Ctime),

		Balance: user.Balance,
		Gold:    user.Gold,
		Silver:  user.Silver,
		Copper:  user.Copper,

		RoleIds: make([]int, 0, 2),
	}

	ip := ctx.Value("ip")
	go self.RecordLogin(user.Username, ip)

	if user.IsRoot {
		me.IsAdmin = true
		return me
	}

	userRoleList := make([]*model.UserRole, 0)
	cursor, err := db.GetCollection("user_role").Find(ctx, bson.M{"uid": user.Uid})
	if err != nil {
		logger.Errorf("获取用户 %q 角色 信息失败：%s", username, err)
		return me
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &userRoleList); err != nil {
		logger.Errorf("获取用户 %q 角色 信息解码失败：%s", username, err)
		return me
	}

	for _, userRole := range userRoleList {
		me.RoleIds = append(me.RoleIds, userRole.Roleid)

		if userRole.Roleid <= model.AdminMinRoleId {
			me.IsAdmin = true
		}
	}

	return me
}

// findUsers 获得用户信息，包内使用。
func (self UserLogic) findUsers(ctx context.Context, s interface{}) []*model.User {
	objLog := GetLogger(ctx)

	uids := slices.StructsIntSlice(s, "Uid")

	filter := bson.M{"_id": bson.M{"$in": uids}}
	cursor, err := db.GetCollection("user_info").Find(ctx, filter)
	if err != nil {
		objLog.Errorln("user logic findUsers not record found:", err)
		return nil
	}
	defer cursor.Close(ctx)

	users := make([]*model.User, 0)
	if err = cursor.All(ctx, &users); err != nil {
		objLog.Errorln("user logic findUsers decode error:", err)
		return nil
	}

	for _, u := range users {
		u.AfterLoad()
	}

	return users
}

func (self UserLogic) findUser(ctx context.Context, uid int) *model.User {
	objLog := GetLogger(ctx)

	user := &model.User{}
	err := db.GetCollection("user_info").FindOne(ctx, bson.M{"_id": uid}).Decode(user)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			objLog.Errorln("user logic findUser not record found:", err)
		}
	}

	user.AfterLoad()

	return user
}

// 会员总数
func (UserLogic) Total() int64 {
	total, err := db.GetCollection("user_info").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		logger.Errorln("UserLogic Total error:", err)
	}
	return total
}

func (UserLogic) IsAdmin(user *model.User) bool {
	if user.IsRoot {
		return true
	}

	for _, roleId := range user.Roleids {
		if roleId <= model.AdminMinRoleId {
			return true
		}
	}

	return false
}

var (
	ErrUsername = errors.New("用户名不存在")
	ErrPasswd  = errors.New("密码错误")
)

// Login 登录；成功返回用户登录信息(user_login)
func (self UserLogic) Login(ctx context.Context, username, passwd string) (*model.UserLogin, error) {
	objLog := GetLogger(ctx)

	userLogin := &model.UserLogin{}
	filter := bson.M{"$or": []bson.M{{"username": username}, {"email": username}}}
	err := db.GetCollection("user_login").FindOne(ctx, filter).Decode(userLogin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			objLog.Infof("user %q is not exists!", username)
			return nil, ErrUsername
		}
		objLog.Errorf("user %q login failure: %s", username, err)
		return nil, errors.New("内部错误，请稍后再试！")
	}
	if userLogin.Uid == 0 {
		objLog.Infof("user %q is not exists!", username)
		return nil, ErrUsername
	}

	user := &model.User{}
	err = db.GetCollection("user_info").FindOne(ctx, bson.M{"_id": userLogin.Uid}).Decode(user)
	if err != nil && err != mongo.ErrNoDocuments {
		objLog.Errorf("user %q login get user info failure: %s", username, err)
		return nil, errors.New("内部错误，请稍后再试！")
	}
	if user.Status > model.UserStatusAudit {
		objLog.Infof("用户 %q 的状态非审核通过, 用户的状态值：%d", username, user.Status)
		var errMap = map[int]error{
			model.UserStatusRefuse: errors.New("您的账号审核拒绝"),
			model.UserStatusFreeze: errors.New("您的账号因为非法发布信息已被冻结，请联系管理员！"),
			model.UserStatusOutage: errors.New("您的账号因为非法发布信息已被停号，请联系管理员！"),
		}
		return nil, errMap[user.Status]
	}

	md5Passwd := goutils.Md5(passwd + userLogin.Passcode)
	if md5Passwd != userLogin.Passwd {
		objLog.Infof("用户名 %q 填写的密码错误", username)
		return nil, ErrPasswd
	}

	go func() {
		self.IncrUserWeight("uid", userLogin.Uid, 1)
		ip := ctx.Value("ip")
		self.RecordLogin(username, ip)
	}()

	return userLogin, nil
}

// UpdatePasswd 更新用户密码
func (self UserLogic) UpdatePasswd(ctx context.Context, username, curPasswd, newPasswd string) (string, error) {
	userLogin := &model.UserLogin{}
	err := db.GetCollection("user_login").FindOne(ctx, bson.M{"username": username}).Decode(userLogin)
	if err != nil {
		return "用户不存在", err
	}

	if userLogin.Passwd != "" {
		_, err = self.Login(ctx, username, curPasswd)
		if err != nil {
			return "原密码填写错误", err
		}
	}

	userLogin = &model.UserLogin{
		Passwd: newPasswd,
	}
	err = userLogin.GenMd5Passwd()
	if err != nil {
		return err.Error(), err
	}

	changeData := bson.M{
		"passwd":   userLogin.Passwd,
		"passcode": userLogin.Passcode,
	}
	_, err = db.GetCollection("user_login").UpdateOne(ctx, bson.M{"username": username}, bson.M{"$set": changeData})
	if err != nil {
		logger.Errorf("用户 %s 更新密码错误：%s", username, err)
		return "对不起，内部服务错误！", err
	}
	return "", nil
}

func (UserLogic) HasPasswd(ctx context.Context, uid int) bool {
	userLogin := &model.UserLogin{}
	err := db.GetCollection("user_login").FindOne(ctx, bson.M{"uid": uid}).Decode(userLogin)
	if err == nil && userLogin.Passwd != "" {
		return true
	}

	return false
}

func (self UserLogic) ResetPasswd(ctx context.Context, email, passwd string) (string, error) {
	objLog := GetLogger(ctx)

	userLogin := &model.UserLogin{
		Passwd: passwd,
	}
	err := userLogin.GenMd5Passwd()
	if err != nil {
		return err.Error(), err
	}

	changeData := bson.M{
		"passwd":   userLogin.Passwd,
		"passcode": userLogin.Passcode,
	}
	_, err = db.GetCollection("user_login").UpdateOne(ctx, bson.M{"email": email}, bson.M{"$set": changeData})
	if err != nil {
		objLog.Errorf("用户 %s 更新密码错误：%s", email, err)
		return "对不起，内部服务错误！", err
	}
	return "", nil
}

// Activate 用户激活
func (self UserLogic) Activate(ctx context.Context, email, uuid string, timestamp int64, sign string) (*model.User, error) {
	objLog := GetLogger(ctx)

	realSign := DefaultEmail.genActivateSign(email, uuid, timestamp)
	if sign != realSign {
		return nil, errors.New("签名非法！")
	}

	user := self.FindOne(ctx, "email", email)
	if user.Uid == 0 {
		return nil, errors.New("邮箱非法")
	}

	user.Status = model.UserStatusAudit

	_, err := db.GetCollection("user_info").UpdateOne(ctx, bson.M{"_id": user.Uid}, bson.M{"$set": bson.M{"status": user.Status}})
	if err != nil {
		objLog.Errorf("activate [%s] failure:%s", email, err)
		return nil, err
	}

	return user, nil
}

// IncrUserWeight 增加或减少用户活跃度
func (UserLogic) IncrUserWeight(field string, value interface{}, weight int) {
	mongoField := field
	if field == "uid" {
		mongoField = "_id"
	}
	_, err := db.GetCollection("user_active").UpdateOne(
		context.Background(),
		bson.M{mongoField: value},
		bson.M{"$inc": bson.M{"weight": weight}},
	)
	if err != nil {
		logger.Errorln("UserActive update Error:", err)
	}
}

func (UserLogic) DecrUserWeight(field string, value interface{}, divide int) {
	if divide <= 0 {
		return
	}

	mongoField := field
	if field == "uid" {
		mongoField = "_id"
	}

	ctx := context.Background()
	coll := db.GetCollection("user_active")

	userActive := &model.UserActive{}
	err := coll.FindOne(ctx, bson.M{mongoField: value}).Decode(userActive)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Errorln("UserActive find Error:", err)
		}
		return
	}

	newWeight := userActive.Weight / divide
	result, err := coll.UpdateOne(ctx, bson.M{mongoField: value}, bson.M{"$set": bson.M{"weight": newWeight}})
	if err != nil {
		logger.Errorln("UserActive update Error:", err)
	} else {
		logger.Debugln("DecrUserWeight affected num:", result.ModifiedCount)
	}
}

// RecordLogin 记录用户最后登录时间和 IP
func (UserLogic) RecordLogin(username string, ipinter interface{}) error {
	change := bson.M{
		"login_time": time.Now(),
	}
	if ip, ok := ipinter.(string); ok && ip != "" {
		change["login_ip"] = ip
	}
	_, err := db.GetCollection("user_login").UpdateOne(
		context.Background(),
		bson.M{"username": username},
		bson.M{"$set": change},
	)
	if err != nil {
		logger.Errorf("记录用户 %q 登录错误：%s", username, err)
	}
	return err
}

// FindActiveUsers 获得活跃用户
func (UserLogic) FindActiveUsers(ctx context.Context, limit int, offset ...int) []*model.UserActive {
	objLog := GetLogger(ctx)

	findOpts := options.Find().SetSort(bson.M{"weight": -1}).SetLimit(int64(limit))
	if len(offset) > 0 && offset[0] > 0 {
		findOpts.SetSkip(int64(offset[0]))
	}

	cursor, err := db.GetCollection("user_active").Find(ctx, bson.M{}, findOpts)
	if err != nil {
		objLog.Errorln("UserLogic FindActiveUsers error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	activeUsers := make([]*model.UserActive, 0)
	if err = cursor.All(ctx, &activeUsers); err != nil {
		objLog.Errorln("UserLogic FindActiveUsers decode error:", err)
		return nil
	}
	return activeUsers
}

func (UserLogic) FindDAUUsers(ctx context.Context, uids []int) map[int]*model.User {
	objLog := GetLogger(ctx)

	filter := bson.M{"_id": bson.M{"$in": uids}}
	cursor, err := db.GetCollection("user_info").Find(ctx, filter)
	if err != nil {
		objLog.Errorln("UserLogic FindDAUUsers error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	userList := make([]*model.User, 0)
	if err = cursor.All(ctx, &userList); err != nil {
		objLog.Errorln("UserLogic FindDAUUsers decode error:", err)
		return nil
	}

	users := make(map[int]*model.User, len(userList))
	for _, u := range userList {
		u.AfterLoad()
		users[u.Uid] = u
	}
	return users
}

// FindNewUsers 最新加入会员
func (UserLogic) FindNewUsers(ctx context.Context, limit int, offset ...int) []*model.User {
	objLog := GetLogger(ctx)

	findOpts := options.Find().SetSort(bson.M{"ctime": -1}).SetLimit(int64(limit))
	if len(offset) > 0 && offset[0] > 0 {
		findOpts.SetSkip(int64(offset[0]))
	}

	cursor, err := db.GetCollection("user_info").Find(ctx, bson.M{}, findOpts)
	if err != nil {
		objLog.Errorln("UserLogic FindNewUsers error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	users := make([]*model.User, 0)
	if err = cursor.All(ctx, &users); err != nil {
		objLog.Errorln("UserLogic FindNewUsers decode error:", err)
		return nil
	}

	for _, u := range users {
		u.AfterLoad()
	}
	return users
}

// FindUserByPage 获取用户列表（分页）：后台用
func (UserLogic) FindUserByPage(ctx context.Context, conds map[string]string, curPage, limit int) ([]*model.User, int) {
	objLog := GetLogger(ctx)

	filter := bson.M{}
	for k, v := range conds {
		filter[k] = v
	}

	total, err := db.GetCollection("user_info").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("UserLogic find count error:", err)
		return nil, 0
	}

	offset := (curPage - 1) * limit
	findOpts := options.Find().
		SetSort(bson.M{"_id": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := db.GetCollection("user_info").Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("UserLogic find error:", err)
		return nil, 0
	}
	defer cursor.Close(ctx)

	userList := make([]*model.User, 0)
	if err = cursor.All(ctx, &userList); err != nil {
		objLog.Errorln("UserLogic find decode error:", err)
		return nil, 0
	}

	for _, u := range userList {
		u.AfterLoad()
	}

	return userList, int(total)
}

func (self UserLogic) AdminUpdateUser(ctx context.Context, uid string, form url.Values) {
	user := self.FindOne(ctx, "uid", uid)
	user.DauAuth = 0

	for k := range form {
		switch k {
		case "topic":
			user.DauAuth |= model.DauAuthTopic
		case "article":
			user.DauAuth |= model.DauAuthArticle
		case "resource":
			user.DauAuth |= model.DauAuthResource
		case "project":
			user.DauAuth |= model.DauAuthProject
		case "wiki":
			user.DauAuth |= model.DauAuthWiki
		case "book":
			user.DauAuth |= model.DauAuthBook
		case "comment":
			user.DauAuth |= model.DauAuthComment
		case "top":
			user.DauAuth |= model.DauAuthTop
		}
	}

	user.IsVip = goutils.MustBool(form.Get("is_vip"), false)
	user.VipExpire = goutils.MustInt(form.Get("vip_expire"))

	updateFields := bson.M{
		"dau_auth":   user.DauAuth,
		"is_vip":     user.IsVip,
		"vip_expire": user.VipExpire,
	}
	db.GetCollection("user_info").UpdateOne(ctx, bson.M{"_id": user.Uid}, bson.M{"$set": updateFields})
}

// GetUserMentions 获取 @ 的 suggest 列表
func (UserLogic) GetUserMentions(term string, limit int, isHttps bool) []map[string]string {
	ctx := context.Background()
	filter := bson.M{"username": bson.M{"$regex": term, "$options": "i"}}
	findOpts := options.Find().SetSort(bson.M{"mtime": -1}).SetLimit(int64(limit))

	cursor, err := db.GetCollection("user_active").Find(ctx, filter, findOpts)
	if err != nil {
		logger.Errorln("UserLogic GetUserMentions Error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	userActives := make([]*model.UserActive, 0)
	if err = cursor.All(ctx, &userActives); err != nil {
		logger.Errorln("UserLogic GetUserMentions decode Error:", err)
		return nil
	}

	users := make([]map[string]string, len(userActives))
	for i, userActive := range userActives {
		user := make(map[string]string, 2)
		user["username"] = userActive.Username
		user["avatar"] = util.Gravatar(userActive.Avatar, userActive.Email, 20, isHttps)
		users[i] = user
	}

	return users
}

// FindNotLoginUsers 获取 loginTime 之前没有登录的用户
func (UserLogic) FindNotLoginUsers(loginTime time.Time) (userList []*model.UserLogin, err error) {
	ctx := context.Background()
	userList = make([]*model.UserLogin, 0)

	filter := bson.M{"login_time": bson.M{"$lt": loginTime}}
	cursor, err := db.GetCollection("user_login").Find(ctx, filter)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &userList)
	return
}

// EmailSubscribe 邮件订阅或取消订阅
func (UserLogic) EmailSubscribe(ctx context.Context, uid, unsubscribe int) {
	_, err := db.GetCollection("user_info").UpdateOne(ctx, bson.M{"_id": uid}, bson.M{"$set": bson.M{"unsubscribe": unsubscribe}})
	if err != nil {
		logger.Errorln("user:", uid, "Email Subscribe Error:", err)
	}
}

func (UserLogic) FindBindUsers(ctx context.Context, uid int) []*model.BindUser {
	bindUsers := make([]*model.BindUser, 0)

	cursor, err := db.GetCollection("bind_user").Find(ctx, bson.M{"uid": uid})
	if err != nil {
		logger.Errorln("user:", uid, "FindBindUsers Error:", err)
		return bindUsers
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &bindUsers); err != nil {
		logger.Errorln("user:", uid, "FindBindUsers decode Error:", err)
	}
	return bindUsers
}

func (UserLogic) doCreateUser(sessCtx mongo.SessionContext, user *model.User, passwd ...string) error {

	if user.Avatar == "" && len(DefaultAvatars) > 0 {
		user.Avatar = DefaultAvatars[rand.Intn(len(DefaultAvatars))]
	}
	user.Open = 0

	user.DauAuth = model.DefaultAuth

	uid, err := db.NextID("user_info")
	if err != nil {
		return err
	}
	user.Uid = uid

	_, err = db.GetCollection("user_info").InsertOne(sessCtx, user)
	if err != nil {
		return err
	}

	userLogin := &model.UserLogin{
		Email:    user.Email,
		Username: user.Username,
		Uid:      user.Uid,
	}
	if len(passwd) > 0 {
		userLogin.Passwd = passwd[0]
		err = userLogin.GenMd5Passwd()
		if err != nil {
			return err
		}
	}

	_, err = db.GetCollection("user_login").InsertOne(sessCtx, userLogin)
	if err != nil {
		return err
	}

	if !user.IsRoot {
		userRole := &model.UserRole{}
		userRole.Roleid = Roles[len(Roles)-1].Roleid
		userRole.Uid = user.Uid
		_, err = db.GetCollection("user_role").InsertOne(sessCtx, userRole)
		if err != nil {
			return err
		}
	}

	userActive := &model.UserActive{
		Uid:      user.Uid,
		Username: user.Username,
		Avatar:   user.Avatar,
		Email:    user.Email,
		Weight:   2,
	}
	_, err = db.GetCollection("user_active").InsertOne(sessCtx, userActive)
	if err != nil {
		return err
	}

	return nil
}

func (UserLogic) DeleteUserContent(ctx context.Context, uid int) error {
	user := &model.User{}
	err := db.GetCollection("user_info").FindOne(ctx, bson.M{"_id": uid}).Decode(user)
	if err != nil || user.Username == "" {
		return err
	}

	feedResult, feedErr := db.GetCollection("feed").DeleteMany(ctx, bson.M{"uid": uid})
	topicResult, topicErr := db.GetCollection("topics").DeleteMany(ctx, bson.M{"uid": uid})
	resourceResult, resourceErr := db.GetCollection("resource").DeleteMany(ctx, bson.M{"uid": uid})
	articleResult, articleErr := db.GetCollection("articles").DeleteMany(ctx, bson.M{"author_txt": user.Username})

	if topicErr == nil && topicResult.DeletedCount > 0 {
		db.GetCollection("topics_ex").DeleteMany(ctx, bson.M{"uid": uid})
	}
	if resourceErr == nil && resourceResult.DeletedCount > 0 {
		db.GetCollection("resource_ex").DeleteMany(ctx, bson.M{"uid": uid})
	}

	_ = feedErr
	_ = feedResult
	_ = articleErr
	_ = articleResult

	return nil
}
