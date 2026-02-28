package apiv1

import (
	"github.com/studygolang/studygolang/context"
	"github.com/studygolang/studygolang/internal/logic"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type ReadingController struct{}

func (self ReadingController) RegisterRoute(g *echo.Group) {
	g.GET("/readings", self.ReadList)
	g.GET("/reading/:id", self.Detail)
}

func (ReadingController) ReadList(ctx echo.Context) error {
	limit := goutils.MustInt(ctx.QueryParam("limit"), perPage)
	lastId := goutils.MustInt(ctx.QueryParam("last_id"))
	rtype := goutils.MustInt(ctx.QueryParam("rtype"))
	var readings interface{}
	if lastId > 0 {
		readings = logic.DefaultReading.FindBy(context.EchoContext(ctx), limit, rtype, lastId)
	} else {
		readings = logic.DefaultReading.FindBy(context.EchoContext(ctx), limit, rtype)
	}
	return success(ctx, map[string]interface{}{
		"list": readings,
	})
}

func (ReadingController) Detail(ctx echo.Context) error {
	id := goutils.MustInt(ctx.Param("id"))
	reading := logic.DefaultReading.FindById(context.EchoContext(ctx), id)
	if reading == nil || reading.Id == 0 {
		return fail(ctx, "晨读不存在")
	}
	return success(ctx, map[string]interface{}{"reading": reading})
}
