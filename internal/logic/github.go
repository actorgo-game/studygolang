// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GithubLogic struct{}

var DefaultGithub = GithubLogic{}

type prInfo struct {
	prURL    string
	username string
	avatar   string
	prTime   time.Time
	hadMerge bool
	number   int
}

var noMoreDataErr = errors.New("pull request: no more data")

func (self GithubLogic) PullRequestEvent(ctx context.Context, body []byte) error {
	objLog := GetLogger(ctx)

	result := gjson.ParseBytes(body)

	thePRURL := result.Get("pull_request.url").String()
	objLog.Infoln("GithubLogic PullRequestEvent, url:", thePRURL)

	_prInfo := &prInfo{
		prURL:    thePRURL,
		username: result.Get("pull_request.user.login").String(),
		avatar:   result.Get("pull_request.user.avatar_url").String(),
		prTime:   result.Get("pull_request.created_at").Time(),
		hadMerge: result.Get("pull_request.merged").Bool(),
	}

	err := self.dealFiles(_prInfo)

	objLog.Infoln("pull request deal successfully!")

	go self.statUserTime()

	return err
}

// IssueEvent 处理 issue 的 GitHub 事件
func (self GithubLogic) IssueEvent(ctx context.Context, body []byte) error {
	objLog := GetLogger(ctx)

	var err error

	result := gjson.ParseBytes(body)
	id := int(result.Get("issue.number").Int())

	labels := result.Get("issue.labels").Array()
	label := ""
	if len(labels) > 0 {
		label = labels[0].Get("name").String()
	}

	title := result.Get("issue.title").String()

	action := result.Get("action").String()
	if action == "opened" {
		err = self.insertIssue(id, title, label)
	} else if action == "labeled" || action == "unlabeled" {
		gcttIssue := &model.GCTTIssue{}
		db.GetCollection("gctt_issue").FindOne(ctx, bson.M{"_id": id}).Decode(gcttIssue)
		if gcttIssue.Id == 0 {
			self.insertIssue(id, title, label)
		} else {
			if label == model.LabelUnClaim {
				gcttIssue.Translator = ""
				gcttIssue.TranslatingAt = 0
			}

			gcttIssue.Label = label
			_, err = db.GetCollection("gctt_issue").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{
				"translator":     gcttIssue.Translator,
				"translating_at": gcttIssue.TranslatingAt,
				"label":          gcttIssue.Label,
			}})
		}
	} else if action == "closed" {
		closedAt := result.Get("issue.closed_at").Time().Unix()
		_, err = db.GetCollection("gctt_issue").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{
			"state":         model.IssueClosed,
			"translated_at": closedAt,
		}})
	} else if action == "reopened" {
		_, err = db.GetCollection("gctt_issue").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{
			"state":         model.IssueOpened,
			"translated_at": 0,
		}})
	}

	if err != nil {
		objLog.Errorln("GithubLogic IssueEvent error:", err)
	}

	return nil
}

// IssueCommentEvent 处理 issue Comment 的 GitHub 事件
func (self GithubLogic) IssueCommentEvent(ctx context.Context, body []byte) error {
	objLog := GetLogger(ctx)
	var err error

	result := gjson.ParseBytes(body)

	id := int(result.Get("issue.number").Int())
	action := result.Get("action").String()

	if action == "created" {
		comments := result.Get("issue.comments").Int()
		if comments == 0 {
			githubUser := result.Get("comment.user.login").String()
			email := self.findUserEmail(githubUser)

			_, err = db.GetCollection("gctt_issue").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{
				"email":          email,
				"translator":     result.Get("comment.user.login").String(),
				"translating_at": result.Get("comment.created_at").Time().Unix(),
			}})
		}
	}

	if err != nil {
		objLog.Errorln("GithubLogic IssueCommentEvent error:", err)
	}

	return nil
}

// RemindTranslator 提醒译者注认领任的翻译进度
func (self GithubLogic) RemindTranslator() error {
	return nil
}

func (self GithubLogic) PullPR(repo string, isAll bool) error {
	if !isAll {
		err := self.pullPR(repo, 1)
		self.statUserTime()
		return err
	}

	var (
		err  error
		page = 1
	)

	for {
		err = self.pullPR(repo, page, "asc")
		if err == noMoreDataErr {
			break
		}

		page++
	}

	self.statUserTime()

	return err
}

func (self GithubLogic) SyncIssues(repo string, isAll bool) error {
	if !isAll {
		err := self.syncIssues(repo, 1)
		return err
	}

	var (
		err  error
		page = 1
	)

	for {
		err = self.syncIssues(repo, page, "asc")
		if err == noMoreDataErr {
			break
		}

		page++
	}

	return err
}

func (self GithubLogic) syncIssues(repo string, page int, directions ...string) error {
	issueListURL := fmt.Sprintf("%s/repos/%s/issues?state=all&per_page=30&page=%d", GithubAPIBaseUrl, repo, page)
	if len(directions) > 0 {
		issueListURL += "&direction=" + directions[0]
	}

	issueListURL = self.addBasicAuth(issueListURL)

	resp, err := http.Get(issueListURL)
	if err != nil {
		logger.Errorln("GithubLogic syncIssues http get error:", err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Errorln("GithubLogic syncIssues read all error:", err)
		return err
	}

	result := gjson.ParseBytes(body)

	if len(result.Array()) == 0 {
		return noMoreDataErr
	}

	ctx := context.Background()
	var outErr error

	result.ForEach(func(key, val gjson.Result) bool {
		if val.Get("pull_request").Exists() {
			return true
		}

		labels := val.Get("labels").Array()
		label := ""
		if len(labels) > 0 {
			label = labels[0].Get("name").String()
		}

		if label != model.LabelUnClaim && label != model.LabelClaimed {
			return true
		}

		id := int(val.Get("number").Int())

		gcttIssue := &model.GCTTIssue{}

		err := db.GetCollection("gctt_issue").FindOne(ctx, bson.M{"_id": id}).Decode(gcttIssue)
		if err != nil && err != mongo.ErrNoDocuments {
			outErr = err
			return true
		}

		var state uint8 = model.IssueClosed
		issueState := val.Get("state").String()
		if issueState == "open" {
			state = model.IssueOpened
		} else {
			gcttIssue.TranslatedAt = val.Get("closed_at").Time().Unix()

			if gcttIssue.State == model.IssueClosed {
				return true
			}
		}
		gcttIssue.State = state
		gcttIssue.Title = val.Get("title").String()
		gcttIssue.Label = label

		if label == model.LabelClaimed {
			translator, createdAt := self.findTranslatorComment(val.Get("comments_url").String())
			if translator == "" {
				translator = val.Get("user.login").String()
				createdAt = val.Get("created_at").Time().Unix()
			}

			gcttIssue.Translator = translator
			gcttIssue.TranslatingAt = createdAt

			gcttIssue.Email = self.findUserEmail(translator)
		}

		if gcttIssue.Id > 0 {
			_, outErr = db.GetCollection("gctt_issue").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": gcttIssue})
		} else {
			gcttIssue.Id = id
			_, outErr = db.GetCollection("gctt_issue").InsertOne(ctx, gcttIssue)
		}

		return true
	})

	return outErr
}

func (self GithubLogic) findTranslatorComment(commentsURL string) (string, int64) {
	commentsURL = self.addBasicAuth(commentsURL)
	resp, err := http.Get(commentsURL)
	if err != nil {
		logger.Errorln("github fetch comments error:", err, "url:", commentsURL)
		return "", 0
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Errorln("github read comments resp error:", err)
		return "", 0
	}
	commentsResult := gjson.ParseBytes(body)
	if len(commentsResult.Array()) == 0 {
		return "", 0
	}

	translatorComment := commentsResult.Array()[0]
	translator := translatorComment.Get("user.login").String()
	createdAt := translatorComment.Get("created_at").Time()

	return translator, createdAt.Unix()
}

func (self GithubLogic) pullPR(repo string, page int, directions ...string) error {
	prListURL := fmt.Sprintf("%s/repos/%s/pulls?state=all&per_page=30&page=%d", GithubAPIBaseUrl, repo, page)

	if len(directions) > 0 {
		prListURL += "&direction=" + directions[0]
	}

	prListURL = self.addBasicAuth(prListURL)

	resp, err := http.Get(prListURL)
	if err != nil {
		logger.Errorln("GithubLogic PullPR get error:", err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Errorln("GithubLogic PullPR read all error:", err)
		return err
	}

	result := gjson.ParseBytes(body)

	if len(result.Array()) == 0 {
		return noMoreDataErr
	}

	var outErr error

	result.ForEach(func(key, val gjson.Result) bool {
		_prInfo := &prInfo{
			prURL:    val.Get("url").String(),
			username: val.Get("user.login").String(),
			avatar:   val.Get("user.avatar_url").String(),
			prTime:   val.Get("created_at").Time(),
			hadMerge: val.Get("merged_at").Type != gjson.Null,
			number:   int(val.Get("number").Int()),
		}

		err = self.dealFiles(_prInfo)
		if err != nil {
			outErr = err
		}

		return true
	})

	return outErr
}

func (self GithubLogic) dealFiles(_prInfo *prInfo) error {
	if _prInfo.prURL == "" {
		return nil
	}

	filesURL := self.addBasicAuth(_prInfo.prURL + "/files")
	resp, err := http.Get(filesURL)
	if err != nil {
		logger.Errorln("github fetch files error:", err, "url:", filesURL)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Errorln("github read files resp error:", err)
		return err
	}
	filesResult := gjson.ParseBytes(body)

	length := len(filesResult.Array())
	if length == 1 {
		err = self.translating(filesResult, _prInfo)
	} else if length == 2 {
		err = self.translated(filesResult, _prInfo)
	} else if length == 3 {
		err = self.translateSilmu(filesResult, _prInfo)
	}

	return err
}

func (self GithubLogic) translating(filesResult gjson.Result, _prInfo *prInfo) error {
	var outErr error
	filesResult.ForEach(func(key, val gjson.Result) bool {
		filename := val.Get("filename").String()
		if !strings.HasPrefix(filename, "sources") {

			if strings.HasPrefix(filename, "translated") {
				filenames := strings.SplitN(filename, "/", 3)
				if len(filenames) < 3 {
					return true
				}
				title := filenames[2]
				if title == "" {
					return true
				}

				err := self.issueTranslated(_prInfo, title)
				if err != nil {
					outErr = err
				}
			}

			return true
		}

		filenames := strings.SplitN(filename, "/", 3)
		if len(filenames) < 3 {
			return true
		}
		title := filenames[2]
		if title == "" {
			return true
		}

		status := val.Get("status").String()
		if status == "modified" && _prInfo.hadMerge {
			err := self.insertOrUpdateGCCT(_prInfo, title, false)
			if err != nil {
				outErr = err
			}
		}
		return true
	})

	return outErr
}

func (self GithubLogic) issueTranslated(_prInfo *prInfo, title string) error {
	ctx := context.Background()
	md5 := goutils.Md5(title)
	gcttGit := &model.GCTTGit{}
	err := db.GetCollection("gctt_git").FindOne(ctx, bson.M{"md5": md5}).Decode(gcttGit)
	if err != nil && err != mongo.ErrNoDocuments {
		logger.Errorln("GithubLogic insertOrUpdateGCCT get error:", err)
		return err
	}

	if gcttGit.Id > 0 {
		return nil
	}

	gcttUser := DefaultGCTT.FindOne(nil, _prInfo.username)

	session, sessErr := db.GetClient().StartSession()
	if sessErr != nil {
		return sessErr
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		if gcttUser.Id == 0 {
			gcttUser.Username = _prInfo.username
			gcttUser.Avatar = _prInfo.avatar
			gcttUser.JoinedAt = _prInfo.prTime.Unix()
			id, idErr := db.NextID("gctt_user")
			if idErr != nil {
				return nil, idErr
			}
			gcttUser.Id = id
			_, txErr := db.GetCollection("gctt_user").InsertOne(sc, gcttUser)
			if txErr != nil {
				return nil, txErr
			}
		}

		gcttGit.Username = _prInfo.username
		gcttGit.Title = title
		gcttGit.Md5 = md5
		gcttGit.PR = _prInfo.number
		gcttGit.TranslatedAt = _prInfo.prTime.Unix()
		id, idErr := db.NextID("gctt_git")
		if idErr != nil {
			return nil, idErr
		}
		gcttGit.Id = id
		_, txErr := db.GetCollection("gctt_git").InsertOne(sc, gcttGit)
		if txErr != nil {
			return nil, txErr
		}
		return nil, nil
	})

	return err
}

func (self GithubLogic) translated(filesResult gjson.Result, _prInfo *prInfo) error {
	var (
		sourceTitle  string
		isTranslated = true
	)

	filesResult.ForEach(func(key, val gjson.Result) bool {
		if !isTranslated {
			return false
		}

		status := val.Get("status").String()
		filename := val.Get("filename").String()

		if status == "removed" {
			if strings.HasPrefix(filename, "sources") {
				filenames := strings.SplitN(filename, "/", 3)
				if len(filenames) < 3 {
					return true
				}
				sourceTitle = filenames[2]
			} else {
				isTranslated = false
			}
		} else if status == "added" {
			if !strings.HasPrefix(filename, "translated") {
				isTranslated = false
			}
		}

		return true
	})

	if !isTranslated || sourceTitle == "" {
		return nil
	}

	return self.insertOrUpdateGCCT(_prInfo, sourceTitle, true)
}

func (self GithubLogic) translateSilmu(filesResult gjson.Result, _prInfo *prInfo) error {
	var (
		sourceTitle  string
		isTranslated = true
	)

	filesResult.ForEach(func(key, val gjson.Result) bool {
		if !isTranslated {
			return false
		}

		status := val.Get("status").String()
		filename := val.Get("filename").String()

		if status == "removed" {
			if strings.HasPrefix(filename, "sources") {
				filenames := strings.SplitN(filename, "/", 3)
				if len(filenames) < 3 {
					return true
				}
				sourceTitle = filenames[2]
			} else {
				isTranslated = false
			}
		} else if status == "added" {
			if !strings.HasPrefix(filename, "translated") {
				isTranslated = false
			}
		} else if status == "modified" {
			if strings.HasPrefix(filename, "sources") {
				filenames := strings.SplitN(filename, "/", 3)
				if len(filenames) < 3 {
					return true
				}
				title := filenames[2]
				if title == "" {
					return true
				}

				self.insertOrUpdateGCCT(_prInfo, title, false)
			}
		}

		return true
	})

	if !isTranslated || sourceTitle == "" {
		return nil
	}

	return self.insertOrUpdateGCCT(_prInfo, sourceTitle, true)
}

func (GithubLogic) insertOrUpdateGCCT(_prInfo *prInfo, title string, isTranslated bool) error {
	ctx := context.Background()
	md5 := goutils.Md5(title)
	gcttGit := &model.GCTTGit{}
	err := db.GetCollection("gctt_git").FindOne(ctx, bson.M{"md5": md5}).Decode(gcttGit)
	if err != nil && err != mongo.ErrNoDocuments {
		logger.Errorln("GithubLogic insertOrUpdateGCCT get error:", err)
		return err
	}
	if gcttGit.Id > 0 {
		if gcttGit.Username != _prInfo.username {
			return nil
		}
	}

	gcttUser := DefaultGCTT.FindOne(nil, _prInfo.username)

	session, sessErr := db.GetClient().StartSession()
	if sessErr != nil {
		return sessErr
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		if gcttUser.Id == 0 {
			gcttUser.Username = _prInfo.username
			gcttUser.Avatar = _prInfo.avatar
			gcttUser.JoinedAt = _prInfo.prTime.Unix()
			id, idErr := db.NextID("gctt_user")
			if idErr != nil {
				return nil, idErr
			}
			gcttUser.Id = id
			_, txErr := db.GetCollection("gctt_user").InsertOne(sc, gcttUser)
			if txErr != nil {
				return nil, txErr
			}
		}

		if gcttGit.Id > 0 {
			if gcttGit.TranslatedAt == 0 && isTranslated {
				gcttGit.TranslatedAt = _prInfo.prTime.Unix()
				gcttGit.PR = _prInfo.number
				_, txErr := db.GetCollection("gctt_git").UpdateOne(sc, bson.M{"_id": gcttGit.Id}, bson.M{"$set": gcttGit})
				if txErr != nil {
					return nil, txErr
				}
			}
			return nil, nil
		}

		gcttGit.PR = _prInfo.number
		gcttGit.Username = _prInfo.username
		gcttGit.Title = title
		gcttGit.Md5 = md5
		gcttGit.TranslatingAt = _prInfo.prTime.Unix()
		id, idErr := db.NextID("gctt_git")
		if idErr != nil {
			return nil, idErr
		}
		gcttGit.Id = id
		_, txErr := db.GetCollection("gctt_git").InsertOne(sc, gcttGit)
		if txErr != nil {
			return nil, txErr
		}
		return nil, nil
	})

	return err
}

func (GithubLogic) statUserTime() {
	ctx := context.Background()
	gcttUsers := make([]*model.GCTTUser, 0)
	cursor, err := db.GetCollection("gctt_user").Find(ctx, bson.M{})
	if err != nil {
		logger.Errorln("GithubLogic statUserTime find error:", err)
		return
	}
	cursor.All(ctx, &gcttUsers)
	cursor.Close(ctx)

	for _, gcttUser := range gcttUsers {
		gcttGits := make([]*model.GCTTGit, 0)
		opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
		gitCursor, gitErr := db.GetCollection("gctt_git").Find(ctx, bson.M{"username": gcttUser.Username, "pr": bson.M{"$ne": 0}}, opts)
		if gitErr != nil {
			logger.Errorln("GithubLogic find gctt git error:", gitErr)
			continue
		}
		gitCursor.All(ctx, &gcttGits)
		gitCursor.Close(ctx)

		var avgTime, lastAt int64
		var words int
		for _, gcttGit := range gcttGits {
			if gcttGit.TranslatingAt != 0 && gcttGit.TranslatedAt != 0 {
				avgTime += gcttGit.TranslatedAt - gcttGit.TranslatingAt
			}

			if gcttGit.TranslatedAt > lastAt {
				lastAt = gcttGit.TranslatedAt
			}

			if gcttGit.Words == 0 && gcttGit.ArticleId > 0 {
				article, _ := DefaultArticle.FindById(nil, gcttGit.ArticleId)
				gcttGit.Words = utf8.RuneCountInString(article.Content)
			}

			words += gcttGit.Words

			db.GetCollection("gctt_git").UpdateOne(ctx, bson.M{"_id": gcttGit.Id}, bson.M{"$set": gcttGit})
		}

		uid := DefaultThirdUser.findUid(gcttUser.Username, model.BindTypeGithub)

		gcttUser.Num = len(gcttGits)
		gcttUser.Words = words
		if gcttUser.Num > 0 {
			gcttUser.AvgTime = int(avgTime) / gcttUser.Num
		}
		gcttUser.LastAt = lastAt
		gcttUser.Uid = uid
		_, err = db.GetCollection("gctt_user").UpdateOne(ctx, bson.M{"_id": gcttUser.Id}, bson.M{"$set": gcttUser})
		if err != nil {
			logger.Errorln("GithubLogic update gctt user error:", err)
		}
	}
}

func (self GithubLogic) insertIssue(id int, title, label string) error {
	gcttIssue := &model.GCTTIssue{
		Id:    id,
		Title: title,
		Label: label,
	}
	_, err := db.GetCollection("gctt_issue").InsertOne(context.Background(), gcttIssue)
	return err
}

func (self GithubLogic) findUserEmail(githubUser string) string {
	ctx := context.Background()
	bindUser := &model.BindUser{}
	db.GetCollection("bind_user").FindOne(ctx, bson.M{"username": githubUser, "type": model.BindTypeGithub}).Decode(bindUser)
	if !strings.HasSuffix(bindUser.Email, "@github.com") {
		return bindUser.Email
	}

	if bindUser.Uid != 0 {
		user := DefaultUser.findUser(nil, bindUser.Uid)
		if !strings.HasSuffix(user.Email, "@github.com") {
			return user.Email
		}
	}

	gcttIssue := &model.GCTTIssue{}
	db.GetCollection("gctt_issue").FindOne(ctx, bson.M{"translator": githubUser, "email": bson.M{"$ne": ""}}).Decode(gcttIssue)
	return gcttIssue.Email
}

func (self GithubLogic) addBasicAuth(netURL string) string {
	password, ok := os.LookupEnv("GITHUB_PASSWORD")
	if ok {
		return netURL[:8] + "polaris1119:" + password + "@" + netURL[8:]
	}

	return netURL
}
