package apiv1

import (
	"github.com/studygolang/studygolang/context"
	"github.com/studygolang/studygolang/internal/logic"
	"github.com/studygolang/studygolang/internal/model"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type MiscController struct{}

func (self MiscController) RegisterRoute(g *echo.Group) {
	g.GET("/mission/daily", self.DailyMission)
	g.GET("/mission/daily/redeem", self.RedeemDaily)
	g.GET("/balance", self.Balance)
	g.GET("/gift", self.GiftList)
	g.POST("/gift/exchange", self.ExchangeGift)
	g.GET("/gift/mine", self.MyGifts)
	g.GET("/top/dau", self.DauRank)
	g.GET("/top/rich", self.RichRank)
}

func (MiscController) DailyMission(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	hasLogin := logic.DefaultMission.HasLoginMission(context.EchoContext(ctx), meVal)
	return success(ctx, map[string]interface{}{"has_login_mission": hasLogin})
}

func (MiscController) RedeemDaily(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	err := logic.DefaultMission.RedeemLoginAward(context.EchoContext(ctx), meVal)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func (MiscController) Balance(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	records := logic.DefaultUserRich.FindBalanceDetail(context.EchoContext(ctx), meVal, curPage)
	return success(ctx, map[string]interface{}{"balance": meVal.Balance, "records": records})
}

func (MiscController) GiftList(ctx echo.Context) error {
	gifts := logic.DefaultGift.FindAllOnline(context.EchoContext(ctx))
	return success(ctx, gifts)
}

func (MiscController) ExchangeGift(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	giftId := goutils.MustInt(ctx.FormValue("gift_id"))
	err := logic.DefaultGift.Exchange(context.EchoContext(ctx), meVal, giftId)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func (MiscController) MyGifts(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	records := logic.DefaultGift.FindExchangeRecords(context.EchoContext(ctx), meVal)
	return success(ctx, records)
}

func (MiscController) DauRank(ctx echo.Context) error {
	num := goutils.MustInt(ctx.QueryParam("limit"), 20)
	data := logic.DefaultRank.FindDAURank(context.EchoContext(ctx), num)
	return success(ctx, data)
}

func (MiscController) RichRank(ctx echo.Context) error {
	data := logic.DefaultRank.FindRichRank(context.EchoContext(ctx))
	return success(ctx, data)
}
