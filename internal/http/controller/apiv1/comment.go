package apiv1

import (
	"net/url"

	"github.com/studygolang/studygolang/context"
	"github.com/studygolang/studygolang/internal/logic"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type CommentController struct{}

func (self CommentController) RegisterRoute(g *echo.Group) {
	g.GET("/object/comments", self.CommentList)
	g.POST("/comment/:objid", self.Create)
	g.POST("/comment/delete", self.Delete)
	g.GET("/at/users", self.AtUsers)
}

func (CommentController) CommentList(ctx echo.Context) error {
	objid := goutils.MustInt(ctx.QueryParam("objid"))
	objtype := goutils.MustInt(ctx.QueryParam("objtype"))
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)
	comments := logic.DefaultComment.FindAll(context.EchoContext(ctx), paginator, "", "objid=? AND objtype=?", objid, objtype)
	total := logic.DefaultComment.Count(context.EchoContext(ctx), "objid=? AND objtype=?", objid, objtype)
	enriched := logic.DefaultComment.EnrichWithUsers(context.EchoContext(ctx), comments)
	return success(ctx, map[string]interface{}{
		"list":     enriched,
		"total":    total,
		"page":     curPage,
		"per_page": perPage,
	})
}

func (CommentController) Create(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	objid := goutils.MustInt(ctx.Param("objid"))
	content := ctx.FormValue("content")
	if content == "" {
		return fail(ctx, "评论内容不能为空")
	}
	form := url.Values{}
	form.Set("objtype", ctx.FormValue("objtype"))
	form.Set("content", content)
	_, err := logic.DefaultComment.Publish(context.EchoContext(ctx), meVal.Uid, objid, form)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func (CommentController) Delete(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	cid := goutils.MustInt(ctx.FormValue("cid"))
	err := logic.DefaultComment.Delete(context.EchoContext(ctx), cid, meVal.Uid, meVal.IsRoot)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func (CommentController) AtUsers(ctx echo.Context) error {
	term := ctx.QueryParam("term")
	users := logic.DefaultUser.GetUserMentions(term, 10, false)
	return success(ctx, users)
}
