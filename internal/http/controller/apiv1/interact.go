package apiv1

import (
	"github.com/studygolang/studygolang/context"
	"github.com/studygolang/studygolang/internal/logic"
	"github.com/studygolang/studygolang/internal/model"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type InteractController struct{}

func (self InteractController) RegisterRoute(g *echo.Group) {
	g.POST("/like/:objid", self.Like)
	g.POST("/favorite/:objid", self.Favorite)
}

func (InteractController) Like(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	objid := goutils.MustInt(ctx.Param("objid"))
	objtype := goutils.MustInt(ctx.FormValue("objtype"))
	likeFlag := goutils.MustInt(ctx.FormValue("flag"), model.FlagLike)
	err := logic.DefaultLike.LikeObject(context.EchoContext(ctx), meVal.Uid, objid, objtype, likeFlag)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func (InteractController) Favorite(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	objid := goutils.MustInt(ctx.Param("objid"))
	objtype := goutils.MustInt(ctx.FormValue("objtype"))
	err := logic.DefaultFavorite.Save(context.EchoContext(ctx), meVal.Uid, objid, objtype)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}
