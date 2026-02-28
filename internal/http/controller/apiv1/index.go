package apiv1

import (
	"github.com/studygolang/studygolang/context"
	"github.com/studygolang/studygolang/internal/logic"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type IndexController struct{}

func (self IndexController) RegisterRoute(g *echo.Group) {
	g.GET("/home", self.Home)
}

func (IndexController) Home(ctx echo.Context) error {
	tab := ctx.QueryParam("tab")
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	paginator := logic.NewPaginator(curPage)
	data := logic.DefaultIndex.FindData(context.EchoContext(ctx), tab, paginator)
	return success(ctx, data)
}
