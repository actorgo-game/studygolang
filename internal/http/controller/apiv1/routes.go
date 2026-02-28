package apiv1

import echo "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group) {
	new(IndexController).RegisterRoute(g)
	new(AccountController).RegisterRoute(g)
	new(TopicController).RegisterRoute(g)
	new(ArticleController).RegisterRoute(g)
	new(ResourceController).RegisterRoute(g)
	new(ProjectController).RegisterRoute(g)
	new(BookController).RegisterRoute(g)
	new(WikiController).RegisterRoute(g)
	new(ReadingController).RegisterRoute(g)
	new(UserController).RegisterRoute(g)
	new(CommentController).RegisterRoute(g)
	new(InteractController).RegisterRoute(g)
	new(SidebarController).RegisterRoute(g)
	new(SearchController).RegisterRoute(g)
	new(MessageController).RegisterRoute(g)
	new(MiscController).RegisterRoute(g)
	new(ImageController).RegisterRoute(g)
}
