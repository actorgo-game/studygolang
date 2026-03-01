// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author:polaris	polaris@studygolang.com

package logic

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jaytaylor/html2text"
	"github.com/polaris1119/config"
	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"github.com/polaris1119/set"
	"github.com/polaris1119/slices"
	"github.com/polaris1119/times"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/text/encoding/simplifiedchinese"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/global"
	"github.com/studygolang/studygolang/internal/model"
)

type ArticleLogic struct{}

var DefaultArticle = ArticleLogic{}

var domainPatch = map[string]string{
	"iteye.com":      "iteye.com",
	"blog.51cto.com": "blog.51cto.com",
}

var articleRe = regexp.MustCompile("[\r　\n  \t\v]+")
var articleSpaceRe = regexp.MustCompile("[ ]+")

// ParseArticle 获取 url 对应的文章并根据规则进行解析
func (self ArticleLogic) ParseArticle(ctx context.Context, articleUrl string, auto bool) (*model.Article, error) {
	articleUrl = strings.TrimSpace(articleUrl)
	if !strings.HasPrefix(articleUrl, "http") {
		articleUrl = "http://" + articleUrl
	}

	articleUrl = self.cleanUrl(articleUrl, auto)

	tmpArticle := &model.Article{}
	err := db.GetCollection("articles").FindOne(ctx, bson.M{"url": articleUrl}).Decode(tmpArticle)
	if err != nil && err != mongo.ErrNoDocuments {
		logger.Infoln(articleUrl, "find error:", err)
		return nil, errors.New("has exists!")
	}
	if tmpArticle.Id != 0 {
		tmpArticle.AfterLoad()
	}
	if tmpArticle.Id != 0 && auto {
		logger.Infoln(articleUrl, "has exists")
		return nil, errors.New("has exists!")
	}

	urlPaths := strings.SplitN(articleUrl, "/", 5)
	domain := urlPaths[2]

	for k, v := range domainPatch {
		if strings.Contains(domain, k) && !strings.Contains(domain, "www."+k) {
			domain = v
			break
		}
	}

	rule := &model.CrawlRule{}
	err = db.GetCollection("crawl_rule").FindOne(ctx, bson.M{"domain": domain}).Decode(rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return self.ParseArticleByAccuracy(articleUrl, tmpArticle, auto)
		}
		logger.Errorln("find rule by domain error:", err)
		return nil, err
	}

	var doc *goquery.Document

	ua := `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.116 Safari/537.36`
	req, err := http.NewRequest("GET", articleUrl, nil)
	if err != nil {
		logger.Errorln("new request error:", err)
		return nil, err
	}
	req.Header.Add("User-Agent", ua)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorln("get response error:", err)
		return nil, err
	}
	defer resp.Body.Close()
	if doc, err = goquery.NewDocumentFromReader(resp.Body); err != nil {
		logger.Errorln("goquery NewDocumentFromReader error:", err)
		return nil, err
	}

	author := ""
	if rule.InUrl {
		index, err := strconv.Atoi(rule.Author)
		if err != nil {
			logger.Errorln("author rule is illegal:", rule.Author, "error:", err)
			return nil, err
		}
		author = urlPaths[index]
	} else {
		authorSelection := doc.Find(rule.Author)
		if authorSelection.Is(rule.Author) {
			author = strings.TrimSpace(authorSelection.Text())
		} else if strings.HasPrefix(rule.Author, "/") {
			re, err := regexp.Compile(rule.Author[1:])
			if err != nil {
				logger.Errorln("author regexp error:", err)
				return nil, err
			}
			body, _ := doc.Find("body").Html()
			authorResult := re.FindStringSubmatch(body)
			if len(authorResult) < 2 {
				logger.Errorln("no author found:", rule.Domain)
				return nil, errors.New("no author found!")
			}

			author = authorResult[1]
		} else {
			author = rule.Author
		}
	}

	filters := config.ConfigFile.MustValueArray("crawl", "filter", ",")
	for _, filter := range filters {
		if filter == author {
			return nil, errors.New(author + "'s article, skip")
		}
	}

	title := ""
	doc.Find(rule.Title).Each(func(i int, selection *goquery.Selection) {
		if title != "" {
			return
		}

		tmpTitle := strings.TrimSpace(selection.Text())
		tmpTitle = strings.TrimSpace(strings.Trim(tmpTitle, "原"))
		tmpTitle = strings.TrimSpace(strings.Trim(tmpTitle, "荐"))
		tmpTitle = strings.TrimSpace(strings.Trim(tmpTitle, "转"))
		tmpTitle = strings.TrimSpace(strings.Trim(tmpTitle, "顶"))
		if tmpTitle != "" {
			title = tmpTitle
		}
	})

	if title == "" {
		logger.Errorln("url:", articleUrl, "parse title error:", err)
		return nil, err
	}

	replacer := strings.NewReplacer("[置顶]", "", "[原]", "", "[转]", "")
	title = strings.TrimSpace(replacer.Replace(title))

	contentSelection := doc.Find(rule.Content)

	imgDeny := false
	extMap := rule.ParseExt()
	if extMap != nil {
		if deny, ok := extMap["img_deny"]; ok {
			imgDeny = goutils.MustBool(deny)
		}
	}

	contentSelection.Find("img").Each(func(i int, s *goquery.Selection) {
		self.transferImage(ctx, s, imgDeny, domain)
	})

	content, err := contentSelection.Html()
	if err != nil {
		logger.Errorln("goquery parse content error:", err)
		return nil, err
	}
	content = strings.TrimSpace(content)
	txt := strings.TrimSpace(contentSelection.Text())
	txt = articleRe.ReplaceAllLiteralString(txt, " ")
	txt = articleSpaceRe.ReplaceAllLiteralString(txt, " ")

	if auto && len(txt) < 300 {
		logger.Errorln(articleUrl, "content is short")
		return nil, errors.New("content is short")
	}

	if auto && strings.Count(content, "<a") > config.ConfigFile.MustInt("crawl", "contain_link", 10) {
		logger.Errorln(articleUrl, "content contains too many link!")
		return nil, errors.New("content contains too many link")
	}

	pubDate := times.Format("Y-m-d H:i:s")
	if rule.PubDate != "" {
		pubDate = strings.TrimSpace(doc.Find(rule.PubDate).First().Text())
	}

	if pubDate == "" {
		pubDate = times.Format("Y-m-d H:i:s")
	} else {
		if len(pubDate) == 16 && auto {
			pubTime, err := time.ParseInLocation("2006-01-02 15:04", pubDate, time.Local)
			if err == nil {
				if pubTime.Add(3 * 30 * 86400 * time.Second).Before(time.Now()) {
					return nil, errors.New("article is old!")
				}
			}
		}
	}

	article := &model.Article{
		Domain:    domain,
		Name:      rule.Name,
		Author:    author,
		AuthorTxt: author,
		Title:     title,
		Content:   content,
		Txt:       txt,
		PubDate:   pubDate,
		Url:       articleUrl,
		Lang:      rule.Lang,
	}

	if extMap != nil {
		err = self.convertByExt(extMap, article)
		if err != nil {
			return nil, err
		}
	}

	if !auto && tmpArticle.Id > 0 {
		updateDoc := bson.M{
			"domain":     article.Domain,
			"name":       article.Name,
			"author":     article.Author,
			"author_txt": article.AuthorTxt,
			"title":      article.Title,
			"content":    article.Content,
			"txt":        article.Txt,
			"pub_date":   article.PubDate,
			"url":        article.Url,
		}
		if article.Lang != 0 {
			updateDoc["lang"] = article.Lang
		}
		if article.Css != "" {
			updateDoc["css"] = article.Css
		}
		_, err = db.GetCollection("articles").UpdateOne(ctx, bson.M{"_id": tmpArticle.Id}, bson.M{"$set": updateDoc})
		if err != nil {
			logger.Errorln("upadate article error:", err)
			return nil, err
		}
		article.Id = tmpArticle.Id
		article.AfterLoad()
		return article, nil
	}

	article.BeforeInsert()
	newID, err := db.NextID("articles")
	if err != nil {
		logger.Errorln("NextID for articles error:", err)
		return nil, err
	}
	article.Id = newID

	_, err = db.GetCollection("articles").InsertOne(ctx, article)
	if err != nil {
		logger.Errorln("insert article error:", err)
		return nil, err
	}

	article.AfterLoad()
	article.AfterInsert()

	return article, nil
}

func (self ArticleLogic) ParseZhihuArticle(ctx context.Context, articleUrl string, rule *model.CrawlRule) (*model.Article, error) {
	var (
		doc *goquery.Document
		err error
	)
	if doc, err = goquery.NewDocument(articleUrl); err != nil {
		logger.Errorln("goquery newdocument error:", err)
		return nil, err
	}

	var (
		jsonContentKey string
		ok             bool
	)

	extMap := rule.ParseExt()
	if jsonContentKey, ok = extMap["json_content"]; !ok {
		return nil, errors.New("zhihu config error, not json_content key")
	}

	jsonContent := doc.Find(jsonContentKey).Text()
	if jsonContent == "" {
		return nil, errors.New("zhihu json content is empty")
	}

	pos := strings.LastIndex(articleUrl, "/")
	articleId := articleUrl[pos+1:]

	result := gjson.Parse(jsonContent)
	database := result.Get("database")
	post := database.Get("Post").Get(articleId)
	author := database.Get("User").Get(post.Get("author").String()).Get("name").String()
	content := post.Get("content").String()
	txt, _ := html2text.FromString(content)
	pubDate, _ := time.Parse("2006-01-02T15:04:05+08:00", post.Get("publishedTime").String())

	article := &model.Article{
		Domain:    rule.Domain,
		Name:      rule.Name,
		Author:    author,
		AuthorTxt: author,
		Title:     post.Get("title").String(),
		Content:   content,
		Txt:       txt,
		PubDate:   times.Format("Y-m-d H:i:s", pubDate),
		Url:       articleUrl,
		Lang:      rule.Lang,
	}

	article.BeforeInsert()
	newID, err := db.NextID("articles")
	if err != nil {
		logger.Errorln("NextID for articles error:", err)
		return nil, err
	}
	article.Id = newID

	_, err = db.GetCollection("articles").InsertOne(ctx, article)
	if err != nil {
		logger.Errorln("insert article error:", err)
		return nil, err
	}

	article.AfterLoad()
	article.AfterInsert()

	return article, nil
}

// Publish 发布文章
func (self ArticleLogic) Publish(ctx context.Context, me *model.Me, form url.Values) (int, error) {
	objLog := GetLogger(ctx)

	var uid = me.Uid

	article := &model.Article{
		Domain:    WebsiteSetting.Domain,
		Name:      WebsiteSetting.Name,
		Author:    me.Username,
		AuthorTxt: me.Username,
		Title:     form.Get("title"),
		Cover:     form.Get("cover"),
		Content:   form.Get("content"),
		Txt:       form.Get("txt"),
		Markdown:  goutils.MustBool(form.Get("markdown"), false),
		PubDate:   times.Format("Y-m-d H:i:s"),
		GCTT:      goutils.MustBool(form.Get("gctt"), false),
	}

	if article.Txt == "" {
		article.Txt = article.Content
	}

	requestIdInter := ctx.Value("request_id")
	if requestIdInter != nil {
		if requestId, ok := requestIdInter.(string); ok {
			_ = requestId
		}
	}

	// GCTT 译文，如果译者关联了本站账号，author 改为译者
	if article.GCTT {
		translator := form.Get("translator")
		gcttUser := &model.GCTTUser{}
		err := db.GetCollection("gctt_user").FindOne(ctx, bson.M{"username": translator}).Decode(gcttUser)
		if err != nil && err != mongo.ErrNoDocuments {
			objLog.Errorln("article publish find gctt user error:", err)
		}

		if gcttUser.Uid > 0 {
			user := DefaultUser.findUser(ctx, gcttUser.Uid)
			article.Author = user.Username
			article.AuthorTxt = user.Username

			uid = user.Uid

			article.OpUser = me.Username
		}
	}

	article.BeforeInsert()
	newID, err := db.NextID("articles")
	if err != nil {
		objLog.Errorln("NextID for articles error:", err)
		return 0, err
	}
	article.Id = newID
	article.Url = strconv.Itoa(article.Id)

	session, err := db.GetClient().StartSession()
	if err != nil {
		objLog.Errorln("StartSession error:", err)
		return 0, err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		_, insertErr := db.GetCollection("articles").InsertOne(sc, article)
		if insertErr != nil {
			return nil, insertErr
		}

		if article.GCTT {
			articleGCTT := &model.ArticleGCTT{
				ArticleID:  article.Id,
				Author:     form.Get("author"),
				AuthorURL:  form.Get("author_url"),
				Translator: form.Get("translator"),
				Checker:    form.Get("checker"),
				URL:        form.Get("url"),
			}

			_, insertErr = db.GetCollection("article_gctt").InsertOne(sc, articleGCTT)
			if insertErr != nil {
				return nil, insertErr
			}
		}

		return nil, nil
	})
	if err != nil {
		objLog.Errorln("publish article transaction error:", err)
		return 0, err
	}

	article.AfterLoad()
	article.AfterInsert()

	go publishObservable.NotifyObservers(uid, model.TypeArticle, article.Id)

	return article.Id, nil
}

func (self ArticleLogic) PublishFromAdmin(ctx context.Context, me *model.Me, form url.Values) error {
	objLog := GetLogger(ctx)

	articleUrl := form.Get("url")
	netUrl, err := url.Parse(articleUrl)
	if err != nil {
		objLog.Errorln("url is illegal:", netUrl)
		return err
	}

	article := &model.Article{
		Domain:    netUrl.Host,
		Name:      form.Get("name"),
		Url:       articleUrl,
		Author:    form.Get("author"),
		AuthorTxt: form.Get("author"),
		Title:     form.Get("title"),
		Content:   form.Get("content"),
		Txt:       form.Get("txt"),
		PubDate:   form.Get("pub_date"),
		Lang:      goutils.MustInt(form.Get("lang")),
		Cover:     form.Get("cover"),
	}

	article.BeforeInsert()
	newID, err := db.NextID("articles")
	if err != nil {
		objLog.Errorln("NextID for articles error:", err)
		return err
	}
	article.Id = newID

	_, err = db.GetCollection("articles").InsertOne(ctx, article)
	if err != nil {
		objLog.Errorln("insert article error:", err)
		return err
	}

	article.AfterInsert()

	return nil
}

func (ArticleLogic) cleanUrl(articleUrl string, auto bool) string {
	pos := strings.LastIndex(articleUrl, "#")
	if pos > 0 {
		articleUrl = articleUrl[:pos]
	}
	if auto {
		pos = strings.Index(articleUrl, "?")
		if pos > 0 {
			articleUrl = articleUrl[:pos]
		}
	}

	return articleUrl
}

func (ArticleLogic) convertByExt(extMap map[string]string, article *model.Article) error {
	var err error
	if css, ok := extMap["css"]; ok {
		article.Css = css
	}

	if charset, ok := extMap["charset"]; ok {
		if charset == "gbk" {
			article.Title, err = simplifiedchinese.GBK.NewDecoder().String(article.Title)
			if err != nil {
				logger.Errorln("convert title gbk to utf8 error:", err)
				return err
			}
			article.Content, err = simplifiedchinese.GBK.NewDecoder().String(article.Content)
			if err != nil {
				logger.Errorln("convert content gbk to utf8 error:", err)
				return err
			}
			article.Txt, err = simplifiedchinese.GBK.NewDecoder().String(article.Txt)
			if err != nil {
				logger.Errorln("convert txt gbk to utf8 error:", err)
				return err
			}
			article.AuthorTxt, err = simplifiedchinese.GBK.NewDecoder().String(article.AuthorTxt)
			if err != nil {
				logger.Errorln("convert txt gbk to utf8 error:", err)
				return err
			}
			article.Author = article.AuthorTxt
		}
	}

	return nil
}

func (ArticleLogic) FindLastList(beginTime string, limit int) ([]*model.Article, error) {
	ctx := context.Background()
	filter := bson.M{
		"ctime":  bson.M{"$gt": beginTime},
		"status": bson.M{"$ne": model.ArticleStatusOffline},
	}
	findOpts := options.Find().
		SetSort(bson.D{{"cmtnum", -1}, {"likenum", -1}, {"viewnum", -1}}).
		SetLimit(int64(limit))

	cursor, err := db.GetCollection("articles").Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	articles := make([]*model.Article, 0)
	if err = cursor.All(ctx, &articles); err != nil {
		return nil, err
	}

	for _, article := range articles {
		article.AfterLoad()
	}

	return articles, nil
}

// Total 博文总数
func (ArticleLogic) Total() int64 {
	ctx := context.Background()
	total, err := db.GetCollection("articles").CountDocuments(ctx, bson.M{})
	if err != nil {
		logger.Errorln("ArticleLogic Total error:", err)
	}
	return total
}

// FindBy 获取抓取的文章列表（分页）
func (self ArticleLogic) FindBy(ctx context.Context, limit int, lastIds ...int) []*model.Article {
	objLog := GetLogger(ctx)

	filter := bson.M{
		"status": bson.M{"$in": []int{model.ArticleStatusNew, model.ArticleStatusOnline}},
	}

	if len(lastIds) > 0 && lastIds[0] > 0 {
		filter["_id"] = bson.M{"$lt": lastIds[0]}
	}

	findOpts := options.Find().
		SetSort(bson.D{{"_id", -1}}).
		SetLimit(int64(limit))

	cursor, err := db.GetCollection("articles").Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("ArticleLogic FindBy Error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	articles := make([]*model.Article, 0)
	if err = cursor.All(ctx, &articles); err != nil {
		objLog.Errorln("ArticleLogic FindBy decode error:", err)
		return nil
	}

	for _, article := range articles {
		article.AfterLoad()
	}

	topCursor, err := db.GetCollection("articles").Find(ctx,
		bson.M{"top": 1},
		options.Find().SetSort(bson.D{{"_id", -1}}))
	if err != nil {
		objLog.Errorln("ArticleLogic Find Top Articles Error:", err)
		return nil
	}
	defer topCursor.Close(ctx)

	topArticles := make([]*model.Article, 0)
	if err = topCursor.All(ctx, &topArticles); err != nil {
		objLog.Errorln("ArticleLogic Find Top Articles decode error:", err)
		return nil
	}

	for _, article := range topArticles {
		article.AfterLoad()
	}

	if len(topArticles) > 0 {
		articles = append(topArticles, articles...)
	}

	self.fillUser(articles)

	return articles
}

func (self ArticleLogic) FindTaGCTTArticles(ctx context.Context, translator string) []*model.Article {
	objLog := GetLogger(ctx)

	gcttCursor, err := db.GetCollection("article_gctt").Find(ctx,
		bson.M{"translator": translator},
		options.Find().SetSort(bson.D{{"_id", -1}}))
	if err != nil {
		objLog.Errorln("ArticleLogic FindTaGCTTArticles gctt error:", err)
		return nil
	}
	defer gcttCursor.Close(ctx)

	articleGCTTs := make([]*model.ArticleGCTT, 0)
	if err = gcttCursor.All(ctx, &articleGCTTs); err != nil {
		objLog.Errorln("ArticleLogic FindTaGCTTArticles gctt decode error:", err)
		return nil
	}

	articleIds := make([]int, len(articleGCTTs))
	for i, articleGCTT := range articleGCTTs {
		articleIds[i] = articleGCTT.ArticleID
	}

	if len(articleIds) == 0 {
		return nil
	}

	articleCursor, err := db.GetCollection("articles").Find(ctx, bson.M{"_id": bson.M{"$in": articleIds}})
	if err != nil {
		objLog.Errorln("ArticleLogic FindTaGCTTArticles article error:", err)
		return nil
	}
	defer articleCursor.Close(ctx)

	articleList := make([]*model.Article, 0)
	if err = articleCursor.All(ctx, &articleList); err != nil {
		objLog.Errorln("ArticleLogic FindTaGCTTArticles article decode error:", err)
		return nil
	}

	articleMap := make(map[int]*model.Article, len(articleList))
	for _, a := range articleList {
		a.AfterLoad()
		articleMap[a.Id] = a
	}

	articles := make([]*model.Article, 0, len(articleMap))
	for _, articleGCTT := range articleGCTTs {
		if article, ok := articleMap[articleGCTT.ArticleID]; ok {
			articles = append(articles, article)
		}
	}

	return articles
}

func (self ArticleLogic) FindByUser(ctx context.Context, username string, limit int) []*model.Article {
	objLog := GetLogger(ctx)

	filter := bson.M{
		"author_txt": username,
		"status":     bson.M{"$lt": model.ArticleStatusOffline},
	}
	findOpts := options.Find().
		SetSort(bson.D{{"_id", -1}}).
		SetLimit(int64(limit))

	cursor, err := db.GetCollection("articles").Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("ArticleLogic FindByUser Error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	articles := make([]*model.Article, 0)
	if err = cursor.All(ctx, &articles); err != nil {
		objLog.Errorln("ArticleLogic FindByUser decode error:", err)
		return nil
	}

	for _, article := range articles {
		article.AfterLoad()
	}

	return articles
}

func (self ArticleLogic) SearchMyArticles(ctx context.Context, me *model.Me, sid int, kw string) []map[string]interface{} {
	objLog := GetLogger(ctx)

	filter := bson.M{"author_txt": me.Username}
	if kw != "" {
		filter["title"] = bson.M{"$regex": kw, "$options": "i"}
	}

	findOpts := options.Find().
		SetSort(bson.D{{"_id", -1}}).
		SetLimit(8)

	cursor, err := db.GetCollection("articles").Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("ArticleLogic SearchMyArticles Error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	articles := make([]*model.Article, 0)
	if err = cursor.All(ctx, &articles); err != nil {
		objLog.Errorln("ArticleLogic SearchMyArticles decode error:", err)
		return nil
	}

	for _, article := range articles {
		article.AfterLoad()
	}

	articleIds := slices.StructsIntSlice(articles, "Id")

	saFilter := bson.M{"sid": sid}
	if len(articleIds) > 0 {
		saFilter["article_id"] = bson.M{"$in": articleIds}
	}

	saCursor, err := db.GetCollection("subject_article").Find(ctx, saFilter)
	if err != nil {
		objLog.Errorln("ArticleLogic SearchMyArticles find subject article Error:", err)
		return nil
	}
	defer saCursor.Close(ctx)

	subjectArticles := make([]*model.SubjectArticle, 0)
	if err = saCursor.All(ctx, &subjectArticles); err != nil {
		objLog.Errorln("ArticleLogic SearchMyArticles subject article decode error:", err)
		return nil
	}

	subjectArticleMap := make(map[int]struct{})
	for _, subjectArticle := range subjectArticles {
		subjectArticleMap[subjectArticle.ArticleId] = struct{}{}
	}

	articleMapSlice := make([]map[string]interface{}, len(articles))
	for i, article := range articles {
		articleMap := map[string]interface{}{
			"id":    article.Id,
			"title": article.Title,
		}
		if _, ok := subjectArticleMap[article.Id]; ok {
			articleMap["had_add"] = 1
		} else {
			articleMap["had_add"] = 0
		}

		articleMapSlice[i] = articleMap
	}

	return articleMapSlice
}

// FindAll 支持多页翻看
func (self ArticleLogic) FindAll(ctx context.Context, paginator *Paginator, orderBy string, querystring string, args ...interface{}) []*model.Article {
	objLog := GetLogger(ctx)

	filter := bson.M{}
	if querystring != "" {
		filter = buildFilter(querystring, args...)
	}
	self.addArticleStatusFilter(filter)

	findOpts := options.Find().
		SetSort(buildArticleSort(orderBy)).
		SetSkip(int64(paginator.Offset())).
		SetLimit(int64(paginator.PerPage()))

	cursor, err := db.GetCollection("articles").Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("ArticleLogic FindAll error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	articles := make([]*model.Article, 0)
	if err = cursor.All(ctx, &articles); err != nil {
		objLog.Errorln("ArticleLogic FindAll decode error:", err)
		return nil
	}

	for _, article := range articles {
		article.AfterLoad()
	}

	self.fillUser(articles)

	return articles
}

func (ArticleLogic) Count(ctx context.Context, querystring string, args ...interface{}) int64 {
	objLog := GetLogger(ctx)

	filter := bson.M{"status": bson.M{"$lt": model.ArticleStatusOffline}}
	if querystring != "" {
		extra := buildFilter(querystring, args...)
		for k, v := range extra {
			filter[k] = v
		}
	}

	total, err := db.GetCollection("articles").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("ArticleLogic Count error:", err)
	}

	return total
}

// FindArticleByPage 获取抓取的文章列表（分页）：后台用
func (ArticleLogic) FindArticleByPage(ctx context.Context, conds map[string]string, curPage, limit int) ([]*model.Article, int) {
	objLog := GetLogger(ctx)

	filter := bson.M{}
	for k, v := range conds {
		filter[k] = v
	}

	total, err := db.GetCollection("articles").CountDocuments(ctx, filter)
	if err != nil {
		objLog.Errorln("find count error:", err)
		return nil, 0
	}

	offset := (curPage - 1) * limit
	findOpts := options.Find().
		SetSort(bson.D{{"_id", -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := db.GetCollection("articles").Find(ctx, filter, findOpts)
	if err != nil {
		objLog.Errorln("find error:", err)
		return nil, 0
	}
	defer cursor.Close(ctx)

	articleList := make([]*model.Article, 0)
	if err = cursor.All(ctx, &articleList); err != nil {
		objLog.Errorln("find decode error:", err)
		return nil, 0
	}

	for _, article := range articleList {
		article.AfterLoad()
	}

	return articleList, int(total)
}

// FindByIds 获取多个文章详细信息
func (self ArticleLogic) FindByIds(ids []int) []*model.Article {
	if len(ids) == 0 {
		return nil
	}

	ctx := context.Background()
	cursor, err := db.GetCollection("articles").Find(ctx, bson.M{
		"_id":    bson.M{"$in": ids},
		"status": bson.M{"$lte": model.ArticleStatusOnline},
	})
	if err != nil {
		logger.Errorln("ArticleLogic FindByIds error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	articles := make([]*model.Article, 0)
	if err = cursor.All(ctx, &articles); err != nil {
		logger.Errorln("ArticleLogic FindByIds decode error:", err)
		return nil
	}

	for _, article := range articles {
		article.AfterLoad()
	}

	self.fillUser(articles)

	return articles
}

// MoveToTopic 将该文章移到主题中
func (self ArticleLogic) MoveToTopic(ctx context.Context, id interface{}, me *model.Me) error {
	objLog := GetLogger(ctx)

	idInt := goutils.MustInt(fmt.Sprintf("%v", id))

	article := &model.Article{}
	err := db.GetCollection("articles").FindOne(ctx, bson.M{"_id": idInt}).Decode(article)
	if err != nil {
		objLog.Errorln("ArticleLogic MoveToTopic find article error:", err)
		return err
	}
	article.AfterLoad()

	if !article.IsSelf {
		return errors.New("不是本站发布的文章，不能移动！")
	}

	user := DefaultUser.FindOne(ctx, "username", article.AuthorTxt)

	newTid, err := db.NextID("topics")
	if err != nil {
		objLog.Errorln("ArticleLogic MoveToTopic NextID error:", err)
		return err
	}

	topic := &model.Topic{
		Tid:           newTid,
		Title:         article.Title,
		Content:       article.Content,
		Nid:           6,
		Uid:           user.Uid,
		Lastreplyuid:  article.Lastreplyuid,
		Lastreplytime: article.Lastreplytime,
		EditorUid:     me.Uid,
		Tags:          article.Tags,
		Ctime:         article.Ctime,
	}

	topicEx := &model.TopicEx{
		Tid:   topic.Tid,
		View:  article.Viewnum,
		Reply: article.Cmtnum,
		Like:  article.Likenum,
	}

	session, err := db.GetClient().StartSession()
	if err != nil {
		objLog.Errorln("ArticleLogic MoveToTopic StartSession error:", err)
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		_, insertErr := db.GetCollection("topics").InsertOne(sc, topic)
		if insertErr != nil {
			objLog.Errorln("ArticleLogic MoveToTopic insert Topic error:", insertErr)
			return nil, insertErr
		}

		_, insertErr = db.GetCollection("topics_ex").InsertOne(sc, topicEx)
		if insertErr != nil {
			objLog.Errorln("ArticleLogic MoveToTopic Insert TopicEx error:", insertErr)
			return nil, insertErr
		}

		_, updateErr := db.GetCollection("feed").UpdateMany(sc,
			bson.M{"objid": article.Id, "objtype": model.TypeArticle},
			bson.M{"$set": bson.M{
				"objid":   topic.Tid,
				"objtype": model.TypeTopic,
				"nid":     topic.Nid,
			}})
		if updateErr != nil {
			objLog.Errorln("ArticleLogic MoveToTopic Update Feed error:", updateErr)
			return nil, updateErr
		}

		if article.Cmtnum > 0 {
			_, updateErr = db.GetCollection("comments").UpdateMany(sc,
				bson.M{"objid": article.Id, "objtype": model.TypeArticle},
				bson.M{"$set": bson.M{
					"objid":   topic.Tid,
					"objtype": model.TypeTopic,
				}})
			if updateErr != nil {
				objLog.Errorln("ArticleLogic MoveToTopic Update Comment error:", updateErr)
				return nil, updateErr
			}

			msgCursor, findErr := db.GetCollection("system_message").Find(sc,
				bson.M{"to": user.Uid},
				options.Find().SetLimit(int64(article.Cmtnum)))
			if findErr != nil {
				objLog.Errorln("ArticleLogic MoveToTopic find system message error:", findErr)
				return nil, findErr
			}
			defer msgCursor.Close(sc)

			systemMsgs := make([]*model.SystemMessage, 0)
			if findErr = msgCursor.All(sc, &systemMsgs); findErr != nil {
				objLog.Errorln("ArticleLogic MoveToTopic decode system message error:", findErr)
				return nil, findErr
			}

			for _, msg := range systemMsgs {
				extMap := msg.GetExt()

				if val, ok := extMap["objid"]; ok {
					objid := int(val.(float64))
					if objid != article.Id {
						continue
					}

					extMap["objid"] = topic.Tid
					extMap["objtype"] = model.TypeTopic

					msg.SetExt(extMap)

					_, updateErr = db.GetCollection("system_message").UpdateOne(sc,
						bson.M{"_id": msg.Id},
						bson.M{"$set": bson.M{"ext": msg.Ext}})
					if updateErr != nil {
						objLog.Errorln("ArticleLogic MoveToTopic update system message error:", updateErr)
						return nil, updateErr
					}
				}
			}
		}

		_, delErr := db.GetCollection("articles").DeleteOne(sc, bson.M{"_id": article.Id})
		if delErr != nil {
			objLog.Errorln("ArticleLogic MoveToTopic delete article error:", delErr)
			return nil, delErr
		}

		return nil, nil
	})
	if err != nil {
		objLog.Errorln("ArticleLogic MoveToTopic transaction error:", err)
		return err
	}

	award := -20
	desc := fmt.Sprintf(`你的《%s》并非文章，应该发布到主题中，已被管理员移到主题里 <a href="/topics/%d">%s</a>`, article.Title, topic.Tid, topic.Title)
	DefaultUserRich.IncrUserRich(user, model.MissionTypePunish, award, desc)

	return nil
}

func (self ArticleLogic) transferImage(ctx context.Context, s *goquery.Selection, imgDeny bool, domain string) {
	if v, ok := s.Attr("data-original-src"); ok {
		self.setImgSrc(ctx, v, imgDeny, s, domain)
	} else if v, ok := s.Attr("data-original"); ok {
		self.setImgSrc(ctx, v, imgDeny, s, domain)
	} else if v, ok := s.Attr("data-src"); ok {
		self.setImgSrc(ctx, v, imgDeny, s, domain)
	} else if v, ok := s.Attr("src"); ok {
		self.setImgSrc(ctx, v, imgDeny, s, domain)
	}
}

func (self ArticleLogic) setImgSrc(ctx context.Context, v string, imgDeny bool, s *goquery.Selection, domain string) {
	if imgDeny {
		if strings.HasPrefix(v, "//") {
			v = "https:" + v
		} else if !strings.HasPrefix(v, "http") {
			v = "http://" + domain + v
		}
		path, err := DefaultUploader.TransferUrl(ctx, v)
		if err == nil {
			s.SetAttr("src", global.App.CDNHttps+path)
		} else {
			s.SetAttr("src", v)
		}
	} else {
		s.SetAttr("src", v)
	}
}

func (ArticleLogic) fillUser(articles []*model.Article) {
	usernameSet := set.New(set.NonThreadSafe)
	uidSet := set.New(set.NonThreadSafe)
	for _, article := range articles {
		if article.IsSelf {
			usernameSet.Add(article.Author)
		}

		if article.Lastreplyuid != 0 {
			uidSet.Add(article.Lastreplyuid)
		}
	}
	if !usernameSet.IsEmpty() {
		userMap := DefaultUser.FindUserInfos(nil, set.StringSlice(usernameSet))
		for _, article := range articles {
			if !article.IsSelf {
				continue
			}

			for _, user := range userMap {
				if article.Author == user.Username {
					article.User = user
					break
				}
			}
		}
	}

	if !uidSet.IsEmpty() {
		replyUserMap := DefaultUser.FindUserInfos(nil, set.IntSlice(uidSet))
		for _, article := range articles {
			if article.Lastreplyuid == 0 {
				continue
			}

			article.LastReplyUser = replyUserMap[article.Lastreplyuid]
		}
	}
}

// findByIds 获取多个文章详细信息 包内使用
func (ArticleLogic) findByIds(ids []int) map[int]*model.Article {
	if len(ids) == 0 {
		return nil
	}

	ctx := context.Background()
	cursor, err := db.GetCollection("articles").Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		logger.Errorln("ArticleLogic findByIds error:", err)
		return nil
	}
	defer cursor.Close(ctx)

	articleList := make([]*model.Article, 0)
	if err = cursor.All(ctx, &articleList); err != nil {
		logger.Errorln("ArticleLogic findByIds decode error:", err)
		return nil
	}

	articles := make(map[int]*model.Article, len(articleList))
	for _, a := range articleList {
		a.AfterLoad()
		articles[a.Id] = a
	}
	return articles
}

// FindByIdAndPreNext 获取当前(id)博文以及前后博文
func (ArticleLogic) FindByIdAndPreNext(ctx context.Context, id int) (curArticle *model.Article, prevNext []*model.Article, err error) {
	objLog := GetLogger(ctx)

	if id == 0 {
		err = errors.New("id 不能为0")
		return
	}

	filter := bson.M{
		"_id":    bson.M{"$gte": id - 5, "$lte": id + 5},
		"status": bson.M{"$ne": model.ArticleStatusOffline},
	}

	cursor, findErr := db.GetCollection("articles").Find(ctx, filter)
	if findErr != nil {
		err = findErr
		objLog.Errorln("ArticleLogic FindByIdAndPreNext Error:", err)
		return
	}
	defer cursor.Close(ctx)

	articles := make([]*model.Article, 0)
	if err = cursor.All(ctx, &articles); err != nil {
		objLog.Errorln("ArticleLogic FindByIdAndPreNext decode error:", err)
		return
	}

	if len(articles) == 0 {
		objLog.Errorln("ArticleLogic FindByIdAndPreNext not find articles, id:", id)
		return
	}

	for _, article := range articles {
		article.AfterLoad()
	}

	prevNext = make([]*model.Article, 2)
	prevId, nextId := articles[0].Id, articles[len(articles)-1].Id
	for _, article := range articles {
		if article.Id < id && article.Id > prevId {
			prevId = article.Id
			prevNext[0] = article
		} else if article.Id > id && article.Id < nextId {
			nextId = article.Id
			prevNext[1] = article
		} else if article.Id == id {
			curArticle = article
		}
	}

	if curArticle == nil {
		objLog.Errorln("ArticleLogic FindByIdAndPreNext not find current article, id:", id)
		return
	}

	if prevId == id {
		prevNext[0] = nil
	}

	if nextId == id {
		prevNext[1] = nil
	}

	if curArticle.IsSelf {
		curArticle.User = DefaultUser.FindOne(ctx, "username", curArticle.Author)
	}

	return
}

func (ArticleLogic) FindArticleGCTT(ctx context.Context, article *model.Article) *model.ArticleGCTT {
	articleGCTT := &model.ArticleGCTT{}

	if !article.GCTT {
		return articleGCTT
	}

	objLog := GetLogger(ctx)

	err := db.GetCollection("article_gctt").FindOne(ctx, bson.M{"_id": article.Id}).Decode(articleGCTT)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			objLog.Errorln("ArticleLogic FindArticleGCTT error:", err)
		}
		return articleGCTT
	}

	articleGCTT.AfterLoad()

	if articleGCTT.ArticleID > 0 {
		gcttUser := DefaultGCTT.FindOne(ctx, articleGCTT.Translator)
		articleGCTT.Avatar = gcttUser.Avatar
	}

	return articleGCTT
}

// Modify 修改文章信息
func (ArticleLogic) Modify(ctx context.Context, user *model.Me, form url.Values) (errMsg string, err error) {
	idInt := goutils.MustInt(form.Get("id"))

	article := &model.Article{}
	err = db.GetCollection("articles").FindOne(ctx, bson.M{"_id": idInt}).Decode(article)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = nil
		}
		errMsg = "对不起，服务器内部错误，请稍后再试！"
		return
	}
	article.AfterLoad()

	if !CanEdit(user, article) {
		err = NotModifyAuthorityErr
		return
	}

	form.Set("op_user", user.Username)

	fields := []string{
		"title", "url", "cover", "author", "author_txt",
		"lang", "pub_date", "content",
		"tags", "status", "op_user",
	}
	setDoc := bson.M{}

	for _, field := range fields {
		val := form.Get(field)
		if val != "" {
			setDoc[field] = val
		}
	}

	_, err = db.GetCollection("articles").UpdateOne(ctx, bson.M{"_id": idInt}, bson.M{"$set": setDoc})
	if err != nil {
		logger.Errorf("更新文章 【%d】 信息失败：%s\n", idInt, err)
		errMsg = "对不起，服务器内部错误，请稍后再试！"
		return
	}

	go modifyObservable.NotifyObservers(user.Uid, model.TypeArticle, idInt)

	return
}

// FindById 获取单条博文
func (ArticleLogic) FindById(ctx context.Context, id interface{}) (*model.Article, error) {
	article := &model.Article{}
	idInt := goutils.MustInt(fmt.Sprintf("%v", id))
	err := db.GetCollection("articles").FindOne(ctx, bson.M{"_id": idInt}).Decode(article)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return article, nil
		}
		logger.Errorln("article logic FindById Error:", err)
		return article, err
	}

	article.AfterLoad()

	return article, nil
}

// getOwner 通过objid获得 article 的所有者
func (ArticleLogic) getOwner(id int) int {
	article := &model.Article{}
	ctx := context.Background()
	err := db.GetCollection("articles").FindOne(ctx, bson.M{"_id": id}).Decode(article)
	if err != nil {
		logger.Errorln("article logic getOwner Error:", err)
		return 0
	}
	article.AfterLoad()

	if article.IsSelf {
		user := DefaultUser.FindOne(nil, "username", article.Author)
		return user.Uid
	}
	return 0
}

func (ArticleLogic) addArticleStatusFilter(filter bson.M) {
	filter["status"] = bson.M{"$lt": model.ArticleStatusOffline}
}

// buildArticleSort converts SQL-style orderBy to bson.D for MongoDB sort
func buildArticleSort(orderBy string) bson.D {
	if orderBy == "" {
		return bson.D{{"_id", -1}}
	}
	var sort bson.D
	parts := splitOrderBy(orderBy)
	for _, p := range parts {
		field := p[0]
		if field == "id" {
			field = "_id"
		}
		dir := 1
		if len(p) > 1 && (p[1] == "DESC" || p[1] == "desc") {
			dir = -1
		}
		sort = append(sort, bson.E{Key: field, Value: dir})
	}
	return sort
}

// 博文评论
type ArticleComment struct{}

// UpdateComment 更新该文章的评论信息
// cid：评论id；objid：被评论对象id；uid：评论者；cmttime：评论时间
func (self ArticleComment) UpdateComment(cid, objid, uid int, cmttime time.Time) {
	ctx := context.Background()
	_, err := db.GetCollection("articles").UpdateOne(ctx, bson.M{"_id": objid}, bson.M{
		"$inc": bson.M{"cmtnum": 1},
		"$set": bson.M{
			"lastreplyuid":  uid,
			"lastreplytime": cmttime,
		},
	})
	if err != nil {
		logger.Errorln("更新回复信息失败：", err)
		return
	}
}

func (self ArticleComment) String() string {
	return "article"
}

// SetObjinfo 实现 CommentObjecter 接口
func (self ArticleComment) SetObjinfo(ids []int, commentMap map[int][]*model.Comment) {
	articles := DefaultArticle.FindByIds(ids)
	if len(articles) == 0 {
		return
	}

	for _, article := range articles {
		if article.Status >= model.ArticleStatusOffline {
			continue
		}
		objinfo := make(map[string]interface{})
		objinfo["title"] = article.Title
		objinfo["uri"] = model.PathUrlMap[model.TypeArticle]
		objinfo["type_name"] = model.TypeNameMap[model.TypeArticle]

		for _, comment := range commentMap[article.Id] {
			comment.Objinfo = objinfo
		}
	}
}

// 博文喜欢
type ArticleLike struct{}

// UpdateLike 更新该文章的喜欢数
// objid：被喜欢对象id；num: 喜欢数(负数表示取消喜欢)
func (self ArticleLike) UpdateLike(objid, num int) {
	ctx := context.Background()
	_, err := db.GetCollection("articles").UpdateOne(ctx, bson.M{"_id": objid}, bson.M{
		"$inc": bson.M{"likenum": num},
	})
	if err != nil {
		logger.Errorln("更新文章喜欢数失败：", err)
	}
}

func (self ArticleLike) String() string {
	return "article"
}

func (ArticleLogic) Delete(ctx context.Context, id int, username string, isRoot bool) error {
	article := &model.Article{}
	err := db.GetCollection("articles").FindOne(ctx, bson.M{"_id": id}).Decode(article)
	if err != nil {
		return errors.New("文章不存在")
	}
	if article.Author != username && article.OpUser != username && !isRoot {
		return errors.New("无权删除")
	}
	_, err = db.GetCollection("articles").DeleteOne(ctx, bson.M{"_id": id})
	return err
}
