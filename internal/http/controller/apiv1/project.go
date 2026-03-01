package apiv1

import (
	"github.com/studygolang/studygolang/context"
	. "github.com/studygolang/studygolang/internal/http"
	"github.com/studygolang/studygolang/internal/logic"
	"github.com/studygolang/studygolang/internal/model"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type ProjectController struct{}

func (self ProjectController) RegisterRoute(g *echo.Group) {
	g.GET("/projects", self.ReadList)
	g.GET("/project/detail", self.Detail)
	g.POST("/project/new", self.Create)
	g.POST("/project/delete", self.Delete)
}

func (ProjectController) ReadList(ctx echo.Context) error {
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)
	projects := logic.DefaultProject.FindAll(context.EchoContext(ctx), paginator, "", "")
	total := logic.DefaultProject.Count(context.EchoContext(ctx), "")
	return success(ctx, map[string]interface{}{
		"list":     projects,
		"total":    total,
		"page":     curPage,
		"per_page": perPage,
	})
}

func (ProjectController) Detail(ctx echo.Context) error {
	uri := ctx.QueryParam("uri")
	project := logic.DefaultProject.FindOne(context.EchoContext(ctx), uri)
	if project == nil {
		return fail(ctx, "项目不存在")
	}
	logic.Views.Incr(Request(ctx), model.TypeProject, project.Id)
	return success(ctx, map[string]interface{}{"project": project})
}

func (ProjectController) Delete(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	id := goutils.MustInt(ctx.FormValue("id"))
	err := logic.DefaultProject.Delete(context.EchoContext(ctx), id, meVal.Username, meVal.IsRoot)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func (ProjectController) Create(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	err := logic.DefaultProject.Publish(context.EchoContext(ctx), meVal, ctx.Request().Form)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}
