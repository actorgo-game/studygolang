package apiv1

import (
	"github.com/studygolang/studygolang/context"
	. "github.com/studygolang/studygolang/internal/http"
	"github.com/studygolang/studygolang/internal/logic"
	"github.com/studygolang/studygolang/internal/model"

	echo "github.com/labstack/echo/v4"
)

type AccountController struct{}

func (self AccountController) RegisterRoute(g *echo.Group) {
	g.POST("/account/login", self.Login)
	g.POST("/account/register", self.Register)
	g.GET("/account/logout", self.Logout)
	g.GET("/user/current", self.CurrentUser)
	g.POST("/account/changepwd", self.ChangePwd)
}

func (AccountController) Login(ctx echo.Context) error {
	username := ctx.FormValue("username")
	passwd := ctx.FormValue("passwd")

	userLogin, err := logic.DefaultUser.Login(context.EchoContext(ctx), username, passwd)
	if err != nil {
		return fail(ctx, err.Error())
	}

	SetLoginCookie(ctx, userLogin.Username)

	user := logic.DefaultUser.FindCurrentUser(context.EchoContext(ctx), userLogin.Username)
	return success(ctx, user)
}

func (AccountController) Register(ctx echo.Context) error {
	errMsg, err := logic.DefaultUser.CreateUser(context.EchoContext(ctx), ctx.Request().Form)
	if err != nil {
		return fail(ctx, errMsg)
	}

	SetLoginCookie(ctx, ctx.FormValue("username"))
	return success(ctx, nil)
}

func (AccountController) Logout(ctx echo.Context) error {
	session := GetCookieSession(ctx)
	session.Options.MaxAge = -1
	session.Save(Request(ctx), ResponseWriter(ctx))
	return success(ctx, nil)
}

func (AccountController) CurrentUser(ctx echo.Context) error {
	me, ok := ctx.Get("user").(*model.Me)
	if !ok || me.Uid == 0 {
		return success(ctx, nil)
	}
	return success(ctx, me)
}

func (AccountController) ChangePwd(ctx echo.Context) error {
	curPasswd := ctx.FormValue("cur_passwd")
	passwd := ctx.FormValue("passwd")
	me, ok := ctx.Get("user").(*model.Me)
	if !ok || me.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	errMsg, err := logic.DefaultUser.UpdatePasswd(context.EchoContext(ctx), me.Username, curPasswd, passwd)
	if err != nil {
		return fail(ctx, errMsg)
	}
	return success(ctx, nil)
}
