package apiv1

import (
	"github.com/studygolang/studygolang/context"
	"github.com/studygolang/studygolang/internal/logic"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type UserController struct{}

func (self UserController) RegisterRoute(g *echo.Group) {
	g.GET("/user/:username", self.Profile)
	g.GET("/user/:username/topics", self.Topics)
	g.GET("/user/:username/articles", self.Articles)
	g.GET("/user/:username/resources", self.Resources)
	g.GET("/user/:username/projects", self.Projects)
	g.GET("/users/active", self.ActiveUsers)
	g.GET("/users/newest", self.NewestUsers)
	g.POST("/user/modify", self.Modify)
}

func (UserController) Profile(ctx echo.Context) error {
	username := ctx.Param("username")
	user := logic.DefaultUser.FindOne(context.EchoContext(ctx), "username", username)
	if user == nil || user.Uid == 0 {
		return fail(ctx, "用户不存在")
	}
	return success(ctx, map[string]interface{}{"user": user})
}

func (UserController) Topics(ctx echo.Context) error {
	username := ctx.Param("username")
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	user := logic.DefaultUser.FindOne(context.EchoContext(ctx), "username", username)
	if user == nil || user.Uid == 0 {
		return fail(ctx, "用户不存在")
	}
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)
	topics := logic.DefaultTopic.FindAll(context.EchoContext(ctx), paginator, "topics", "uid=?", user.Uid)
	return success(ctx, map[string]interface{}{"list": topics, "page": curPage})
}

func (UserController) Articles(ctx echo.Context) error {
	username := ctx.Param("username")
	user := logic.DefaultUser.FindOne(context.EchoContext(ctx), "username", username)
	if user == nil || user.Uid == 0 {
		return fail(ctx, "用户不存在")
	}
	articles := logic.DefaultArticle.FindByUser(context.EchoContext(ctx), username, perPage)
	return success(ctx, map[string]interface{}{"list": articles})
}

func (UserController) Resources(ctx echo.Context) error {
	username := ctx.Param("username")
	user := logic.DefaultUser.FindOne(context.EchoContext(ctx), "username", username)
	if user == nil || user.Uid == 0 {
		return fail(ctx, "用户不存在")
	}
	resources := logic.DefaultResource.FindRecent(context.EchoContext(ctx), user.Uid)
	return success(ctx, map[string]interface{}{"list": resources})
}

func (UserController) Projects(ctx echo.Context) error {
	username := ctx.Param("username")
	user := logic.DefaultUser.FindOne(context.EchoContext(ctx), "username", username)
	if user == nil || user.Uid == 0 {
		return fail(ctx, "用户不存在")
	}
	projects := logic.DefaultProject.FindRecent(context.EchoContext(ctx), username)
	return success(ctx, map[string]interface{}{"list": projects})
}

func (UserController) ActiveUsers(ctx echo.Context) error {
	limit := goutils.MustInt(ctx.QueryParam("limit"), 12)
	users := logic.DefaultUser.FindActiveUsers(context.EchoContext(ctx), limit)
	return success(ctx, users)
}

func (UserController) NewestUsers(ctx echo.Context) error {
	limit := goutils.MustInt(ctx.QueryParam("limit"), 12)
	users := logic.DefaultUser.FindNewUsers(context.EchoContext(ctx), limit)
	return success(ctx, users)
}

func (UserController) Modify(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	errMsg, err := logic.DefaultUser.Update(context.EchoContext(ctx), meVal, ctx.Request().Form)
	if err != nil {
		return fail(ctx, errMsg)
	}
	return success(ctx, nil)
}
