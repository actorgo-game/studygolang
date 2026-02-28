// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/polaris1119/logger"

	"github.com/polaris1119/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

var githubConf *oauth2.Config
var giteaConf *oauth2.Config

const GithubAPIBaseUrl = "https://api.github.com"
const GiteaAPIBaseUrl = "https://gitea.com/api/v1"

func init() {
	githubConf = &oauth2.Config{
		ClientID:     config.ConfigFile.MustValue("github", "client_id"),
		ClientSecret: config.ConfigFile.MustValue("github", "client_secret"),
		Scopes:       []string{"user:email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}

	giteaConf = &oauth2.Config{
		ClientID:     config.ConfigFile.MustValue("gitea", "client_id"),
		ClientSecret: config.ConfigFile.MustValue("gitea", "client_secret"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://gitea.com/login/oauth/authorize",
			TokenURL: "https://gitea.com/login/oauth/access_token",
		},
	}
}

type ThirdUserLogic struct{}

var DefaultThirdUser = ThirdUserLogic{}

func (ThirdUserLogic) GithubAuthCodeUrl(ctx context.Context, redirectURL string) string {
	githubConf.RedirectURL = redirectURL
	return githubConf.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

func (self ThirdUserLogic) LoginFromGithub(ctx context.Context, code string) (*model.User, error) {
	objLog := GetLogger(ctx)

	githubUser, token, err := self.githubTokenAndUser(ctx, code)
	if err != nil {
		objLog.Errorln("LoginFromGithub githubTokenAndUser error:", err)
		return nil, err
	}

	bindUser := &model.BindUser{}
	err = db.GetCollection("bind_user").FindOne(ctx, bson.M{"username": githubUser.Login, "type": model.BindTypeGithub}).Decode(bindUser)
	if err != nil && err != mongo.ErrNoDocuments {
		objLog.Errorln("LoginFromGithub Get BindUser error:", err)
		return nil, err
	}

	if bindUser.Uid > 0 {
		change := bson.M{
			"access_token":  token.AccessToken,
			"refresh_token": token.RefreshToken,
		}
		if !token.Expiry.IsZero() {
			change["expire"] = int(token.Expiry.Unix())
		}
		_, err = db.GetCollection("bind_user").UpdateOne(ctx, bson.M{"uid": bindUser.Uid}, bson.M{"$set": change})
		if err != nil {
			objLog.Errorln("LoginFromGithub update token error:", err)
			return nil, err
		}

		user := DefaultUser.FindOne(ctx, "uid", bindUser.Uid)
		return user, nil
	}

	exists := DefaultUser.EmailOrUsernameExists(ctx, githubUser.Email, githubUser.Login)
	if exists {
		objLog.Errorln("LoginFromGithub Github 对应的用户信息被占用")
		return nil, errors.New("Github 对应的用户信息被占用，可能你注册过本站，用户名密码登录试试！")
	}

	session, sessErr := db.GetClient().StartSession()
	if sessErr != nil {
		objLog.Errorln("LoginFromGithub StartSession error:", sessErr)
		return nil, sessErr
	}
	defer session.EndSession(ctx)

	if githubUser.Email == "" {
		githubUser.Email = githubUser.Login + "@github.com"
	}
	user := &model.User{
		Email:    githubUser.Email,
		Username: githubUser.Login,
		Name:     githubUser.Name,
		City:     githubUser.Location,
		Company:  githubUser.Company,
		Github:   githubUser.Login,
		Website:  githubUser.Blog,
		Avatar:   githubUser.AvatarUrl,
		IsThird:  1,
		Status:   model.UserStatusAudit,
	}

	var retUser *model.User
	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		txErr := DefaultUser.doCreateUser(sc, user)
		if txErr != nil {
			return nil, txErr
		}

		bindUser = &model.BindUser{
			Uid:          user.Uid,
			Type:         model.BindTypeGithub,
			Email:        user.Email,
			Tuid:         githubUser.Id,
			Username:     githubUser.Login,
			Name:         githubUser.Name,
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			Avatar:       githubUser.AvatarUrl,
		}
		if !token.Expiry.IsZero() {
			bindUser.Expire = int(token.Expiry.Unix())
		}
		var id int
		id, txErr = db.NextID("bind_user")
		if txErr != nil {
			return nil, txErr
		}
		bindUser.Id = id
		_, txErr = db.GetCollection("bind_user").InsertOne(sc, bindUser)
		if txErr != nil {
			return nil, txErr
		}

		retUser = user
		return nil, nil
	})

	if err != nil {
		objLog.Errorln("LoginFromGithub transaction error:", err)
		return nil, err
	}

	return retUser, nil
}

func (self ThirdUserLogic) BindGithub(ctx context.Context, code string, me *model.Me) error {
	objLog := GetLogger(ctx)

	githubUser, token, err := self.githubTokenAndUser(ctx, code)
	if err != nil {
		objLog.Errorln("LoginFromGithub githubTokenAndUser error:", err)
		return err
	}

	bindUser := &model.BindUser{}
	err = db.GetCollection("bind_user").FindOne(ctx, bson.M{"username": githubUser.Login, "type": model.BindTypeGithub}).Decode(bindUser)
	if err != nil && err != mongo.ErrNoDocuments {
		objLog.Errorln("LoginFromGithub Get BindUser error:", err)
		return err
	}

	if bindUser.Uid > 0 {
		bindUser.AccessToken = token.AccessToken
		bindUser.RefreshToken = token.RefreshToken
		if !token.Expiry.IsZero() {
			bindUser.Expire = int(token.Expiry.Unix())
		}
		_, err = db.GetCollection("bind_user").UpdateOne(ctx, bson.M{"uid": bindUser.Uid}, bson.M{"$set": bindUser})
		if err != nil {
			objLog.Errorln("LoginFromGithub update token error:", err)
			return err
		}

		return nil
	}

	bindUser = &model.BindUser{
		Uid:          me.Uid,
		Type:         model.BindTypeGithub,
		Email:        githubUser.Email,
		Tuid:         githubUser.Id,
		Username:     githubUser.Login,
		Name:         githubUser.Name,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Avatar:       githubUser.AvatarUrl,
	}
	if !token.Expiry.IsZero() {
		bindUser.Expire = int(token.Expiry.Unix())
	}
	id, idErr := db.NextID("bind_user")
	if idErr != nil {
		objLog.Errorln("LoginFromGithub NextID error:", idErr)
		return idErr
	}
	bindUser.Id = id
	_, err = db.GetCollection("bind_user").InsertOne(ctx, bindUser)
	if err != nil {
		objLog.Errorln("LoginFromGithub insert bindUser error:", err)
		return err
	}

	return nil
}

func (ThirdUserLogic) GiteaAuthCodeUrl(ctx context.Context, redirectURL string) string {
	giteaConf.RedirectURL = redirectURL
	return giteaConf.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

func (self ThirdUserLogic) LoginFromGitea(ctx context.Context, code string) (*model.User, error) {
	objLog := GetLogger(ctx)

	giteaUser, token, err := self.giteaTokenAndUser(ctx, code)
	if err != nil {
		objLog.Errorln("LoginFromGithub githubTokenAndUser error:", err)
		return nil, err
	}

	bindUser := &model.BindUser{}
	err = db.GetCollection("bind_user").FindOne(ctx, bson.M{"username": giteaUser.UserName, "type": model.BindTypeGitea}).Decode(bindUser)
	if err != nil && err != mongo.ErrNoDocuments {
		objLog.Errorln("LoginFromGithub Get BindUser error:", err)
		return nil, err
	}

	if bindUser.Uid > 0 {
		change := bson.M{
			"access_token":  token.AccessToken,
			"refresh_token": token.RefreshToken,
		}
		if !token.Expiry.IsZero() {
			change["expire"] = int(token.Expiry.Unix())
		}
		_, err = db.GetCollection("bind_user").UpdateOne(ctx, bson.M{"uid": bindUser.Uid}, bson.M{"$set": change})
		if err != nil {
			objLog.Errorln("LoginFromGithub update token error:", err)
			return nil, err
		}

		user := DefaultUser.FindOne(ctx, "uid", bindUser.Uid)
		return user, nil
	}

	exists := DefaultUser.EmailOrUsernameExists(ctx, giteaUser.Email, giteaUser.UserName)
	if exists {
		objLog.Errorln("LoginFromGitea Gitea 对应的用户信息被占用")
		return nil, errors.New("Gitea 对应的用户信息被占用，可能你注册过本站，用户名密码登录试试！")
	}

	session, sessErr := db.GetClient().StartSession()
	if sessErr != nil {
		objLog.Errorln("LoginFromGitea StartSession error:", sessErr)
		return nil, sessErr
	}
	defer session.EndSession(ctx)

	if giteaUser.Email == "" {
		giteaUser.Email = giteaUser.UserName + "@gitea.com"
	}
	user := &model.User{
		Email:    giteaUser.Email,
		Username: giteaUser.UserName,
		Name:     model.DisplayName(giteaUser),
		City:     "",
		Company:  "",
		Gitea:    giteaUser.UserName,
		Website:  "",
		Avatar:   giteaUser.AvatarURL,
		IsThird:  1,
		Status:   model.UserStatusAudit,
	}

	var retUser *model.User
	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		txErr := DefaultUser.doCreateUser(sc, user)
		if txErr != nil {
			return nil, txErr
		}

		bindUser = &model.BindUser{
			Uid:          user.Uid,
			Type:         model.BindTypeGithub,
			Email:        user.Email,
			Tuid:         int(giteaUser.ID),
			Username:     giteaUser.UserName,
			Name:         model.DisplayName(giteaUser),
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			Avatar:       giteaUser.AvatarURL,
		}
		if !token.Expiry.IsZero() {
			bindUser.Expire = int(token.Expiry.Unix())
		}
		var id int
		id, txErr = db.NextID("bind_user")
		if txErr != nil {
			return nil, txErr
		}
		bindUser.Id = id
		_, txErr = db.GetCollection("bind_user").InsertOne(sc, bindUser)
		if txErr != nil {
			return nil, txErr
		}

		retUser = user
		return nil, nil
	})

	if err != nil {
		objLog.Errorln("LoginFromGitea transaction error:", err)
		return nil, err
	}

	return retUser, nil
}

func (self ThirdUserLogic) BindGitea(ctx context.Context, code string, me *model.Me) error {
	objLog := GetLogger(ctx)

	giteaUser, token, err := self.giteaTokenAndUser(ctx, code)
	if err != nil {
		objLog.Errorln("LoginFromGitea githubTokenAndUser error:", err)
		return err
	}

	bindUser := &model.BindUser{}
	err = db.GetCollection("bind_user").FindOne(ctx, bson.M{"username": giteaUser.UserName, "type": model.BindTypeGitea}).Decode(bindUser)
	if err != nil && err != mongo.ErrNoDocuments {
		objLog.Errorln("LoginFromGitea Get BindUser error:", err)
		return err
	}

	if bindUser.Uid > 0 {
		bindUser.AccessToken = token.AccessToken
		bindUser.RefreshToken = token.RefreshToken
		if !token.Expiry.IsZero() {
			bindUser.Expire = int(token.Expiry.Unix())
		}
		_, err = db.GetCollection("bind_user").UpdateOne(ctx, bson.M{"uid": bindUser.Uid}, bson.M{"$set": bindUser})
		if err != nil {
			objLog.Errorln("LoginFromGitea update token error:", err)
			return err
		}

		return nil
	}

	bindUser = &model.BindUser{
		Uid:          me.Uid,
		Type:         model.BindTypeGithub,
		Email:        giteaUser.Email,
		Tuid:         int(giteaUser.ID),
		Username:     giteaUser.UserName,
		Name:         model.DisplayName(giteaUser),
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Avatar:       giteaUser.AvatarURL,
	}
	if !token.Expiry.IsZero() {
		bindUser.Expire = int(token.Expiry.Unix())
	}
	id, idErr := db.NextID("bind_user")
	if idErr != nil {
		objLog.Errorln("LoginFromGitea NextID error:", idErr)
		return idErr
	}
	bindUser.Id = id
	_, err = db.GetCollection("bind_user").InsertOne(ctx, bindUser)
	if err != nil {
		objLog.Errorln("LoginFromGitea insert bindUser error:", err)
		return err
	}

	return nil
}

func (ThirdUserLogic) UnBindUser(ctx context.Context, bindId interface{}, me *model.Me) error {
	if !DefaultUser.HasPasswd(ctx, me.Uid) {
		return errors.New("请先设置密码！")
	}
	_, err := db.GetCollection("bind_user").DeleteOne(ctx, bson.M{"_id": bindId, "uid": me.Uid})
	return err
}

func (ThirdUserLogic) findUid(thirdUsername string, typ int) int {
	bindUser := &model.BindUser{}
	err := db.GetCollection("bind_user").FindOne(context.Background(), bson.M{"username": thirdUsername, "type": typ}).Decode(bindUser)
	if err != nil {
		logger.Errorln("ThirdUserLogic findUid error:", err)
	}

	return bindUser.Uid
}

func (ThirdUserLogic) githubTokenAndUser(ctx context.Context, code string) (*model.GithubUser, *oauth2.Token, error) {
	token, err := githubConf.Exchange(ctx, code)
	if err != nil {
		return nil, nil, err
	}

	httpClient := githubConf.Client(ctx, token)
	resp, err := httpClient.Get(GithubAPIBaseUrl + "/user")
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	githubUser := &model.GithubUser{}
	err = json.Unmarshal(respBytes, githubUser)
	if err != nil {
		return nil, nil, err
	}

	if githubUser.Id == 0 {
		return nil, nil, errors.New("get github user info error")
	}

	return githubUser, token, nil
}

func (ThirdUserLogic) giteaTokenAndUser(ctx context.Context, code string) (*model.GiteaUser, *oauth2.Token, error) {
	token, err := giteaConf.Exchange(ctx, code)
	if err != nil {
		return nil, nil, err
	}

	httpClient := giteaConf.Client(ctx, token)
	resp, err := httpClient.Get(GiteaAPIBaseUrl + "/user")
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	giteaUser := &model.GiteaUser{}
	err = json.Unmarshal(respBytes, giteaUser)
	if err != nil {
		return nil, nil, err
	}

	if giteaUser.ID == 0 {
		return nil, nil, errors.New("get gitea user info error")
	}

	return giteaUser, token, nil
}
