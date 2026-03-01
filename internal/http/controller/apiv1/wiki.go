package apiv1

import (
	"github.com/studygolang/studygolang/context"
	"github.com/studygolang/studygolang/internal/logic"
	"github.com/studygolang/studygolang/internal/model"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type WikiController struct{}

func (self WikiController) RegisterRoute(g *echo.Group) {
	g.GET("/wiki", self.ReadList)
	g.GET("/wiki/:uri", self.Detail)
	g.POST("/wiki/new", self.Create)
	g.POST("/wiki/modify", self.Modify)
	g.POST("/wiki/delete", self.Delete)
}

func (WikiController) ReadList(ctx echo.Context) error {
	limit := goutils.MustInt(ctx.QueryParam("limit"), perPage)
	lastId := goutils.MustInt(ctx.QueryParam("last_id"))
	var wikiList []*model.Wiki
	if lastId > 0 {
		wikiList = logic.DefaultWiki.FindBy(context.EchoContext(ctx), limit, lastId)
	} else {
		wikiList = logic.DefaultWiki.FindBy(context.EchoContext(ctx), limit)
	}
	result := make([]map[string]interface{}, 0, len(wikiList))
	for _, w := range wikiList {
		item := map[string]interface{}{
			"id":      w.Id,
			"title":   w.Title,
			"uri":     w.Uri,
			"uid":     w.Uid,
			"viewnum": w.Viewnum,
			"tags":    w.Tags,
			"ctime":   w.Ctime,
		}
		if w.Users != nil {
			if u, ok := w.Users[w.Uid]; ok {
				item["user"] = u
			}
		}
		result = append(result, item)
	}
	total := logic.DefaultWiki.Total()
	return success(ctx, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

func (WikiController) Detail(ctx echo.Context) error {
	uri := ctx.Param("uri")
	wiki := logic.DefaultWiki.FindOne(context.EchoContext(ctx), uri)
	if wiki == nil || wiki.Id == 0 {
		if id := goutils.MustInt(uri); id > 0 {
			wiki = logic.DefaultWiki.FindById(context.EchoContext(ctx), id)
		}
	}
	if wiki == nil || wiki.Id == 0 {
		return fail(ctx, "Wiki不存在")
	}

	result := map[string]interface{}{"wiki": wiki}
	if wiki.Uid > 0 {
		userMap := logic.DefaultUser.FindUserInfos(context.EchoContext(ctx), []int{wiki.Uid})
		if u, ok := userMap[wiki.Uid]; ok {
			result["wiki_user"] = u
		}
	}
	return success(ctx, result)
}

func (WikiController) Create(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	formParams, err := ctx.FormParams()
	if err != nil {
		return fail(ctx, "参数解析失败")
	}
	err = logic.DefaultWiki.Create(context.EchoContext(ctx), meVal, formParams)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func (WikiController) Modify(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	formParams, err := ctx.FormParams()
	if err != nil {
		return fail(ctx, "参数解析失败")
	}
	err = logic.DefaultWiki.Modify(context.EchoContext(ctx), meVal, formParams)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func (WikiController) Delete(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	id := goutils.MustInt(ctx.FormValue("id"))
	err := logic.DefaultWiki.Delete(context.EchoContext(ctx), id, meVal.Uid, meVal.IsRoot)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}
