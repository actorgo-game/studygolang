package apiv1

import (
	"github.com/studygolang/studygolang/context"
	. "github.com/studygolang/studygolang/internal/http"
	"github.com/studygolang/studygolang/internal/logic"
	"github.com/studygolang/studygolang/internal/model"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type TopicController struct{}

func (self TopicController) RegisterRoute(g *echo.Group) {
	g.GET("/topics", self.TopicList)
	g.GET("/topics/last", self.TopicsLast)
	g.GET("/topic/detail", self.Detail)
	g.GET("/topics/node/:nid", self.NodeTopics)
	g.GET("/nodes", self.Nodes)
	g.POST("/topics/new", self.Create)
	g.POST("/topics/modify", self.Modify)
	g.POST("/topics/delete", self.Delete)
	g.POST("/topic/set_top", self.SetTop)
}

func (TopicController) TopicList(ctx echo.Context) error {
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	tab := ctx.QueryParam("tab")
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)
	topics := logic.DefaultTopic.FindAll(context.EchoContext(ctx), paginator, "", tab)
	total := logic.DefaultTopic.Count(context.EchoContext(ctx), tab)
	return success(ctx, map[string]interface{}{
		"list":     topics,
		"total":    total,
		"page":     curPage,
		"per_page": perPage,
	})
}

func (TopicController) TopicsLast(ctx echo.Context) error {
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)
	topics := logic.DefaultTopic.FindAll(context.EchoContext(ctx), paginator, "", "last")
	return success(ctx, map[string]interface{}{
		"list": topics,
		"page": curPage,
	})
}

func (TopicController) Detail(ctx echo.Context) error {
	tid := goutils.MustInt(ctx.QueryParam("tid"))
	topic, replies, err := logic.DefaultTopic.FindByTid(context.EchoContext(ctx), tid)
	if err != nil {
		return fail(ctx, "主题不存在")
	}
	logic.Views.Incr(Request(ctx), model.TypeTopic, tid)
	return success(ctx, map[string]interface{}{
		"topic":   topic,
		"replies": replies,
	})
}

func (TopicController) NodeTopics(ctx echo.Context) error {
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	nid := goutils.MustInt(ctx.Param("nid"))
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)
	topics := logic.DefaultTopic.FindAll(context.EchoContext(ctx), paginator, "", "", nid)
	return success(ctx, map[string]interface{}{
		"list": topics,
		"page": curPage,
	})
}

func (TopicController) Nodes(ctx echo.Context) error {
	nodes := logic.DefaultNode.FindAll(context.EchoContext(ctx))
	return success(ctx, nodes)
}

func (TopicController) Create(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	tid, err := logic.DefaultTopic.Publish(context.EchoContext(ctx), meVal, ctx.Request().Form)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, map[string]interface{}{"tid": tid})
}

func (TopicController) Modify(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	errMsg, err := logic.DefaultTopic.Modify(context.EchoContext(ctx), meVal, ctx.Request().Form)
	if err != nil {
		return fail(ctx, errMsg)
	}
	return success(ctx, nil)
}

func (TopicController) Delete(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	tid := goutils.MustInt(ctx.FormValue("tid"))
	err := logic.DefaultTopic.Delete(context.EchoContext(ctx), tid, meVal.Uid, meVal.IsRoot)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func (TopicController) SetTop(ctx echo.Context) error {
	tid := goutils.MustInt(ctx.FormValue("tid"))
	err := logic.DefaultTopic.SetTop(context.EchoContext(ctx), me(ctx), tid)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func me(ctx echo.Context) *model.Me {
	if meVal, ok := ctx.Get("user").(*model.Me); ok {
		return meVal
	}
	return &model.Me{}
}
