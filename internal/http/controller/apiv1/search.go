package apiv1

import (
	"fmt"
	"strings"

	stdctx "context"

	"github.com/studygolang/studygolang/context"
	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/logic"
	"github.com/studygolang/studygolang/internal/model"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SearchController struct{}

func (self SearchController) RegisterRoute(g *echo.Group) {
	g.GET("/search", self.Search)
}

func (SearchController) Search(ctx echo.Context) error {
	q := ctx.QueryParam("q")
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	if q == "" {
		return fail(ctx, "请输入搜索关键词")
	}

	field := ctx.QueryParam("field")
	start := (curPage - 1) * perPage
	result, err := logic.DefaultSearcher.DoSearch(q, field, start, perPage)
	if err == nil && result != nil && result.NumFound > 0 {
		list := make([]map[string]interface{}, 0, len(result.Docs))
		for _, doc := range result.Docs {
			list = append(list, map[string]interface{}{
				"title":   doc.HlTitle,
				"content": doc.HlContent,
				"url":     buildSearchURL(doc),
			})
		}
		return success(ctx, map[string]interface{}{"list": list, "total": result.NumFound})
	}

	list := mongoSearch(context.EchoContext(ctx), q, curPage, perPage)
	return success(ctx, map[string]interface{}{"list": list})
}

func buildSearchURL(doc *model.Document) string {
	switch doc.Objtype {
	case model.TypeTopic:
		return fmt.Sprintf("/topics/%d", doc.Objid)
	case model.TypeArticle:
		return fmt.Sprintf("/articles/%d", doc.Objid)
	case model.TypeResource:
		return fmt.Sprintf("/resources/%d", doc.Objid)
	case model.TypeProject:
		return fmt.Sprintf("/p/%d", doc.Objid)
	case model.TypeWiki:
		return fmt.Sprintf("/wiki/%d", doc.Objid)
	case model.TypeBook:
		return fmt.Sprintf("/book/%d", doc.Objid)
	default:
		return "#"
	}
}

func mongoSearch(ctx stdctx.Context, q string, page, limit int) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)
	escaped := primitive.Regex{Pattern: escapeRegex(q), Options: "i"}
	skip := int64((page - 1) * limit)
	each := int64(limit)

	topicFilter := bson.M{"title": bson.M{"$regex": escaped}, "flag": bson.M{"$lt": model.FlagAuditDelete}}
	cursor, err := db.GetCollection("topics").Find(ctx, topicFilter, options.Find().SetSort(bson.M{"_id": -1}).SetSkip(skip).SetLimit(each).SetProjection(bson.M{"_id": 1, "title": 1, "content": 1}))
	if err == nil {
		defer cursor.Close(ctx)
		var topics []struct {
			Tid     int    `bson:"_id"`
			Title   string `bson:"title"`
			Content string `bson:"content"`
		}
		if cursor.All(ctx, &topics) == nil {
			for _, t := range topics {
				results = append(results, map[string]interface{}{
					"title":   t.Title,
					"content": truncate(t.Content, 200),
					"url":     fmt.Sprintf("/topics/%d", t.Tid),
				})
			}
		}
	}

	articleFilter := bson.M{"title": bson.M{"$regex": escaped}}
	cursor2, err := db.GetCollection("articles").Find(ctx, articleFilter, options.Find().SetSort(bson.M{"_id": -1}).SetSkip(skip).SetLimit(each).SetProjection(bson.M{"_id": 1, "title": 1, "txt": 1}))
	if err == nil {
		defer cursor2.Close(ctx)
		var articles []struct {
			Id    int    `bson:"_id"`
			Title string `bson:"title"`
			Txt   string `bson:"txt"`
		}
		if cursor2.All(ctx, &articles) == nil {
			for _, a := range articles {
				results = append(results, map[string]interface{}{
					"title":   a.Title,
					"content": truncate(a.Txt, 200),
					"url":     fmt.Sprintf("/articles/%d", a.Id),
				})
			}
		}
	}

	resourceFilter := bson.M{"title": bson.M{"$regex": escaped}}
	cursor3, err := db.GetCollection("resource").Find(ctx, resourceFilter, options.Find().SetSort(bson.M{"_id": -1}).SetSkip(skip).SetLimit(each).SetProjection(bson.M{"_id": 1, "title": 1, "content": 1}))
	if err == nil {
		defer cursor3.Close(ctx)
		var resources []struct {
			Id      int    `bson:"_id"`
			Title   string `bson:"title"`
			Content string `bson:"content"`
		}
		if cursor3.All(ctx, &resources) == nil {
			for _, r := range resources {
				results = append(results, map[string]interface{}{
					"title":   r.Title,
					"content": truncate(r.Content, 200),
					"url":     fmt.Sprintf("/resources/%d", r.Id),
				})
			}
		}
	}

	projectFilter := bson.M{"$or": []bson.M{{"name": bson.M{"$regex": escaped}}, {"category": bson.M{"$regex": escaped}}}}
	cursor4, err := db.GetCollection("open_project").Find(ctx, projectFilter, options.Find().SetSort(bson.M{"_id": -1}).SetSkip(skip).SetLimit(each).SetProjection(bson.M{"_id": 1, "name": 1, "category": 1, "uri": 1, "desc": 1}))
	if err == nil {
		defer cursor4.Close(ctx)
		var projects []struct {
			Id       int    `bson:"_id"`
			Name     string `bson:"name"`
			Category string `bson:"category"`
			Uri      string `bson:"uri"`
			Desc     string `bson:"desc"`
		}
		if cursor4.All(ctx, &projects) == nil {
			for _, p := range projects {
				uri := p.Uri
				if uri == "" {
					uri = fmt.Sprintf("%d", p.Id)
				}
				results = append(results, map[string]interface{}{
					"title":   p.Category + p.Name,
					"content": truncate(p.Desc, 200),
					"url":     fmt.Sprintf("/p/%s", uri),
				})
			}
		}
	}

	return results
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func escapeRegex(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`, `.`, `\.`, `*`, `\*`, `+`, `\+`, `?`, `\?`,
		`(`, `\(`, `)`, `\)`, `[`, `\[`, `]`, `\]`, `{`, `\{`,
		`}`, `\}`, `^`, `\^`, `$`, `\$`, `|`, `\|`,
	)
	return replacer.Replace(s)
}
