// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"net/url"
	"strconv"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/global"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/polaris1119/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuthorityLogic struct{}

var DefaultAuthority = AuthorityLogic{}

// GetUserMenu 获取用户菜单
func (self AuthorityLogic) GetUserMenu(ctx context.Context, user *model.Me, uri string) ([]*model.Authority, map[int][]*model.Authority, int) {
	var (
		aidMap = make(map[int]bool)
		err    error
	)

	if !user.IsRoot {
		aidMap, err = self.userAuthority(user)
		if err != nil {
			return nil, nil, 0
		}
	}

	authLocker.RLock()
	defer authLocker.RUnlock()

	userMenu1 := make([]*model.Authority, 0, 4)
	userMenu2 := make(map[int][]*model.Authority)
	curMenu1 := 0

	for _, authority := range Authorities {
		if _, ok := aidMap[authority.Aid]; ok || user.IsRoot {
			if authority.Menu1 == 0 {
				userMenu1 = append(userMenu1, authority)
				userMenu2[authority.Aid] = make([]*model.Authority, 0, 4)
			} else if authority.Menu2 == 0 {
				userMenu2[authority.Menu1] = append(userMenu2[authority.Menu1], authority)
			}
			if authority.Route == uri {
				curMenu1 = authority.Menu1
			}
		}
	}

	return userMenu1, userMenu2, curMenu1
}

// 获取整个菜单
func (AuthorityLogic) GetMenus() ([]*model.Authority, map[string][][]string) {
	var (
		menu1 = make([]*model.Authority, 0, 10)
		menu2 = make(map[string][][]string)
	)

	for _, authority := range Authorities {
		if authority.Menu1 == 0 {
			menu1 = append(menu1, authority)
			aid := strconv.Itoa(authority.Aid)
			menu2[aid] = make([][]string, 0, 4)
		} else if authority.Menu2 == 0 {
			m := strconv.Itoa(authority.Menu1)
			oneMenu2 := []string{strconv.Itoa(authority.Aid), authority.Name}
			menu2[m] = append(menu2[m], oneMenu2)
		}
	}

	return menu1, menu2
}

// 除了一级、二级菜单之外的权限（路由）
func (AuthorityLogic) GeneralAuthorities() map[int][]*model.Authority {
	auths := make(map[int][]*model.Authority)

	for _, authority := range Authorities {
		if authority.Menu1 == 0 {
			auths[authority.Aid] = make([]*model.Authority, 0, 8)
		} else if authority.Menu2 != 0 {
			auths[authority.Menu1] = append(auths[authority.Menu1], authority)
		}
	}

	return auths
}

// 判断用户是否有某个权限
func (self AuthorityLogic) HasAuthority(user *model.Me, route string) bool {
	if user.IsRoot {
		return true
	}

	aidMap, err := self.userAuthority(user)
	if err != nil {
		logger.Errorln("HasAuthority:Read user authority error:", err)
		return false
	}

	authLocker.RLock()
	defer authLocker.RUnlock()

	for _, authority := range Authorities {
		if _, ok := aidMap[authority.Aid]; ok {
			if route == authority.Route {
				return true
			}
		}
	}

	return false
}

func (AuthorityLogic) FindAuthoritiesByPage(ctx context.Context, conds map[string]string, curPage, limit int) ([]*model.Authority, int) {
	objLog := GetLogger(ctx)

	filter := bson.M{}
	for k, v := range conds {
		filter[k] = v
	}

	offset := int64((curPage - 1) * limit)

	auhtorities := make([]*model.Authority, 0)
	cursor, err := db.GetCollection("authority").Find(ctx, filter,
		options.Find().SetSkip(offset).SetLimit(int64(limit)))
	if err != nil {
		objLog.Errorln("find error:", err)
		return nil, 0
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &auhtorities); err != nil {
		objLog.Errorln("find decode error:", err)
		return nil, 0
	}

	total, err := db.GetCollection("authority").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("find count error:", err)
		return nil, 0
	}

	return auhtorities, int(total)
}

func (AuthorityLogic) FindById(ctx context.Context, aid int) *model.Authority {
	objLog := GetLogger(ctx)

	if aid == 0 {
		return nil
	}

	authority := &model.Authority{}
	err := db.GetCollection("authority").FindOne(ctx, bson.M{"_id": aid}).Decode(authority)
	if err != nil {
		objLog.Errorln("authority FindById error:", err)
		return nil
	}

	return authority
}

func (AuthorityLogic) Save(ctx context.Context, form url.Values, opUser string) (errMsg string, err error) {
	objLog := GetLogger(ctx)

	authority := &model.Authority{}
	err = schemaDecoder.Decode(authority, form)
	if err != nil {
		objLog.Errorln("authority schema Decoder error", err)
		errMsg = err.Error()
		return
	}

	authority.OpUser = opUser

	if authority.Aid != 0 {
		_, err = db.GetCollection("authority").ReplaceOne(ctx, bson.M{"_id": authority.Aid}, authority)
	} else {
		authority.Aid, err = db.NextID("authority")
		if err != nil {
			errMsg = "内部服务器错误"
			objLog.Errorln(errMsg, ":", err)
			return
		}
		_, err = db.GetCollection("authority").InsertOne(ctx, authority)
	}

	if err != nil {
		errMsg = "内部服务器错误"
		objLog.Errorln(errMsg, ":", err)
		return
	}

	global.AuthorityChan <- struct{}{}

	return
}

func (AuthorityLogic) Del(aid int) error {
	ctx := context.Background()
	_, err := db.GetCollection("authority").DeleteOne(ctx, bson.M{"_id": aid})

	global.AuthorityChan <- struct{}{}

	return err
}

func (AuthorityLogic) userAuthority(user *model.Me) (map[int]bool, error) {
	ctx := context.Background()
	userRoles := make([]*model.UserRole, 0)
	cursor, err := db.GetCollection("user_role").Find(ctx, bson.M{"uid": user.Uid})
	if err != nil {
		logger.Errorln("userAuthority userole read fail:", err)
		return nil, err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &userRoles); err != nil {
		logger.Errorln("userAuthority userole decode fail:", err)
		return nil, err
	}

	roleAuthLocker.RLock()

	aidMap := make(map[int]bool)
	for _, userRole := range userRoles {
		for _, aid := range RoleAuthorities[userRole.Roleid] {
			aidMap[aid] = true
		}
	}

	roleAuthLocker.RUnlock()

	return aidMap, nil
}
