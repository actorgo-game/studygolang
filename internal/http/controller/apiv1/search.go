package apiv1

import (
	"github.com/studygolang/studygolang/internal/logic"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type SearchController struct{}

func (self SearchController) RegisterRoute(g *echo.Group) {
	g.GET("/search", self.Search)
}

func (SearchController) Search(ctx echo.Context) error {
	q := ctx.QueryParam("q")
	field := ctx.QueryParam("field")
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	if q == "" {
		return fail(ctx, "请输入搜索关键词")
	}
	start := (curPage - 1) * perPage
	result, err := logic.DefaultSearcher.DoSearch(q, field, start, perPage)
	if err != nil {
		return fail(ctx, "搜索出错")
	}
	return success(ctx, result)
}
