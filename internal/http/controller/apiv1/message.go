package apiv1

import (
	"github.com/studygolang/studygolang/context"
	"github.com/studygolang/studygolang/internal/logic"
	"github.com/studygolang/studygolang/internal/model"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type MessageController struct{}

func (self MessageController) RegisterRoute(g *echo.Group) {
	g.GET("/message/system", self.SysMsgList)
	g.GET("/message/inbox", self.InboxList)
	g.GET("/message/outbox", self.OutboxList)
	g.POST("/message/send", self.Send)
	g.POST("/message/delete", self.Delete)
}

func (MessageController) SysMsgList(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)
	messages := logic.DefaultMessage.FindSysMsgsByUid(context.EchoContext(ctx), meVal.Uid, paginator)
	return success(ctx, map[string]interface{}{"list": messages, "page": curPage})
}

func (MessageController) InboxList(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)
	messages := logic.DefaultMessage.FindToMsgsByUid(context.EchoContext(ctx), meVal.Uid, paginator)
	return success(ctx, map[string]interface{}{"list": messages, "page": curPage})
}

func (MessageController) OutboxList(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)
	messages := logic.DefaultMessage.FindFromMsgsByUid(context.EchoContext(ctx), meVal.Uid, paginator)
	return success(ctx, map[string]interface{}{"list": messages, "page": curPage})
}

func (MessageController) Send(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	to := goutils.MustInt(ctx.FormValue("to"))
	content := ctx.FormValue("content")
	ok = logic.DefaultMessage.SendMessageTo(context.EchoContext(ctx), meVal.Uid, to, content)
	if !ok {
		return fail(ctx, "发送失败")
	}
	return success(ctx, nil)
}

func (MessageController) Delete(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	id := ctx.FormValue("id")
	msgtype := ctx.FormValue("msgtype")
	ok = logic.DefaultMessage.DeleteMessage(context.EchoContext(ctx), id, msgtype)
	if !ok {
		return fail(ctx, "删除失败")
	}
	return success(ctx, nil)
}
