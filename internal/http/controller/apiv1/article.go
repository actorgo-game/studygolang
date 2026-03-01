package apiv1

import (
	"github.com/studygolang/studygolang/context"
	. "github.com/studygolang/studygolang/internal/http"
	"github.com/studygolang/studygolang/internal/logic"
	"github.com/studygolang/studygolang/internal/model"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type ArticleController struct{}

func (self ArticleController) RegisterRoute(g *echo.Group) {
	g.GET("/articles", self.ReadList)
	g.GET("/article/detail", self.Detail)
	g.POST("/articles/new", self.Create)
	g.POST("/articles/modify", self.Modify)
	g.POST("/articles/delete", self.Delete)
}

func (ArticleController) ReadList(ctx echo.Context) error {
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)
	articles := logic.DefaultArticle.FindAll(context.EchoContext(ctx), paginator, "", "")
	total := logic.DefaultArticle.Count(context.EchoContext(ctx), "")
	return success(ctx, map[string]interface{}{
		"list":     articles,
		"total":    total,
		"page":     curPage,
		"per_page": perPage,
	})
}

func (ArticleController) Detail(ctx echo.Context) error {
	id := goutils.MustInt(ctx.QueryParam("id"))
	article, err := logic.DefaultArticle.FindById(context.EchoContext(ctx), id)
	if err != nil {
		return fail(ctx, "文章不存在")
	}
	if article.Id == 0 {
		return fail(ctx, "文章不存在")
	}
	logic.Views.Incr(Request(ctx), model.TypeArticle, id)
	return success(ctx, map[string]interface{}{"article": article})
}

func (ArticleController) Create(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	formParams, _ := ctx.FormParams()
	_, err := logic.DefaultArticle.Publish(context.EchoContext(ctx), meVal, formParams)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func (ArticleController) Delete(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	id := goutils.MustInt(ctx.FormValue("id"))
	err := logic.DefaultArticle.Delete(context.EchoContext(ctx), id, meVal.Username, meVal.IsRoot)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func (ArticleController) Modify(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	formParams, _ := ctx.FormParams()
	errMsg, err := logic.DefaultArticle.Modify(context.EchoContext(ctx), meVal, formParams)
	if err != nil {
		return fail(ctx, errMsg)
	}
	return success(ctx, nil)
}
