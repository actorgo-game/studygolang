package apiv1

import (
	"github.com/studygolang/studygolang/context"
	"github.com/studygolang/studygolang/internal/logic"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
	"github.com/polaris1119/times"
)

type SidebarController struct{}

func (self SidebarController) RegisterRoute(g *echo.Group) {
	g.GET("/websites/stat", self.SiteStat)
	g.GET("/topics/recent", self.RecentTopics)
	g.GET("/articles/recent", self.RecentArticles)
	g.GET("/comments/recent", self.RecentComments)
	g.GET("/nodes/hot", self.HotNodes)
	g.GET("/rank/view", self.ViewRank)
	g.GET("/friend/links", self.FriendLinks)
}

func (SidebarController) SiteStat(ctx echo.Context) error {
	commentTotal := logic.DefaultComment.Count(context.EchoContext(ctx), "")
	return success(ctx, map[string]interface{}{
		"user":     logic.DefaultUser.Total(),
		"topic":    logic.DefaultTopic.Total(),
		"article":  logic.DefaultArticle.Total(),
		"comment":  commentTotal,
		"resource": logic.DefaultResource.Total(),
		"project":  logic.DefaultProject.Total(),
		"book":     logic.DefaultGoBook.Total(),
		"wiki":     logic.DefaultWiki.Total(),
	})
}

func (SidebarController) RecentTopics(ctx echo.Context) error {
	topics := logic.DefaultTopic.FindRecent(10)
	return success(ctx, topics)
}

func (SidebarController) RecentArticles(ctx echo.Context) error {
	articles := logic.DefaultArticle.FindBy(context.EchoContext(ctx), 10)
	return success(ctx, articles)
}

func (SidebarController) RecentComments(ctx echo.Context) error {
	comments := logic.DefaultComment.FindRecent(context.EchoContext(ctx), 0, -1, 10)
	return success(ctx, comments)
}

func (SidebarController) HotNodes(ctx echo.Context) error {
	nodes := logic.DefaultTopic.FindHotNodes(context.EchoContext(ctx))
	return success(ctx, nodes)
}

func (SidebarController) ViewRank(ctx echo.Context) error {
	objtype := goutils.MustInt(ctx.QueryParam("objtype"))
	limit := goutils.MustInt(ctx.QueryParam("limit"), 10)
	ymd := times.Format("ymd")
	data := logic.DefaultRank.FindDayRank(context.EchoContext(ctx), objtype, ymd, limit)
	return success(ctx, data)
}

func (SidebarController) FriendLinks(ctx echo.Context) error {
	links := logic.DefaultFriendLink.FindAll(context.EchoContext(ctx))
	return success(ctx, links)
}
