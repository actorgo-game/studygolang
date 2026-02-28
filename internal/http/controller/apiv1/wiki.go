package apiv1

import (
	"github.com/studygolang/studygolang/context"
	"github.com/studygolang/studygolang/internal/logic"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type WikiController struct{}

func (self WikiController) RegisterRoute(g *echo.Group) {
	g.GET("/wiki", self.ReadList)
	g.GET("/wiki/:uri", self.Detail)
	g.POST("/wiki/new", self.Create)
	g.POST("/wiki/modify", self.Modify)
}

func (WikiController) ReadList(ctx echo.Context) error {
	limit := goutils.MustInt(ctx.QueryParam("limit"), perPage)
	lastId := goutils.MustInt(ctx.QueryParam("last_id"))
	var wikis interface{}
	if lastId > 0 {
		wikis = logic.DefaultWiki.FindBy(context.EchoContext(ctx), limit, lastId)
	} else {
		wikis = logic.DefaultWiki.FindBy(context.EchoContext(ctx), limit)
	}
	return success(ctx, map[string]interface{}{
		"list": wikis,
	})
}

func (WikiController) Detail(ctx echo.Context) error {
	uri := ctx.Param("uri")
	wiki := logic.DefaultWiki.FindOne(context.EchoContext(ctx), uri)
	if wiki == nil || wiki.Id == 0 {
		return fail(ctx, "Wiki不存在")
	}
	return success(ctx, map[string]interface{}{"wiki": wiki})
}

func (WikiController) Create(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	err := logic.DefaultWiki.Create(context.EchoContext(ctx), meVal, ctx.Request().Form)
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
	err := logic.DefaultWiki.Modify(context.EchoContext(ctx), meVal, ctx.Request().Form)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}
