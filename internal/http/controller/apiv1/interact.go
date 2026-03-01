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
	g.GET("/like/:objid", self.HadLike)
	g.POST("/favorite/:objid", self.Favorite)
	g.GET("/favorite/:objid", self.HadFavorite)
}

func (InteractController) Like(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	objid := goutils.MustInt(ctx.Param("objid"))
	objtype := goutils.MustInt(ctx.FormValue("objtype"))

	had := logic.DefaultLike.HadLike(context.EchoContext(ctx), meVal.Uid, objid, objtype)
	if had == model.FlagLike {
		err := logic.DefaultLike.LikeObject(context.EchoContext(ctx), meVal.Uid, objid, objtype, model.FlagCancel)
		if err != nil {
			return fail(ctx, err.Error())
		}
		return success(ctx, map[string]interface{}{"liked": false})
	}

	err := logic.DefaultLike.LikeObject(context.EchoContext(ctx), meVal.Uid, objid, objtype, model.FlagLike)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, map[string]interface{}{"liked": true})
}

func (InteractController) HadLike(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return success(ctx, map[string]interface{}{"liked": false})
	}
	objid := goutils.MustInt(ctx.Param("objid"))
	objtype := goutils.MustInt(ctx.QueryParam("objtype"))
	had := logic.DefaultLike.HadLike(context.EchoContext(ctx), meVal.Uid, objid, objtype)
	return success(ctx, map[string]interface{}{"liked": had == model.FlagLike})
}

func (InteractController) Favorite(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	objid := goutils.MustInt(ctx.Param("objid"))
	objtype := goutils.MustInt(ctx.FormValue("objtype"))

	had := logic.DefaultFavorite.HadFavorite(context.EchoContext(ctx), meVal.Uid, objid, objtype)
	if had == 1 {
		err := logic.DefaultFavorite.Cancel(context.EchoContext(ctx), meVal.Uid, objid, objtype)
		if err != nil {
			return fail(ctx, err.Error())
		}
		return success(ctx, map[string]interface{}{"favorited": false})
	}

	err := logic.DefaultFavorite.Save(context.EchoContext(ctx), meVal.Uid, objid, objtype)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, map[string]interface{}{"favorited": true})
}

func (InteractController) HadFavorite(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return success(ctx, map[string]interface{}{"favorited": false})
	}
	objid := goutils.MustInt(ctx.Param("objid"))
	objtype := goutils.MustInt(ctx.QueryParam("objtype"))
	had := logic.DefaultFavorite.HadFavorite(context.EchoContext(ctx), meVal.Uid, objid, objtype)
	return success(ctx, map[string]interface{}{"favorited": had == 1})
}
