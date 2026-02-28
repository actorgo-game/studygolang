package apiv1

import (
	"github.com/studygolang/studygolang/context"
	. "github.com/studygolang/studygolang/internal/http"
	"github.com/studygolang/studygolang/internal/logic"
	"github.com/studygolang/studygolang/internal/model"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type ResourceController struct{}

func (self ResourceController) RegisterRoute(g *echo.Group) {
	g.GET("/resources", self.ReadList)
	g.GET("/resource/detail", self.Detail)
	g.POST("/resources/new", self.Create)
}

func (ResourceController) ReadList(ctx echo.Context) error {
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)
	resources, total := logic.DefaultResource.FindAll(context.EchoContext(ctx), paginator, "resources", "")
	return success(ctx, map[string]interface{}{
		"list":     resources,
		"total":    total,
		"page":     curPage,
		"per_page": perPage,
	})
}

func (ResourceController) Detail(ctx echo.Context) error {
	id := goutils.MustInt(ctx.QueryParam("id"))
	resource, _ := logic.DefaultResource.FindById(context.EchoContext(ctx), id)
	if resource == nil || len(resource) == 0 {
		return fail(ctx, "资源不存在")
	}
	logic.Views.Incr(Request(ctx), model.TypeResource, id)
	return success(ctx, map[string]interface{}{"resource": resource})
}

func (ResourceController) Create(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	err := logic.DefaultResource.Publish(context.EchoContext(ctx), meVal, ctx.Request().Form)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}
