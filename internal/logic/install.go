package logic

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/polaris1119/config"
	xcontext "golang.org/x/net/context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InstallLogic struct{}

var DefaultInstall = InstallLogic{}

func (InstallLogic) CreateTable(ctx xcontext.Context) error {
	objLog := GetLogger(ctx)

	bgCtx := context.Background()

	collections := []string{
		"user_info", "user_login", "user_active", "user_role", "bind_user",
		"topics", "topics_ex", "topics_node", "topic_append", "recommend_node",
		"articles", "article_gctt", "crawl_rule", "auto_crawl_rule",
		"comments", "resource", "resource_ex", "resource_category",
		"feed", "message", "system_message", "favorite", "like",
		"view_record", "view_source", "dynamic", "download",
		"gift", "gift_redeem", "user_exchange_record",
		"open_project", "subject", "subject_admin", "subject_article", "subject_follower",
		"morning_reading", "interview_question", "learning_material", "friend_link",
		"advertisement", "page_ad", "wiki", "book",
		"role", "role_authority", "authority",
		"website_setting", "user_setting", "default_avatar",
		"image", "search_stat", "mission", "user_login_mission",
		"user_balance_detail", "user_recharge",
		"wechat_user", "wechat_auto_reply",
		"gctt_user", "gctt_git", "gctt_issue", "gctt_timeline",
		"github_user", "counters",
	}

	for _, name := range collections {
		err := db.MasterDB.CreateCollection(bgCtx, name)
		if err != nil {
			if cmdErr, ok := err.(mongo.CommandError); ok && cmdErr.Code == 48 {
				continue
			}
			objLog.Errorln("create collection error:", name, err)
		}
	}

	indexes := map[string][]mongo.IndexModel{
		"user_info": {
			{Keys: bson.D{{"username", 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{"email", 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{"status", 1}}},
		},
		"user_login": {
			{Keys: bson.D{{"username", 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{"email", 1}}},
		},
		"topics": {
			{Keys: bson.D{{"uid", 1}}},
			{Keys: bson.D{{"nid", 1}}},
			{Keys: bson.D{{"ctime", -1}}},
			{Keys: bson.D{{"flag", 1}}},
			{Keys: bson.D{{"top", 1}}},
		},
		"articles": {
			{Keys: bson.D{{"domain", 1}}},
			{Keys: bson.D{{"status", 1}}},
			{Keys: bson.D{{"ctime", -1}}},
			{Keys: bson.D{{"url", 1}}},
		},
		"comments": {
			{Keys: bson.D{{"objid", 1}, {"objtype", 1}}},
			{Keys: bson.D{{"uid", 1}}},
		},
		"resource": {
			{Keys: bson.D{{"uid", 1}}},
			{Keys: bson.D{{"ctime", -1}}},
			{Keys: bson.D{{"catid", 1}}},
		},
		"feed": {
			{Keys: bson.D{{"uid", 1}}},
			{Keys: bson.D{{"created_at", -1}}},
			{Keys: bson.D{{"objtype", 1}, {"objid", 1}}},
		},
		"like": {
			{Keys: bson.D{{"uid", 1}, {"objid", 1}, {"objtype", 1}}, Options: options.Index().SetUnique(true)},
		},
		"favorite": {
			{Keys: bson.D{{"uid", 1}, {"objid", 1}, {"objtype", 1}}, Options: options.Index().SetUnique(true)},
		},
		"message": {
			{Keys: bson.D{{"to", 1}, {"hasread", 1}}},
		},
		"view_record": {
			{Keys: bson.D{{"uid", 1}, {"objid", 1}, {"objtype", 1}}},
		},
		"open_project": {
			{Keys: bson.D{{"uri", 1}}},
			{Keys: bson.D{{"ctime", -1}}},
		},
		"wiki": {
			{Keys: bson.D{{"uid", 1}}},
		},
		"book": {
			{Keys: bson.D{{"uid", 1}}},
		},
		"bind_user": {
			{Keys: bson.D{{"uid", 1}}},
			{Keys: bson.D{{"type", 1}, {"tuid", 1}}},
		},
	}

	for coll, idxModels := range indexes {
		_, err := db.GetCollection(coll).Indexes().CreateMany(bgCtx, idxModels)
		if err != nil {
			objLog.Errorln("create indexes error:", coll, err)
		}
	}

	return nil
}

func (InstallLogic) InitTable(ctx xcontext.Context) error {
	objLog := GetLogger(ctx)
	bgCtx := context.Background()

	total, err := db.GetCollection("role").CountDocuments(bgCtx, bson.M{})
	if err != nil {
		return err
	}
	if total > 0 {
		return nil
	}

	initFile := config.ROOT + "/config/init.json"
	buf, err := ioutil.ReadFile(initFile)
	if err != nil {
		objLog.Errorln("init table, read init file error:", err)
		return err
	}

	var initData map[string][]interface{}
	if err = json.Unmarshal(buf, &initData); err != nil {
		objLog.Errorln("init table, parse json error:", err)
		return err
	}

	for collName, docs := range initData {
		if len(docs) == 0 {
			continue
		}
		_, err := db.GetCollection(collName).InsertMany(bgCtx, docs)
		if err != nil {
			objLog.Errorln("init table insert error:", collName, err)
		}
	}

	return nil
}

func (InstallLogic) IsTableExist(ctx xcontext.Context) bool {
	bgCtx := context.Background()
	names, err := db.MasterDB.ListCollectionNames(bgCtx, bson.M{})
	if err != nil {
		return false
	}

	for _, name := range names {
		if name == "user_info" {
			return true
		}
	}
	return false
}

func (InstallLogic) HadRootUser(ctx xcontext.Context) bool {
	bgCtx := context.Background()
	user := &model.User{}
	err := db.GetCollection("user_info").FindOne(bgCtx, bson.M{"is_root": true}).Decode(user)
	if err != nil {
		return false
	}
	return user.Uid != 0
}
